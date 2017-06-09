package operate

import (
	"NfsAgent/comm"
	"NfsAgent/mlog"
	"NfsAgent/util"
	"bytes"
	"encoding/gob"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

func NewLocalFileResourceEx() *LocalfileResourceEx {
	res := LocalfileResourceEx{
		ff_lock:   new(sync.RWMutex),
		FhFileMap: make(map[LocalFileFh]LocalFileEx),
		nf_lock:   new(sync.RWMutex),
		NameFhMap: make(map[string]LocalFileFh),
		mm_lock:   new(sync.Mutex),
		MmunMap:   make(map[string]int),
	}
	return &res
}

func LocalGetFh(b []byte) (Fh, error) {

	var buf bytes.Buffer
	buf.Write(b)

	dnc := gob.NewDecoder(&buf)

	var fh LocalFileFh

	err := dnc.Decode(&fh)

	if err != nil {
		mlog.Error("LocalGetFh god encode fh error :", err)
		return fh, err
	}

	return fh, nil
}

func LocalFhToByte(fh Fh) ([]byte, error) {

	var b []byte

	lfh := fh.(LocalFileFh)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(lfh)
	if err != nil {
		mlog.Error("LocalFhToByte god encode fh error :", err)
		return b, nil
	}

	return buf.Bytes(), nil
}

type LocalFileFh struct {
	Sid int64
	Cli string
	//Token string
}

type LocalFileEx struct {
	Name string
	io.ReadWriter
	io.ReaderAt
	io.WriterAt
}

func (f *LocalFileEx) Stat() (os.FileInfo, error) {
	return os.Lstat(f.Name)
}

func (f LocalFileEx) ReadAt(b []byte, off int64) (n int, err error) {

	file, err := os.OpenFile(f.Name, os.O_RDWR, os.ModeType)

	if err != nil {
		mlog.Error("LocalFileEx open file err:", err)
		return 0, err
	}

	n, err = file.ReadAt(b, off)

	file.Close()

	return n, err
}

func (f LocalFileEx) WriteAt(b []byte, off int64) (n int, err error) {

	file, err := os.OpenFile(f.Name, os.O_RDWR, os.ModePerm)

	if err != nil {
		mlog.Error("LocalFileEx open file err:", err)
		return 0, err
	}

	n, err = file.WriteAt(b, off)

	file.Close()

	return n, err
}

type LocalfileResourceEx struct {
	ff_lock   *sync.RWMutex
	FhFileMap map[LocalFileFh]LocalFileEx

	nf_lock   *sync.RWMutex
	NameFhMap map[string]LocalFileFh

	mm_lock *sync.Mutex
	MmunMap map[string]int
}

func (f *LocalfileResourceEx) FileDump() {
	mlog.Info("timer dump fhmap len :", len(f.FhFileMap), "NameMap len:", len(f.NameFhMap), "Mmunmap len :", len(f.MmunMap))
}

func (c *LocalfileResourceEx) Unmnt(dirpath string) error {

	c.nf_lock.Lock()
	defer c.nf_lock.Unlock()
	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()
	c.mm_lock.Lock()
	defer c.mm_lock.Unlock()

	cnt, ok := c.MmunMap[dirpath]

	if ok {

		c.MmunMap[dirpath] = cnt - 1

		if cnt <= 0 {
			for p, i := range c.NameFhMap {
				if filepath.HasPrefix(p, dirpath) == true {
					delete(c.NameFhMap, p)
					delete(c.FhFileMap, i)
				}
			}
			delete(c.MmunMap, dirpath)
		}

	} else {

		mlog.Warning("LocalfileResourceEx Unmnt no in path :", dirpath)
	}

	return nil
}

func (c *LocalfileResourceEx) Mnt(dirpath string, clientid string) (Fh, error) {

	var fh LocalFileFh
	var file LocalFileEx

	file.Name = dirpath

	f, err := os.Open(dirpath)

	if err != nil {
		mlog.Error("LocalfileResourceEx Mnt OpenDir error:", err)
		return 0, err
	}

	f.Close()

	fh.Sid, err = util.GetIdEx()

	if err != nil {
		mlog.Error("LocalfileResourceEx Mnt GetIdEx error :", err)
		return 0, err
	}

	fh.Cli = clientid

	c.ff_lock.Lock()
	c.FhFileMap[fh] = file
	c.ff_lock.Unlock()

	c.nf_lock.Lock()
	c.NameFhMap[dirpath] = fh
	c.nf_lock.Unlock()

	c.mm_lock.Lock()
	cnt, ok := c.MmunMap[dirpath]

	if ok {
		c.MmunMap[dirpath] = cnt + 1
	} else {
		c.MmunMap[dirpath] = 1
	}

	c.mm_lock.Unlock()

	return fh, nil
}

func (c *LocalfileResourceEx) GetFile(fh Fh) (FileBase, error) {
	c.ff_lock.RLock()
	defer c.ff_lock.RUnlock()

	lfh := fh.(LocalFileFh)

	file, ok := c.FhFileMap[lfh]

	if ok {
		return file, nil
	} else {
		mlog.Error("LocalfileResourceEx GetFile Fh not found")
		return nil, comm.ENOFH
	}
}

func (c *LocalfileResourceEx) Look(fh Fh, name string) (ofh Fh, obj_s *syscall.Stat_t, dir_s *syscall.Stat_t, rerr error) {
	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	lfh := fh.(LocalFileFh)

	file, ok := c.FhFileMap[lfh]

	if ok {

		dir_str := file.Name

		file_path := dir_str + "/" + name

		dirs, err := os.Lstat(dir_str)

		if err != nil {
			mlog.Error("LocalfileResourceEx Look dir stat error :", err)
			return 0, nil, nil, err
		}

		dirst := dirs.Sys().(*syscall.Stat_t)

		fs, err := os.Lstat(file_path)

		if err != nil {
			mlog.Error("LocalfileResourceEx Look obj stat error :", err)
			return 0, nil, dirst, err
		}

		fst := fs.Sys().(*syscall.Stat_t)

		c.nf_lock.Lock()
		defer c.nf_lock.Unlock()

		var nfh LocalFileFh

		ffh, ok := c.NameFhMap[file_path]

		if ok {

			nfh = ffh

		} else {

			nfh.Cli = lfh.Cli
			nfh.Sid, err = util.GetIdEx()

			if err != nil {
				mlog.Error("LocalfileResourceEx GetIdEx error :", err)
				return 0, nil, nil, err
			}

			var nfile LocalFileEx
			nfile.Name = file_path
			//nfile.Info = fs

			c.FhFileMap[nfh] = nfile

			c.NameFhMap[file_path] = nfh
		}

		return nfh, fst, dirst, nil

	} else {
		mlog.Error("LocalfileResourceEx Look Fh not found")
		return 0, nil, nil, comm.ENOFH
	}
}

func (c *LocalfileResourceEx) Stat(fh Fh) (*syscall.Stat_t, error) {

	c.ff_lock.RLock()
	defer c.ff_lock.RUnlock()

	lfh := fh.(LocalFileFh)

	file, ok := c.FhFileMap[lfh]

	if ok {

		mlog.Debug("LocalfileResourceEx Stat file path :", file.Name)

		stat, err := os.Lstat(file.Name)

		if err != nil {
			mlog.Error("LocalfileResourceEx Stat error :", err)
			return nil, err
		}

		s := stat.Sys().(*syscall.Stat_t)

		return s, nil

	} else {
		mlog.Error("LocalfileResourceEx Stat Fh not found")
		return nil, comm.ENOFH
	}
}

func (c *LocalfileResourceEx) ReadDir(fh Fh) ([]DirItemInfo, error) {

	c.ff_lock.RLock()
	defer c.ff_lock.RUnlock()

	lfh := fh.(LocalFileFh)

	file, ok := c.FhFileMap[lfh]

	if ok {

		fileinfos, err := ioutil.ReadDir(file.Name)

		mlog.Debug("LocalfileResourceEx ReadDir dir :", file.Name)

		if err != nil {
			mlog.Error("LocalfileResourceEx readdir error:", err)
			return nil, err
		}

		var infos []DirItemInfo

		for _, i := range fileinfos {

			name := i.Name()

			s := i.Sys().(*syscall.Stat_t)

			var info DirItemInfo
			info.Name = name
			info.Stat = s

			infos = append(infos, info)
		}

		return infos, nil

	} else {
		return nil, comm.ENOFH
	}
}

func (c *LocalfileResourceEx) Remove(fh Fh, name string) (bef *syscall.Stat_t, aft *syscall.Stat_t, err error) {

	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	lfh := fh.(LocalFileFh)

	file, ok := c.FhFileMap[lfh]

	if ok {

		dir := file.Name

		bef_stat, err := os.Lstat(dir)

		if err != nil {
			mlog.Error("LocalfileResourceEx bef stat error :", err)
			return nil, nil, err
		}

		bs := bef_stat.Sys().(*syscall.Stat_t)

		file_path := dir + "/" + name

		tmps, err := os.Lstat(file_path)

		if err != nil {
			mlog.Error("LocalfileResourceEx Remove tmp stat error :", err)
			return nil, nil, err
		}

		if tmps.IsDir() {
			err = os.RemoveAll(file_path)
		} else {
			err = os.Remove(file_path)
		}

		mlog.Debug("LocalfileResourceEx Remove path :", file_path)

		if err != nil {
			mlog.Error("LocalfileResourceEx os Remove error :", err)
			return nil, nil, err
		}

		aft_stat, err := os.Lstat(dir)

		if err != nil {
			mlog.Error("LocalfileResourceEx aft stat error :", err)
			return nil, nil, err
		}

		as := aft_stat.Sys().(*syscall.Stat_t)

		c.nf_lock.Lock()
		defer c.nf_lock.Unlock()

		rfh, ok := c.NameFhMap[file_path]

		if ok {
			delete(c.NameFhMap, file_path)
			delete(c.FhFileMap, rfh)
		} else {
			mlog.Error("LocalfileResourceEx NameFhMap not found file_path :", file_path)
		}

		return bs, as, nil

	} else {
		mlog.Error("LocalfileResource Remove fh not found")
		return nil, nil, comm.ENOFH
	}
}

func (c *LocalfileResourceEx) Rename(ofh Fh, nfh Fh, old, new_name string) (fbef *syscall.Stat_t, faft *syscall.Stat_t, tbef *syscall.Stat_t, taft *syscall.Stat_t, err error) {

	c.ff_lock.RLock()
	defer c.ff_lock.RUnlock()

	lofh := ofh.(LocalFileFh)

	ofile, ok := c.FhFileMap[lofh]

	if !ok {
		mlog.Error("LocalfileResourceEx Rename ofh not found ")
		return nil, nil, nil, nil, comm.ENOFH
	}

	lnfh := nfh.(LocalFileFh)

	nfile, ok := c.FhFileMap[lnfh]

	if !ok {
		mlog.Error("LocalfileResourceEx Rename nfh not found ")
		return nil, nil, nil, nil, comm.ENOFH
	}

	old_path := ofile.Name + "/" + old
	new_path := nfile.Name + "/" + new_name

	obs, err := os.Lstat(ofile.Name)

	if err != nil {
		mlog.Error("LocalfileResourceEx Rename os Lstat obs error :", err)
		return nil, nil, nil, nil, err
	}

	obst := obs.Sys().(*syscall.Stat_t)

	nbs, err := os.Lstat(nfile.Name)

	if err != nil {
		mlog.Error("LocalfileResourceEx Rename os Lstat nbs error :", err)
		return nil, nil, nil, nil, err
	}

	nbst := nbs.Sys().(*syscall.Stat_t)

	err = os.Rename(old_path, new_path)

	if err != nil {
		mlog.Error("LocalfileResourceEx Rename os Rename error :", err)
		return nil, nil, nil, nil, err
	}

	mlog.Debug("LocalfileResourceEx Rename old_path :", old_path, "new_path:", new_path)

	ofs, err := os.Lstat(ofile.Name)

	if err != nil {
		mlog.Error("LocalfileResourceEx Rename os Lstat ofs error :", err)
		return nil, nil, nil, nil, err
	}

	ofst := ofs.Sys().(*syscall.Stat_t)

	nfs, err := os.Lstat(nfile.Name)

	if err != nil {
		mlog.Error("LocalfileResourceEx Rename os Lstat nfs error :", err)
		return nil, nil, nil, nil, err
	}

	nfst := nfs.Sys().(*syscall.Stat_t)
	/*
		c.nf_lock.Lock()
		defer c.nf_lock.Unlock()

		old_file_fh, ok := c.NameFhMap[old_path]

		if ok {

			old_file, ok := c.FhFileMap[old_file_fh]

			if ok {
				old_file.Name = new_path
				c.FhFileMap[old_file_fh] = old_file
			}

			delete(c.NameFhMap, old_path)
		}
	*/
	return obst, nbst, ofst, nfst, nil

}

func (c *LocalfileResourceEx) ReadDirPlus(fh Fh) ([]DirItemInfo, error) {
	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	lfh := fh.(LocalFileFh)

	file, ok := c.FhFileMap[lfh]

	if ok {

		fileinfos, err := ioutil.ReadDir(file.Name)

		mlog.Debug("LocalfileResourceEx ReadDirPlus dir :", file.Name)

		if err != nil {
			mlog.Error("LocalfileResourceEx ReadDirPlus error:", err)
			return nil, err
		}

		c.nf_lock.Lock()
		defer c.nf_lock.Unlock()

		var infos []DirItemInfo

		for _, i := range fileinfos {

			var fh LocalFileFh

			fh.Sid, err = util.GetIdEx()

			if err != nil {
				mlog.Error("LocalfileResourceEx ReadDirPlus error :", err)
				continue
			}

			name := i.Name()

			s := i.Sys().(*syscall.Stat_t)

			path := file.Name + "/" + name

			var nfile LocalFileEx

			nfile.Name = path

			var info DirItemInfo
			info.Name = name
			info.Stat = s
			info.Fh = fh

			c.FhFileMap[fh] = nfile
			c.NameFhMap[path] = fh

			infos = append(infos, info)
		}

		return infos, nil

	} else {
		return nil, comm.ENOFH
	}
}

func (c *LocalfileResourceEx) ReadLink(fh Fh, b []byte) (*syscall.Stat_t, error) {

	c.ff_lock.RLock()
	defer c.ff_lock.RUnlock()

	lfh := fh.(LocalFileFh)

	file, ok := c.FhFileMap[lfh]

	if ok {

		n, err := syscall.Readlink(file.Name, b)

		if err != nil {
			mlog.Error("LocalfileResourceEx ReadLink syscall.Readlink error:", err)
			return nil, err
		}

		mlog.Debug("LocalfileResourceEx ReadLink read n :", n)

		stat, err := os.Lstat(file.Name)

		if err != nil {
			mlog.Error("LocalfileResourceEx ReadLink stat error :", err)
			return nil, err
		}

		s := stat.Sys().(*syscall.Stat_t)

		return s, nil

	} else {

		mlog.Error("LocalfileResourceEx ReadLink fh not found")

		return nil, comm.ENOFH
	}
}

func (c *LocalfileResourceEx) Read(fh Fh, p []byte, count int64, offset int64) (int, *syscall.Stat_t, error) {

	c.ff_lock.RLock()
	defer c.ff_lock.RUnlock()

	lfh := fh.(LocalFileFh)

	file, ok := c.FhFileMap[lfh]

	if ok {

		n, err := file.ReadAt(p, offset)

		if err != nil && err != io.EOF {
			mlog.Error("LocalfileResourceEx read error:", err)
			return n, nil, err
		}

		stat, err := file.Stat()

		if err != nil {
			mlog.Error("LocalfileResourceEx read stat error :", err)
			return 0, nil, err
		}

		fs := stat.Sys().(*syscall.Stat_t)

		if int64(n)+offset > fs.Size {
			mlog.Warning("LocalfileResourceEx read EOF path :", file.Name, "offset:", offset, "n:", n, "size", fs.Size, "count:", count)
		}

		mlog.Debug("LocalfileResourceEx read path:", file.Name, "read n:", n, "len:", count)

		return n, fs, err

	} else {
		return 0, nil, comm.ENOFH
	}
}

func (c *LocalfileResourceEx) Write(fh Fh, p []byte, count int64, offset int64) (n int, bef *syscall.Stat_t, aft *syscall.Stat_t, err error) {

	c.ff_lock.RLock()
	defer c.ff_lock.RUnlock()

	lfh := fh.(LocalFileFh)

	file, ok := c.FhFileMap[lfh]

	if ok {

		bs, err := file.Stat()

		if err != nil {
			mlog.Error("LocalfileResourceEx Write bef stat error :", err)
			return 0, nil, nil, err
		}

		bss := bs.Sys().(*syscall.Stat_t)

		n, err := file.WriteAt(p, offset)

		if err != nil {
			mlog.Error("LocalfileResourceEx Write error:", err)
			return n, nil, nil, err
		}

		mlog.Debug("LocalfileResourceEx Write path :", file.Name, "write n:", n, "len:", len(p), "offset:", offset)

		afs, err := file.Stat()

		if err != nil {
			mlog.Error("LocalfileResourceEx Write afs stat error :", err)
			return 0, nil, nil, err
		}

		fss := afs.Sys().(*syscall.Stat_t)

		return n, bss, fss, err

	} else {
		return 0, nil, nil, comm.ENOFH
	}
}

func (c *LocalfileResourceEx) Create(fh Fh, name string, mode int, file_mode uint32, verf [8]byte, uid int, gid int) (obj_fh Fh, obj_s *syscall.Stat_t, bef *syscall.Stat_t, aft *syscall.Stat_t, err error) {

	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	lfh := fh.(LocalFileFh)

	file, ok := c.FhFileMap[lfh]

	if ok {

		dir := file.Name

		path := dir + "/" + name

		flag := os.O_RDWR | os.O_CREATE | os.O_TRUNC

		if mode != 0 {
			flag = flag | os.O_EXCL
		}

		befs, err := file.Stat()

		if err != nil {
			mlog.Error("LocalfileResourceEx Create f stat error :", err)
			return 0, nil, nil, nil, err
		}

		bfs := befs.Sys().(*syscall.Stat_t)

		//fm := uint32(syscall.S_IRUSR | syscall.S_IWUSR | syscall.S_IXUSR | syscall.S_IRGRP | syscall.S_IWG | syscall.S_IXGRP |
		//syscall.S_IROTH | syscall.S_IWOTH | syscall.S_IXOTH)
		fm := uint32(0777)

		if file_mode != 0 {
			fm = file_mode
		}

		oldMask := syscall.Umask(0)

		nf, err := os.OpenFile(path, flag, os.FileMode(fm))

		syscall.Umask(oldMask)

		if err != nil {
			mlog.Error("LocalfileResourceEx Create err :", err)
			return 0, nil, nil, nil, err
		}

		if mode == 2 {
			var ubuf syscall.Utimbuf

			ubuf.Actime = int64(verf[0] | verf[1]<<8 | verf[2]<<16 | verf[3]<<24)
			ubuf.Modtime = int64(verf[4] | verf[5]<<8 | verf[6]<<16 | verf[7]<<24)

			err := syscall.Utime(path, &ubuf)

			if err != nil {
				mlog.Error("LocalfileResourceEx Create syscall.Utime error :", err)
				return 0, nil, nil, nil, err
			}
		}
		if uid != 0 || gid != 0 {
			err = os.Chown(path, uid, gid)

			if err != nil {
				mlog.Error("LocalfileResourceEx Create os.Chown error:", err)
				return 0, nil, nil, nil, err
			}
		}

		aft, err := file.Stat()

		if err != nil {
			mlog.Error("LocalfileResourceEx Create f stat error :", err)
			return 0, nil, nil, nil, err
		}

		afts := aft.Sys().(*syscall.Stat_t)

		ns, err := nf.Stat()

		if err != nil {
			mlog.Error("LocalfileResourceEx nf.Stat error :", err)
			return 0, nil, nil, nil, err
		}

		nst := ns.Sys().(*syscall.Stat_t)

		var nfh LocalFileFh

		nfh.Cli = lfh.Cli
		nfh.Sid, err = util.GetIdEx()

		if err != nil {
			mlog.Error("LocalfileResourceEx GetIdEx error :", err)
			return 0, nil, nil, nil, err
		}

		var lnf LocalFileEx
		lnf.Name = path

		c.FhFileMap[nfh] = lnf

		c.nf_lock.Lock()
		defer c.nf_lock.Unlock()

		c.NameFhMap[path] = nfh

		return nfh, nst, bfs, afts, nil

	} else {
		return 0, nil, nil, nil, comm.ENOFH
	}
}

func (c *LocalfileResourceEx) Mkdir(fh Fh, name string, attr Setttr, uid int, gid int) (nfh Fh, objs *syscall.Stat_t, bef *syscall.Stat_t, aft *syscall.Stat_t, err error) {
	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	lfh := fh.(LocalFileFh)

	file, ok := c.FhFileMap[lfh]

	if ok {

		dir := file.Name

		path := dir + "/" + name

		bstat, err := file.Stat()

		if err != nil {
			mlog.Error("LocalfileResourceEx Mkdir bef stat error :", err)
			return 0, nil, nil, nil, err
		}

		bs := bstat.Sys().(*syscall.Stat_t)

		mlog.Debug("LocalfileResourceEx Mkdir mode :", attr.Mode)

		oldMask := syscall.Umask(0)

		err = syscall.Mkdir(path, attr.Mode)
		//err = os.Mkdir(path, os.FileMode(attr.Mode))

		syscall.Umask(oldMask)

		if err != nil {
			mlog.Error("LocalfileResourceEx os Mkdir error :", err)
			return 0, nil, bs, bs, err
		}

		if attr.Uid != 0 || attr.Gid != 0 {
			mlog.Debug("LocalfileResourceEx syscall.Chown path :", path, "uid:", attr.Uid, "gid:", attr.Gid)
			err = syscall.Chown(path, int(attr.Uid), int(attr.Gid))
			if err != nil {
				mlog.Error("LocalfileResourceEx syscall.Chown error :", err)
				return 0, nil, nil, nil, err
			}
		}

		if uid != 0 || gid != 0 {
			mlog.Debug("LocalfileResourceEx x syscall.Chown path :", path, "uid:", uid, "gid:", gid)
			err = syscall.Chown(path, uid, gid)
			if err != nil {
				mlog.Error("LocalfileResourceEx x syscall.Chown error :", err)
				return 0, nil, nil, nil, err
			}
		}

		astat, err := file.Stat()

		if err != nil {
			mlog.Error("LocalfileResource Mkdir aft stat error :", err)
			return 0, nil, nil, nil, err
		}

		as := astat.Sys().(*syscall.Stat_t)

		var nf LocalFileEx
		nf.Name = path

		nstat, err := nf.Stat()

		if err != nil {
			mlog.Error("LocalfileResourceEx nstat error :", err)
			return 0, nil, nil, nil, err
		}

		ns := nstat.Sys().(*syscall.Stat_t)

		var nfh LocalFileFh
		nfh.Cli = lfh.Cli
		nfh.Sid, err = util.GetIdEx()

		if err != nil {
			mlog.Error("LocalfileResourceEx GetIdEx error :", err)
			return 0, nil, nil, nil, err
		}

		c.FhFileMap[nfh] = nf

		c.nf_lock.Lock()
		defer c.nf_lock.Unlock()

		c.NameFhMap[path] = nfh

		return nfh, ns, bs, as, nil

	} else {
		mlog.Error("LocalfileResourceEx Mkdir fh not found")

		return 0, nil, nil, nil, comm.ENOFH
	}
}

func (c *LocalfileResourceEx) Symlink(fh Fh, name string, path string, uid int, gid int) (nfh Fh, objs *syscall.Stat_t, bef *syscall.Stat_t, aft *syscall.Stat_t, err error) {
	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	lfh := fh.(LocalFileFh)

	file, ok := c.FhFileMap[lfh]

	if ok {

		dir := file.Name
		file_path := dir + "/" + name

		bstat, err := os.Lstat(dir)

		if err != nil {
			mlog.Error("LocalfileResourceEx Symlink bef stat error :", err)
			return 0, nil, nil, nil, err
		}

		bs := bstat.Sys().(*syscall.Stat_t)

		mlog.Debug("LocalfileResourceEx Symlink path :", path, "file_path:", file_path)

		err = syscall.Symlink(path, file_path)

		if err != nil {
			mlog.Error("LocalfileResourceEx Symlink syscall.Symlink error :", err)
			return 0, nil, nil, nil, err
		}

		if uid != 0 || gid != 0 {
			mlog.Debug("LocalfileResourceEx Symlink syscall.Chown path :", file_path, "uid:", uid, "gid:", gid)
			err = syscall.Lchown(file_path, uid, gid)
			if err != nil {
				mlog.Error("LocalfileResourceEx Symlink syscall.Chown error :", err)
				return 0, nil, nil, nil, err
			}
		}

		astat, err := os.Lstat(dir)

		if err != nil {
			mlog.Error("LocalfileResourceEx Symlink aft stat error :", err)
			return 0, nil, nil, nil, err
		}

		as := astat.Sys().(*syscall.Stat_t)

		nstat, err := os.Lstat(file_path)

		if err != nil {
			mlog.Error("LocalfileResourceEx Symlink nstat error :", err)
			return 0, nil, nil, nil, err
		}

		ns := nstat.Sys().(*syscall.Stat_t)

		var nfh LocalFileFh

		nfh.Cli = lfh.Cli
		nfh.Sid, err = util.GetIdEx()

		if err != nil {
			mlog.Error("LocalfileResourceEx Symlink GetIdEx error :", err)
			return 0, nil, nil, nil, err
		}

		var lnf LocalFileEx
		lnf.Name = file_path

		c.FhFileMap[nfh] = lnf

		c.nf_lock.Lock()
		defer c.nf_lock.Unlock()

		c.NameFhMap[path] = nfh

		return nfh, ns, bs, as, nil

	} else {
		mlog.Error("LocalfileResourceEx fh not found")
		return 0, nil, nil, nil, comm.ENOFH
	}
}

func (c *LocalfileResourceEx) Link(fh Fh, name string, path string, uid int, gid int) (nfh Fh, objs *syscall.Stat_t, bef *syscall.Stat_t, aft *syscall.Stat_t, err error) {
	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	lfh := fh.(LocalFileFh)

	file, ok := c.FhFileMap[lfh]

	if ok {

		dir := file.Name
		file_path := dir + "/" + name

		bstat, err := os.Lstat(dir)

		if err != nil {
			mlog.Error("LocalfileResourceEx Link bef stat error :", err)
			return 0, nil, nil, nil, err
		}

		bs := bstat.Sys().(*syscall.Stat_t)

		mlog.Debug("LocalfileResourceEx Link path :", path, "file_path:", file_path)

		err = syscall.Link(path, file_path)

		if err != nil {
			mlog.Error("LocalfileResourceEx Link syscall.Symlink error :", err)
			return 0, nil, nil, nil, err
		}

		if uid != 0 || gid != 0 {
			mlog.Debug("LocalfileResourceEx Link syscall.Chown path :", file_path, "uid:", uid, "gid:", gid)
			err = syscall.Chown(file_path, uid, gid)
			if err != nil {
				mlog.Error("LocalfileResourceEx Link syscall.Chown error :", err)
				return 0, nil, nil, nil, err
			}
		}

		astat, err := os.Lstat(dir)

		if err != nil {
			mlog.Error("LocalfileResourceEx Link aft stat error :", err)
			return 0, nil, nil, nil, err
		}

		as := astat.Sys().(*syscall.Stat_t)

		nstat, err := os.Lstat(file_path)

		if err != nil {
			mlog.Error("LocalfileResourceEx Link nstat error :", err)
			return 0, nil, nil, nil, err
		}

		ns := nstat.Sys().(*syscall.Stat_t)

		var nfh LocalFileFh

		nfh.Cli = lfh.Cli
		nfh.Sid, err = util.GetIdEx()

		if err != nil {
			mlog.Error("LocalfileResourceEx Link GetIdEx error :", err)
			return 0, nil, nil, nil, err
		}

		var lnf LocalFileEx
		lnf.Name = file_path

		c.FhFileMap[nfh] = lnf

		c.nf_lock.Lock()
		defer c.nf_lock.Unlock()

		c.NameFhMap[path] = nfh

		return nfh, ns, bs, as, nil

	} else {
		mlog.Error("LocalfileResourceEx fh not found")
		return 0, nil, nil, nil, comm.ENOFH
	}
}
