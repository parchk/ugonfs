package operate

import (
	"NfsAgent/mlog"
	"NfsAgent/util"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"syscall"
)

func NewLocalFileResource() *LocalfileResource {
	res := LocalfileResource{
		fc_lock:     new(sync.RWMutex),
		FhCliMap:    make(map[ClientInfo]Fh),
		ff_lock:     new(sync.RWMutex),
		FhFileMap:   make(map[Fh]LocalFile),
		nf_lock:     new(sync.RWMutex),
		NameFileMap: make(map[string]Fh),
	}

	return &res
}

type ClientInfo struct {
	Dirpath  string
	Clientid string
}

type LocalFile struct {
	F    *os.File
	Name string
	io.ReadWriter
	io.ReaderAt
	io.WriterAt
}

type LocalfileResource struct {
	fc_lock  *sync.RWMutex
	FhCliMap map[ClientInfo]Fh

	ff_lock   *sync.RWMutex
	FhFileMap map[Fh]LocalFile

	nf_lock     *sync.RWMutex
	NameFileMap map[string]Fh
}

func (c *LocalfileResource) Mnt(dirpath string, clientid string) (Fh, error) {

	file, err := os.Open(dirpath)

	if err != nil {
		mlog.Error("LocalfileResource Mnt OpenDir error:", err)
		return 0, err
	}

	fh := Fh(file.Fd())

	c.fc_lock.Lock()
	defer c.fc_lock.Unlock()

	var cli_tmp ClientInfo
	cli_tmp.Dirpath = dirpath
	cli_tmp.Clientid = clientid

	nfh, ok := c.FhCliMap[cli_tmp]

	if ok {
		return nfh, nil
	} else {
		c.FhCliMap[cli_tmp] = fh

		c.ff_lock.Lock()
		defer c.ff_lock.Unlock()
		var lfile LocalFile
		lfile.F = file
		lfile.Name = dirpath
		c.FhFileMap[fh] = lfile
		return fh, nil
	}
}

func (c *LocalfileResource) GetFile(fh Fh) (FileBase, error) {
	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	file, ok := c.FhFileMap[fh]

	if ok {
		return file, nil
	} else {
		mlog.Error("LocalfileResource Stat Fh not found")
		return nil, errors.New("fh not found")
	}
}

func (c *LocalfileResource) Look(fh Fh, name string) (ofh Fh, obj_s *syscall.Stat_t, dir_s *syscall.Stat_t, rerr error) {
	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	file, ok := c.FhFileMap[fh]

	if ok {

		//f := file.(*os.File)
		//f := file.F

		dir_str := file.Name

		file_path := dir_str + "/" + name
		/*
			fs, err := f.Stat()

			if err != nil {
				mlog.Error("LocalfileResource Look dir stat error :", err)
				return 0, nil, nil, err
			}
		*/

		fs, err := os.Lstat(dir_str)

		if err != nil {
			mlog.Error("LocalfileResource Look dir stat error :", err)
			return 0, nil, nil, err
		}

		fss := fs.Sys().(*syscall.Stat_t)

		mlog.Debug("LocalfileResource Look  fss ")

		ffh, ok := c.NameFileMap[file_path]

		if ok {

			mlog.Debug("LocalfileResource Look exist file_path:", file_path)

			ff, ok := c.FhFileMap[ffh]

			if ok {

				mlog.Debug("LocalfileResource Look exist file:", ffh)

				//tf := ff.(*os.File)
				//tf := ff.F
				/*
					tfstat, err := tf.Stat()

					if err != nil {
						mlog.Error("LocalfileResource Look Stat error :", err)
						return 0, nil, nil, err
					}
				*/

				mlog.Debug("LocalfileResource Look  ff name :", ff.Name)

				tfstat, err := os.Lstat(ff.Name)

				if err != nil {
					mlog.Error("LocalfileResource Look Stat error :", err)
					return 0, nil, nil, err
				}

				tfs := tfstat.Sys().(*syscall.Stat_t)

				return ffh, tfs, fss, err

			} else {
				mlog.Error("LocalfileResource Look Stat Fh not found")
				return 0, nil, nil, errors.New("fh not found")
			}

		} else {

			mlog.Debug("LocalfileResource Look create file_path:", file_path)

			fileinfo, err := os.Lstat(file_path)

			if err != nil {
				mlog.Error("LocalfileResource Look os stat error:", err, "file_path:", file_path)
				return 0, nil, nil, err
			}

			var file *os.File
			var lfile LocalFile
			var nfd Fh
			lfile.Name = file_path

			if fileinfo.IsDir() {
				file, err = os.Open(file_path)
				if err != nil {
					mlog.Error("LocalfileResource Look open file err:", err)
					return 0, nil, nil, err
				}
				nfd = Fh(file.Fd())

			} else if fileinfo.Mode()&os.ModeSymlink == 0 {
				file, err = os.OpenFile(file_path, os.O_RDWR, os.ModeType)
				if err != nil {
					mlog.Error("LocalfileResource Look open file err:", err)
					return 0, nil, nil, err
				}
				nfd = Fh(file.Fd())

			} else {

				id, err := util.GetId()

				if err != nil {
					mlog.Error("LocalfileResource Look uitl GetId error:", err)
					return 0, nil, nil, err
				}

				nfd = Fh(id)
			}

			c.NameFileMap[file_path] = nfd

			lfile.F = file
			c.FhFileMap[nfd] = lfile
			/*
				stat, err := file.Stat()

				if err != nil {
					mlog.Error("LocalfileResource stat err :", err)
					return 0, nil, nil, err
				}
			*/

			stat, err := os.Lstat(file_path)

			if err != nil {
				mlog.Error("LocalfileResource stat err :", err)
				return 0, nil, nil, err
			}

			s := stat.Sys().(*syscall.Stat_t)

			return nfd, s, fss, nil
		}

	} else {
		mlog.Error("LocalfileResource Stat Fh not found")
		return 0, nil, nil, errors.New("fh not found")
	}
}

func (c *LocalfileResource) Stat(fh Fh) (*syscall.Stat_t, error) {

	c.ff_lock.RLock()
	defer c.ff_lock.RUnlock()

	file, ok := c.FhFileMap[fh]

	if ok {

		//f := file.(*os.File)

		var path string

		path = file.Name

		stat, err := os.Lstat(path)

		if err != nil {
			mlog.Error("LocalfileResource Stat error :", err)
			return nil, err
		}

		s := stat.Sys().(*syscall.Stat_t)

		return s, nil

	} else {
		mlog.Error("LocalfileResource Stat Fh not found")
		return nil, errors.New("fh not found")
	}
}

func (c *LocalfileResource) ReadDir(fh Fh) ([]DirItemInfo, error) {
	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	file, ok := c.FhFileMap[fh]
	if ok {

		//f := file.(*os.File)

		f := file.F

		dir := f.Name()
		fileinfos, err := ioutil.ReadDir(dir)

		mlog.Debug("LocalfileResource ReadDir dir :", dir)

		if err != nil {
			mlog.Error("LocalfileResource readdir error:", err)
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
		return nil, errors.New("fh not found")
	}
}

func (c *LocalfileResource) Remove(fh Fh, name string) (bef *syscall.Stat_t, aft *syscall.Stat_t, err error) {
	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	file, ok := c.FhFileMap[fh]

	if ok {
		//f := file.(*os.File)

		f := file.F

		dir := f.Name()

		bef_stat, err := f.Stat()

		if err != nil {
			mlog.Error("LocalfileResource bef stat error :", err)
			return nil, nil, err
		}

		bs := bef_stat.Sys().(*syscall.Stat_t)

		file_path := dir + "/" + name

		err = os.Remove(file_path)

		if err != nil {
			mlog.Error("LocalfileResource os Remove error :", err)
			return nil, nil, err
		}

		aft_stat, err := f.Stat()

		if err != nil {
			mlog.Error("LocalfileResource aft stat error :", err)
			return nil, nil, err
		}

		as := aft_stat.Sys().(*syscall.Stat_t)

		c.nf_lock.Lock()
		defer c.nf_lock.Unlock()

		fh, ok := c.NameFileMap[file_path]

		if ok {
			delete(c.FhFileMap, fh)
			delete(c.NameFileMap, file_path)
		} else {
			mlog.Warning("LocalfileResource remove NameFileMap not found")
		}

		return bs, as, nil

	} else {
		mlog.Error("LocalfileResource Remove fh not found")
		return nil, nil, errors.New("fh not found")
	}

}

func (c *LocalfileResource) Rename(ofh Fh, nfh Fh, old, new_name string) (fbef *syscall.Stat_t, faft *syscall.Stat_t, tbef *syscall.Stat_t, taft *syscall.Stat_t, err error) {

	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	ofile, ok := c.FhFileMap[ofh]

	if !ok {
		mlog.Error("LocalfileResource Rename ofh not found ")
		return nil, nil, nil, nil, errors.New("fh not found")
	}

	nfile, ok := c.FhFileMap[nfh]

	if !ok {
		mlog.Error("LocalfileResource Rename nfh not found ")
		return nil, nil, nil, nil, errors.New("fh not found")
	}
	/*
		of := ofile.(*os.File)
		nf := nfile.(*os.File)
	*/
	of := ofile.F
	nf := nfile.F

	old_path := of.Name() + "/" + old
	new_path := nf.Name() + "/" + new_name

	obs, err := os.Lstat(of.Name())

	if err != nil {
		mlog.Error("LocalfileResource Rename os Lstat obs error :", err)
		return nil, nil, nil, nil, err
	}

	obst := obs.Sys().(*syscall.Stat_t)

	nbs, err := os.Lstat(nf.Name())

	if err != nil {
		mlog.Error("LocalfileResource Rename os Lstat nbs error :", err)
		return nil, nil, nil, nil, err
	}

	nbst := nbs.Sys().(*syscall.Stat_t)

	err = os.Rename(old_path, new_path)

	if err != nil {
		mlog.Error("LocalfileResource Rename os Rename error :", err)
		return nil, nil, nil, nil, err
	}

	ofs, err := os.Lstat(of.Name())

	if err != nil {
		mlog.Error("LocalfileResource Rename os Lstat ofs error :", err)
		return nil, nil, nil, nil, err
	}

	ofst := ofs.Sys().(*syscall.Stat_t)

	nfs, err := os.Lstat(nf.Name())

	if err != nil {
		mlog.Error("LocalfileResource Rename os Lstat nfs error :", err)
		return nil, nil, nil, nil, err
	}

	nfst := nfs.Sys().(*syscall.Stat_t)

	return obst, nbst, ofst, nfst, nil

}

func (c *LocalfileResource) ReadLink(fh Fh, b []byte) (*syscall.Stat_t, error) {
	c.ff_lock.RLock()
	defer c.ff_lock.RUnlock()

	file, ok := c.FhFileMap[fh]

	if ok {

		//f := file.(*os.File)

		f := file.F

		var path string

		if f == nil {
			path = file.Name
		} else {
			path = f.Name()
		}

		n, err := syscall.Readlink(path, b)

		if err != nil {
			mlog.Error("LocalfileResource ReadLink syscall.Readlink error:", err)
			return nil, err
		}

		mlog.Debug("LocalfileResource ReadLink read n :", n)

		stat, err := os.Lstat(path)

		if err != nil {
			mlog.Error("LocalfileResource ReadLink stat error :", err)
			return nil, err
		}

		s := stat.Sys().(*syscall.Stat_t)
		return s, nil

	} else {
		mlog.Error("LocalfileResource ReadLink fh not found")
		return nil, errors.New("fh not found")
	}
}

func (c *LocalfileResource) Read(fh Fh, p []byte, count int64, offset int64) (int, *syscall.Stat_t, error) {

	c.ff_lock.RLock()
	defer c.ff_lock.RUnlock()

	file, ok := c.FhFileMap[fh]
	if ok {

		//f := file.(*os.File)

		f := file.F

		stat, err := f.Stat()

		if err != nil {
			mlog.Error("LocalfileResource read stat error :", err)
			return 0, nil, err
		}

		fs := stat.Sys().(*syscall.Stat_t)

		n, err := f.ReadAt(p, offset)

		if err != nil && err != io.EOF {
			mlog.Error("LocalfileResource read error:", err)
			return n, nil, err
		}
		return n, fs, err

	} else {
		return 0, nil, errors.New("fh not found")
	}
}

func (c *LocalfileResource) Write(fh Fh, p []byte, count int64, offset int64) (n int, bef *syscall.Stat_t, aft *syscall.Stat_t, err error) {
	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	file, ok := c.FhFileMap[fh]
	if ok {
		//f := file.(*os.File)
		f := file.F

		bs, err := f.Stat()

		if err != nil {
			mlog.Error("LocalfileResource Write bef stat error :", err)
			return 0, nil, nil, nil
		}

		bss := bs.Sys().(*syscall.Stat_t)

		n, err := f.WriteAt(p, offset)
		if err != nil {
			mlog.Error("LocalfileResource Write error:", err)
			return n, nil, nil, err
		}

		afs, err := f.Stat()

		if err != nil {
			mlog.Error("LocalfileResource Write afs stat error :", err)
			return 0, nil, nil, nil
		}

		fss := afs.Sys().(*syscall.Stat_t)

		return n, bss, fss, err

	} else {
		return 0, nil, nil, errors.New("fh not found")
	}
}

func (c *LocalfileResource) create(fh Fh, name string) (obj_fh Fh, obj_s *syscall.Stat_t, bef *syscall.Stat_t, aft *syscall.Stat_t, err error) {
	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	file, ok := c.FhFileMap[fh]
	if ok {
		//f := file.(*os.File)
		f := file.F

		befs, err := f.Stat()

		if err != nil {
			mlog.Error("LocalfileResource Create os bef stat error :", err)
			return 0, nil, nil, nil, err
		}

		bef = befs.Sys().(*syscall.Stat_t)

		dir_str := f.Name()
		file_path := dir_str + "/" + name
		nfile, err := os.Create(file_path)
		if err != nil {
			mlog.Error("LocalfileResource Create os create error :", err)
			return 0, nil, nil, nil, err
		}

		stat, err := nfile.Stat()
		if err != nil {
			mlog.Error("LocalfileResource Create os stat error :", err)
			return 0, nil, nil, nil, err
		}

		obj_s = stat.Sys().(*syscall.Stat_t)

		afts, err := f.Stat()

		if err != nil {
			mlog.Error("LocalfileResource Create os aft stat error :", err)
			return 0, nil, nil, nil, err
		}

		aft = afts.Sys().(*syscall.Stat_t)

		obj_fh = Fh(nfile.Fd())

		var lnfile LocalFile
		lnfile.F = nfile
		lnfile.Name = file_path

		c.FhFileMap[obj_fh] = lnfile

		c.nf_lock.Lock()
		defer c.nf_lock.Unlock()

		c.NameFileMap[file_path] = obj_fh

		return obj_fh, obj_s, bef, aft, nil

	} else {
		return 0, nil, nil, nil, errors.New("fh no found")
	}

}

func (c *LocalfileResource) Create(fh Fh, name string, mode int, file_mode uint32, verf [8]byte) (obj_fh Fh, obj_s *syscall.Stat_t, bef *syscall.Stat_t, aft *syscall.Stat_t, err error) {

	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	file, ok := c.FhFileMap[fh]

	if ok {
		//f := file.(*os.File)

		f := file.F

		dir := f.Name()

		path := dir + "/" + name

		flag := os.O_RDWR | os.O_CREATE | os.O_TRUNC

		if mode != 0 {
			flag = flag | os.O_EXCL
		}

		befs, err := f.Stat()

		if err != nil {
			mlog.Error("LocalfileResource Create f stat error :", err)
			return 0, nil, nil, nil, err
		}

		bfs := befs.Sys().(*syscall.Stat_t)

		fm := uint32(syscall.S_IRUSR | syscall.S_IWUSR | syscall.S_IRGRP |
			syscall.S_IROTH)

		if file_mode != 0 {
			fm = file_mode
		}

		nf, err := os.OpenFile(path, flag, os.FileMode(fm))

		if err != nil {
			mlog.Error("LocalfileResource Create err :", err)
			return 0, nil, nil, nil, err
		}

		if mode == 2 {
			var ubuf syscall.Utimbuf

			ubuf.Actime = int64(verf[0] | verf[1]<<8 | verf[2]<<16 | verf[3]<<24)
			ubuf.Modtime = int64(verf[4] | verf[5]<<8 | verf[6]<<16 | verf[7]<<24)

			err := syscall.Utime(path, &ubuf)

			if err != nil {
				mlog.Error("LocalfileResource Create syscall.Utime error :", err)
				return 0, nil, nil, nil, err
			}
		}

		aft, err := f.Stat()

		if err != nil {
			mlog.Error("LocalfileResource Create f stat error :", err)
			return 0, nil, nil, nil, err
		}

		afts := aft.Sys().(*syscall.Stat_t)

		ns, err := nf.Stat()

		if err != nil {
			mlog.Error("LocalfileResource nf.Stat error :", err)
			return 0, nil, nil, nil, err
		}

		nst := ns.Sys().(*syscall.Stat_t)

		nfh := Fh(nf.Fd())

		var lnf LocalFile

		lnf.F = nf
		lnf.Name = path

		c.FhFileMap[nfh] = lnf
		c.nf_lock.Lock()
		defer c.nf_lock.Unlock()
		c.NameFileMap[path] = nfh

		return nfh, nst, bfs, afts, nil

	} else {
		return 0, nil, nil, nil, errors.New("fh not found")
	}

}

func (c *LocalfileResource) Mkdir(fh Fh, name string, mode uint32) (nfh Fh, objs *syscall.Stat_t, bef *syscall.Stat_t, aft *syscall.Stat_t, err error) {
	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	file, ok := c.FhFileMap[fh]

	if ok {
		//f := file.(*os.File)
		f := file.F

		dir := f.Name()

		path := dir + "/" + name

		bstat, err := f.Stat()

		if err != nil {
			mlog.Error("LocalfileResource Mkdir bef stat error :", err)
			return 0, nil, nil, nil, err
		}

		bs := bstat.Sys().(*syscall.Stat_t)

		err = syscall.Mkdir(path, mode)

		//err = os.Mkdir(path, os.FileMode(mode))

		if err != nil {
			mlog.Error("LocalfileResource os Mkdir error :", err)
			return 0, nil, nil, nil, err
		}

		astat, err := f.Stat()

		if err != nil {
			mlog.Error("LocalfileResource Mkdir aft stat error :", err)
			return 0, nil, nil, nil, err
		}

		as := astat.Sys().(*syscall.Stat_t)

		nf, err := os.Open(path)

		if err != nil {
			mlog.Error("LocalfileResource os open error :", err)
			return 0, nil, nil, nil, err
		}

		nstat, err := nf.Stat()

		if err != nil {
			mlog.Error("LocalfileResource nstat error :", err)
			return 0, nil, nil, nil, err
		}

		ns := nstat.Sys().(*syscall.Stat_t)

		nfh := Fh(nf.Fd())

		var lnf LocalFile
		lnf.F = nf
		lnf.Name = path

		c.FhFileMap[nfh] = lnf
		c.NameFileMap[path] = nfh

		return nfh, ns, bs, as, nil

	} else {
		mlog.Error("LocalfileResource Mkdir fh not found")
		return 0, nil, nil, nil, errors.New("fh not found")
	}
}

func (c *LocalfileResource) Symlink(fh Fh, name string, path string) (nfh Fh, objs *syscall.Stat_t, bef *syscall.Stat_t, aft *syscall.Stat_t, err error) {
	c.ff_lock.Lock()
	defer c.ff_lock.Unlock()

	file, ok := c.FhFileMap[fh]

	if ok {
		//f := file.(*os.File)

		//f := file.F

		//dir := f.Name()
		dir := file.Name
		file_path := dir + "/" + name

		bstat, err := os.Lstat(dir)

		if err != nil {
			mlog.Error("LocalfileResource Symlink bef stat error :", err)
			return 0, nil, nil, nil, err
		}

		bs := bstat.Sys().(*syscall.Stat_t)

		mlog.Debug("LocalfileResource Symlink path :", path, "file_path:", file_path)

		err = syscall.Symlink(path, file_path)

		if err != nil {
			mlog.Error("LocalfileResource Symlink syscall.Symlink error :", err)
			return 0, nil, nil, nil, err
		}

		astat, err := os.Lstat(dir)

		if err != nil {
			mlog.Error("LocalfileResource Symlink aft stat error :", err)
			return 0, nil, nil, nil, err
		}

		as := astat.Sys().(*syscall.Stat_t)

		nf, err := os.Open(path)

		if err != nil {
			mlog.Error("LocalfileResource Symlink os open error :", err)
			return 0, nil, nil, nil, err
		}

		/*
			nstat, err := nf.Stat()

			if err != nil {
				mlog.Error("LocalfileResource Symlink nstat error :", err)
				return 0, nil, nil, nil, err
			}
		*/

		nstat, err := os.Lstat(file_path)

		if err != nil {
			mlog.Error("LocalfileResource Symlink nstat error :", err)
			return 0, nil, nil, nil, err
		}

		ns := nstat.Sys().(*syscall.Stat_t)

		nfh := Fh(nf.Fd())

		var lnf LocalFile
		lnf.F = nf
		lnf.Name = file_path

		c.FhFileMap[nfh] = lnf
		c.NameFileMap[path] = nfh

		return nfh, ns, bs, as, nil

	} else {
		mlog.Error("LocalfileResource fh not found")
		return 0, nil, nil, nil, errors.New("fh not found")
	}
}
