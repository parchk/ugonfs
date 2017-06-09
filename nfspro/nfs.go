package nfspro

import (
	"NfsAgent/mlog"
	"NfsAgent/operate"
	"bytes"
	"time"
	//"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	//"os"
	"strings"
	"syscall"

	"github.com/prashanthpai/sunrpc"
	"github.com/rasky/go-xdr/xdr2"
)

func GetPreAttr(stat *syscall.Stat_t) pre_op_attr {
	var attr pre_op_attr
	attr.Attributes_follow = true
	attr.Attributes.Ctime.Nseconds = uint32(stat.Ctim.Nsec)
	attr.Attributes.Ctime.Seconds = uint32(stat.Ctim.Sec)
	attr.Attributes.Mtime.Nseconds = uint32(stat.Mtim.Nsec)
	attr.Attributes.Mtime.Seconds = uint32(stat.Mtim.Sec)
	attr.Attributes.Size = size3(stat.Size)

	return attr
}

func GetPosAttr(stat *syscall.Stat_t) post_op_attr {
	var attr post_op_attr
	attr.Attributes_follow = true
	attr.Attributes.Atime.Nseconds = uint32(stat.Atim.Nsec)
	attr.Attributes.Atime.Seconds = uint32(stat.Atim.Sec)
	attr.Attributes.Ctime.Nseconds = uint32(stat.Ctim.Nsec)
	attr.Attributes.Ctime.Seconds = uint32(stat.Ctim.Sec)
	attr.Attributes.Fileid = fileid3(stat.Ino)
	attr.Attributes.Fsid = stat.Dev
	attr.Attributes.Gid = gid3(stat.Gid)
	attr.Attributes.Mode = mode3(stat.Mode)
	attr.Attributes.Mtime.Nseconds = uint32(stat.Mtim.Nsec)
	attr.Attributes.Mtime.Seconds = uint32(stat.Mtim.Sec)
	attr.Attributes.Nlink = uint32(stat.Nlink)
	attr.Attributes.Rdev.Specdata1 = uint32((stat.Rdev >> 8) & 0xFF)
	attr.Attributes.Rdev.Specdata2 = uint32(stat.Rdev & 0xFF)
	attr.Attributes.Size = size3(stat.Size)

	if (stat.Mode & syscall.S_IFMT) == syscall.S_IFDIR {
		attr.Attributes.Type = NF3DIR
	}
	if (stat.Mode & syscall.S_IFMT) == syscall.S_IFSOCK {
		attr.Attributes.Type = NF3SOCK
	}
	if (stat.Mode & syscall.S_IFMT) == syscall.S_IFLNK {
		attr.Attributes.Type = NF3LNK
	}
	if (stat.Mode & syscall.S_IFMT) == syscall.S_IFIFO {
		attr.Attributes.Type = NF3FIFO
	}
	if (stat.Mode & syscall.S_IFMT) == syscall.S_IFREG {
		attr.Attributes.Type = NF3REG
	}

	attr.Attributes.Uid = uid3(stat.Uid)
	attr.Attributes.Used = size3(stat.Blksize * 512)

	return attr
}

func InsertEntryPlus3(h, d *entryplus3, p int) bool {
	if h.Nextentry == nil {
		h.Nextentry = d
		return true
	}
	i := 0
	n := h
	for n.Nextentry != nil {
		i++
		if i == p {
			if n.Nextentry.Nextentry == nil {
				n.Nextentry = d
				return true
			} else {
				d.Nextentry = n.Nextentry
				n.Nextentry = d.Nextentry
				return true
			}
		}
		n = n.Nextentry
		if n.Nextentry == nil {
			n.Nextentry = d
			return true
		}
	}
	return false
}

func InsertEntry3(h, d *entry3, p int) bool {
	if h.Nextentry == nil {
		h.Nextentry = d
		return true
	}
	i := 0
	n := h
	for n.Nextentry != nil {
		i++
		if i == p {
			if n.Nextentry.Nextentry == nil {
				n.Nextentry = d
				return true
			} else {
				d.Nextentry = n.Nextentry
				n.Nextentry = d.Nextentry
				return true
			}
		}
		n = n.Nextentry
		if n.Nextentry == nil {
			n.Nextentry = d
			return true
		}
	}
	return false
}

type Nfs struct {
}

func (t *Nfs) Nfs3_NULL(cargs *sunrpc.CallArags, reply *Void) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_NULL ", fmt.Sprintf("%+v", cargs.Auth_unix))

	return nil
}

func (t *Nfs) Nfs3_GETATTR(cargs *sunrpc.CallArags, reply *GETATTR3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_GETATTR Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args GETATTR3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_GETATTR xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_GETATTR args :", fmt.Sprintf("%+v", args))

	if len(args.Object.Data) == 0 || len(args.Object.Data) < 4 {
		mlog.Warning("Nfs3_GETATTR args object data error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	//	fh := binary.BigEndian.Uint32(args.Object.Data)

	fh, err := operate.LocalGetFh(args.Object.Data)

	if err != nil {
		mlog.Error("Nfs3_GETATTR GetFh error:", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	stat, err := operate.Resc.Stat(operate.Fh(fh))

	if (err != nil && strings.Contains(err.Error(), "no such file or directory") == false) && (err.Error() != "fh not found") {
		mlog.Error("Nfs3_GETATTR operate.Resc.Stat error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	if (err != nil && strings.Contains(err.Error(), "no such file or directory") == true) || (err != nil && err.Error() == "fh not found") {
		mlog.Error("Nfs3_GETATTR no dir error :", err)
		reply.Status = NFS3ERR_STALE
		//reply.Status = NFS3ERR_NOENT
		return nil
	}

	reply.Status = NFS3_OK
	reply.Resok.Obj_attributes.Atime.Nseconds = uint32(stat.Atim.Nsec)
	reply.Resok.Obj_attributes.Atime.Seconds = uint32(stat.Atim.Sec)
	reply.Resok.Obj_attributes.Ctime.Nseconds = uint32(stat.Ctim.Nsec)
	reply.Resok.Obj_attributes.Ctime.Seconds = uint32(stat.Ctim.Sec)
	reply.Resok.Obj_attributes.Fileid = fileid3(stat.Ino)
	reply.Resok.Obj_attributes.Fsid = stat.Dev
	reply.Resok.Obj_attributes.Gid = gid3(stat.Gid)
	reply.Resok.Obj_attributes.Mode = mode3(stat.Mode)
	reply.Resok.Obj_attributes.Mtime.Nseconds = uint32(stat.Mtim.Nsec)
	reply.Resok.Obj_attributes.Mtime.Seconds = uint32(stat.Mtim.Sec)
	reply.Resok.Obj_attributes.Nlink = uint32(stat.Nlink)
	reply.Resok.Obj_attributes.Rdev.Specdata1 = uint32((stat.Rdev >> 8) & 0xFF)
	reply.Resok.Obj_attributes.Rdev.Specdata2 = uint32(stat.Rdev & 0xFF)
	reply.Resok.Obj_attributes.Size = size3(stat.Size)

	if (stat.Mode & syscall.S_IFMT) == syscall.S_IFDIR {
		reply.Resok.Obj_attributes.Type = NF3DIR
	}
	if (stat.Mode & syscall.S_IFMT) == syscall.S_IFSOCK {
		reply.Resok.Obj_attributes.Type = NF3SOCK
	}
	if (stat.Mode & syscall.S_IFMT) == syscall.S_IFLNK {
		reply.Resok.Obj_attributes.Type = NF3LNK
	}
	if (stat.Mode & syscall.S_IFMT) == syscall.S_IFIFO {
		reply.Resok.Obj_attributes.Type = NF3FIFO
	}
	if (stat.Mode & syscall.S_IFMT) == syscall.S_IFREG {
		reply.Resok.Obj_attributes.Type = NF3REG
	}

	reply.Resok.Obj_attributes.Uid = uid3(stat.Uid)
	reply.Resok.Obj_attributes.Used = size3(stat.Blksize * 512)

	mlog.Debug("Nfs3_GETATTR reply :", fmt.Sprintf("%+v", reply))

	return nil
}

func (t *Nfs) Nfs3_SETATTR(cargs *sunrpc.CallArags, reply *SETATTR3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_SETATTR Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args SETATTR3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_SETATTR xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_SETATTR args:", fmt.Sprintf("%+v", args))

	if len(args.Object.Data) == 0 || len(args.Object.Data) < 4 {
		mlog.Debug("Nfs3_SETATTR args error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	//fh := binary.BigEndian.Uint32(args.Object.Data)

	fh, err := operate.LocalGetFh(args.Object.Data)

	if err != nil {
		mlog.Error("Nfs3_SETATTR GetFh error:", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	bef, err := operate.Resc.Stat(operate.Fh(fh))

	if err != nil {
		mlog.Error("Nfs3_SETATTR operate.Resc.Stat bef error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	file, err := operate.Resc.GetFile(operate.Fh(fh))

	if err != nil {
		mlog.Error("Nfs3_SETATTR operate.Resc.GetFile error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	//f := file.(*os.File)

	lfile := file.(operate.LocalFileEx)

	mlog.Debug("Nfs3_SETATTR path :", lfile.Name)

	if args.New_attributes.Size.Set_it == true {
		mlog.Debug("Nfs3_SETATTR syscall.Truncate path :", lfile.Name, " size:", args.New_attributes.Size.Size)
		err = syscall.Truncate(lfile.Name, int64(args.New_attributes.Size.Size))
		if err != nil {
			mlog.Error("Nfs3_SETATTR syscall.Ftruncate error :", err)
			reply.Status = NFS3ERR_INVAL
			return nil
		}
	}

	if args.New_attributes.Gid.Set_it == true || args.New_attributes.Uid.Set_it == true {
		mlog.Debug("Nfs3_SETATTR syscall.Chown path :", lfile.Name, " UID:", args.New_attributes.Uid.Uid, "GID:", args.New_attributes.Gid.Gid)
		err = syscall.Chown(lfile.Name, int(args.New_attributes.Uid.Uid), int(args.New_attributes.Gid.Gid))
		if err != nil {
			mlog.Error("Nfs3_SETATTR syscall.Chown error:", err)
			reply.Status = NFS3ERR_INVAL
			return nil
		}
	}

	if args.New_attributes.Mode.Set_it == true {
		mlog.Debug("Nfs3_SETATTR syscall.Chmod path :", lfile.Name, " size:", uint32(args.New_attributes.Mode.Mode))
		err = syscall.Chmod(lfile.Name, uint32(args.New_attributes.Mode.Mode))
		if err != nil {
			mlog.Error("Nfs3_SETATTR syscall.Chmod error:", err)
			reply.Status = NFS3ERR_INVAL
			return nil
		}
	}

	bef_pos := GetPosAttr(bef)

	if args.New_attributes.Mtime.Set_it != DONT_CHANGE || args.New_attributes.Atime.Set_it != DONT_CHANGE {

		var utimebuf syscall.Utimbuf

		if args.New_attributes.Mtime.Set_it == SET_TO_SERVER_TIME {
			utimebuf.Modtime = time.Now().Unix()
		} else if args.New_attributes.Mtime.Set_it == SET_TO_CLIENT_TIME {
			utimebuf.Modtime = int64(args.New_attributes.Mtime.Mtime.Seconds)
		} else {
			utimebuf.Modtime = int64(bef_pos.Attributes.Mtime.Seconds)
		}

		if args.New_attributes.Atime.Set_it == SET_TO_SERVER_TIME {
			utimebuf.Actime = time.Now().Unix()
		} else if args.New_attributes.Atime.Set_it == SET_TO_CLIENT_TIME {
			utimebuf.Actime = int64(args.New_attributes.Atime.Atime.Seconds)
		} else {
			utimebuf.Actime = int64(bef_pos.Attributes.Atime.Seconds)
		}

		mlog.Debug("Nfs3_SETATTR syscall.Utime path :", lfile.Name, "utimbuf :", fmt.Sprintf("%+v", utimebuf))

		err := syscall.Utime(lfile.Name, &utimebuf)

		if err != nil {
			mlog.Error("Nfs3_SETATTR syscall.Utime error :", err, "path :", lfile.Name)
			reply.Status = NFS3ERR_INVAL
			return nil
		}
	}

	aft, err := operate.Resc.Stat(operate.Fh(fh))

	if err != nil {
		mlog.Error("Nfs3_SETATTR operate.Resc.Stat aft error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	reply.Status = NFS3_OK
	reply.Resok.Obj_wcc.Before = GetPreAttr(bef)
	reply.Resok.Obj_wcc.After = GetPosAttr(aft)

	return nil
}

func (t *Nfs) Nfs3_LOOKUP(cargs *sunrpc.CallArags, reply *LOOKUP3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_LOOKUP Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args LOOKUP3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_LOOKUP xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_LOOKUP args :", fmt.Sprintf("%+v", args))

	if len(args.What.Dir.Data) == 0 || len(args.What.Dir.Data) < 4 {
		mlog.Warning("Nfs3_LOOKUP args object data error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	//fh := binary.BigEndian.Uint32(args.What.Dir.Data)

	fh, err := operate.LocalGetFh(args.What.Dir.Data)

	if err != nil {
		mlog.Error("Nfs3_LOOKUP GetFh error:", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	ofh, stat, dir_s, err := operate.Resc.Look(operate.Fh(fh), args.What.Name)

	if err != nil && strings.Contains(err.Error(), "no such file or directory") == false && err.Error() != "fh not found" {

		mlog.Error("Nfs3_LOOKUP operate.Resc.Look error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	if (err != nil && strings.Contains(err.Error(), "no such file or directory") == true) || (err != nil && err.Error() == "fh not found") {
		mlog.Error("Nfs3_LOOKUP operate.Resc.Look no dir error :", err)
		reply.Status = NFS3ERR_NOENT

		s, err := operate.Resc.Stat(operate.Fh(fh))
		if err != nil {
			mlog.Error("Nfs3_LOOKUP operate.Resc.Look error:", err)
			return nil
		}

		reply.Resfail.Dir_attributes = GetPosAttr(s)

		return nil
	}

	dir_attr := GetPosAttr(dir_s)
	obj_attr := GetPosAttr(stat)

	reply.Status = NFS3_OK
	reply.Resok.Dir_attributes = dir_attr
	reply.Resok.Obj_attributes = obj_attr
	//reply.Resok.Object.Data = make([]byte, 4)

	//binary.BigEndian.PutUint32(reply.Resok.Object.Data, uint32(ofh))

	reply.Resok.Object.Data, err = operate.LocalFhToByte(ofh)

	if err != nil {
		mlog.Error("Nfs3_LOOKUP FhToByte error :", err)
		return nil
	}

	mlog.Debug("Nfs3_LOOKUP reply :", fmt.Sprintf("%+v", reply))

	return nil
}

func (t *Nfs) Nfs3_ACCESS(cargs *sunrpc.CallArags, reply *ACCESS3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_ACCESS Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args ACCESS3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_ACCESS xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_ACCESS args :", fmt.Sprintf("%+v", args))

	if len(args.Object.Data) == 0 || len(args.Object.Data) < 4 {
		mlog.Warning("Nfs3_ACCESS args error")
		reply.Status = NFS3ERR_INVAL
		reply.Resfail.Obj_attributes.Attributes_follow = false
		return nil
	}

	//fh := binary.BigEndian.Uint32(args.Object.Data)

	fh, err := operate.LocalGetFh(args.Object.Data)

	if err != nil {
		mlog.Error("Nfs3_ACCESS GetFh error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	stat, err := operate.Resc.Stat(operate.Fh(fh))
	/*
		if err != nil {
			mlog.Error("Nfs3_ACCESS operate.Resc.Stat error :", err)
			reply.Status = NFS3ERR_INVAL
			return nil
		}
	*/

	if err != nil && strings.Contains(err.Error(), "no such file or directory") == false && err.Error() != "fh not found" {
		mlog.Error("Nfs3_ACCESS operate.Resc.Stat error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	if (err != nil && strings.Contains(err.Error(), "no such file or directory") == true) || (err != nil && err.Error() == "fh not found") {
		mlog.Error("Nfs3_ACCESS no file fh error :", err)
		reply.Status = NFS3ERR_STALE
		return nil
	}

	attr := GetPosAttr(stat)

	reply.Status = NFS3_OK
	reply.Resok.Obj_attributes = attr

	var access uint

	mode := uint32(attr.Attributes.Mode)

	group_flag := false

	for _, i := range cargs.Auth_unix.Gids {
		if uint32(i) == uint32(attr.Attributes.Gid) {
			group_flag = true
			break
		}
	}

	if cargs.Auth_unix.Uid == uint32(attr.Attributes.Uid) {
		if (mode & syscall.S_IRUSR) != 0 {
			access |= ACCESS3_READ
		}
		if (mode & syscall.S_IWUSR) != 0 {
			access |= ACCESS3_MODIFY | ACCESS3_EXTEND
		}
		if (mode & syscall.S_IXUSR) != 0 {
			access |= ACCESS3_EXECUTE
			access |= ACCESS3_READ
		}
	} else if group_flag == true {
		if (mode & syscall.S_IRGRP) != 0 {
			access |= ACCESS3_READ
		}
		if (mode & syscall.S_IWGRP) != 0 {
			access |= ACCESS3_MODIFY | ACCESS3_EXTEND
		}

		if (mode & syscall.S_IXGRP) != 0 {
			access |= ACCESS3_EXECUTE
			access |= ACCESS3_READ
		}
	} else {
		if (mode & syscall.S_IROTH) != 0 {
			access |= ACCESS3_READ
		}

		if (mode & syscall.S_IWOTH) != 0 {
			access |= ACCESS3_MODIFY | ACCESS3_EXTEND
		}

		if (mode & syscall.S_IXOTH) != 0 {
			access |= ACCESS3_EXECUTE
			access |= ACCESS3_READ
		}
	}

	if cargs.Auth_unix.Uid == 0 {
		access |= ACCESS3_READ | ACCESS3_MODIFY | ACCESS3_EXTEND
	}

	if attr.Attributes.Type == NF3DIR {
		if access&(ACCESS3_READ|ACCESS3_EXECUTE) != 0 {
			access |= ACCESS3_LOOKUP
		}

		if access&ACCESS3_MODIFY != 0 {
			access |= ACCESS3_DELETE
		}
		access = access &^ ACCESS3_EXECUTE
	}

	/*
		if attr.Attributes.Type == NF3DIR {
			reply.Resok.Access |= ACCESS3_READ | ACCESS3_MODIFY | ACCESS3_EXTEND | ACCESS3_LOOKUP
			//reply.Resok.Access &= args.Access
		} else {
			reply.Resok.Access |= ACCESS3_READ | ACCESS3_MODIFY | ACCESS3_EXTEND | ACCESS3_EXECUTE
			//reply.Resok.Access &= args.Access
		}
	*/

	reply.Resok.Access = access & args.Access

	mlog.Debug("Nfs3_ACCESS reply:", fmt.Sprintf("%+v", reply))

	return nil
}

func (t *Nfs) Nfs3_READLINK(cargs *sunrpc.CallArags, reply *READLINK3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_READLINK Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args READLINK3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_READLINK xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_READLINK args :", args)

	if len(args.Symlink.Data) == 0 || len(args.Symlink.Data) < 4 {
		mlog.Warning("Nfs3_READLINK args error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	//fh := binary.BigEndian.Uint32(args.Symlink.Data)

	fh, err := operate.LocalGetFh(args.Symlink.Data)

	if err != nil {
		mlog.Error("Nfs3_READLINK GetFh error:", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mb := make([]byte, 1024)

	mlog.Debug("Nfs3_READLINK fh :", fh)

	stat, err := operate.Resc.ReadLink(operate.Fh(fh), mb)

	if err != nil {
		mlog.Error("Nfs3_READLINK operate.Resc.ReadLink error:", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	reply.Status = NFS3_OK
	reply.Resok.Symlink_attributes = GetPosAttr(stat)
	reply.Resok.Data = nfspath3(mb)

	return nil
}

func (t *Nfs) Nfs3_READ(cargs *sunrpc.CallArags, reply *READ3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_READ Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args READ3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_READ xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_READ args :", fmt.Sprintf("%+v", args))

	//fh := binary.BigEndian.Uint32(args.File.Data)

	fh, err := operate.LocalGetFh(args.File.Data)

	if err != nil {
		mlog.Error("Nfs3_READ GetFh error:", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	b := make([]byte, args.Count)

	n, fs, err := operate.Resc.Read(operate.Fh(fh), b, int64(args.Count), int64(args.Offset))

	if err != nil && err != io.EOF {
		mlog.Error("Nfs3_READ operate.Resc.Read error :", err)
		reply.Status = NFS3ERR_IO
		return nil
	}

	reply.Status = NFS3_OK
	reply.Resok.Count = count3(n)
	reply.Resok.Data = b

	if err != nil && err == io.EOF {
		mlog.Warning("Nfs3_READ ((((((( eror ")
		reply.Resok.Eof = true
	}

	if int64(args.Offset)+int64(n) >= fs.Size {
		mlog.Debug("Nfs3_READ )))))) eror n :", n, "count :", args.Count, "szie :", fs.Size, "offset :", args.Offset)
		reply.Resok.Eof = true
	}

	attr := GetPosAttr(fs)
	reply.Resok.File_attributes = attr

	return nil
}

func (t *Nfs) Nfs3_WRITE(cargs *sunrpc.CallArags, reply *WRITE3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_WRITE Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args WRITE3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_WRITE xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_WRITE file :", args.File.Data)
	mlog.Debug("Nfs3_WRITE @@@ statbel :", args.Stable, "count:", args.Count, "offset :", args.Offset, "len:", len(args.Data))

	//fh := binary.BigEndian.Uint32(args.File.Data)

	fh, err := operate.LocalGetFh(args.File.Data)

	if err != nil {
		mlog.Error("Nfs3_WRITE GetFh error:", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	b := make([]byte, args.Count)

	b = append(b[0:0], args.Data...)

	if uint32(args.Count) != uint32(len(b)) {
		mlog.Error("Nfs3_WRITE ******* args.count :", args.Count, "len :", len(b))
	}

	n, bfs, afs, err := operate.Resc.Write(operate.Fh(fh), b, int64(args.Count), int64(args.Offset))

	if err != nil && err != io.EOF {
		mlog.Error("Nfs3_WRITE operate.Resc.Read error :", err)
		reply.Status = NFS3ERR_IO
		return nil
	}

	mlog.Debug("Nfs3_WRITE write n :", n)

	if uint32(len(args.Data)) != uint32(args.Count) {
		mlog.Error("************ what")
	}

	if uint32(args.Count) != uint32(n) {
		mlog.Error("&&&&&&&&&&&& what")
	}

	reply.Status = NFS3_OK
	reply.Resok.Count = count3(n)
	reply.Resok.Committed = FILE_SYNC
	reply.Resok.File_wcc.After = GetPosAttr(afs)
	reply.Resok.File_wcc.Before = GetPreAttr(bfs)

	return nil
}

func (t *Nfs) Nfs3_CREATE(cargs *sunrpc.CallArags, reply *CREATE3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_CREATE Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args CREATE3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_CREATE xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_CREATE args :", fmt.Sprintf("%+v", args))

	if len(args.Where.Dir.Data) == 0 || len(args.Where.Dir.Data) < 4 {
		mlog.Warning("Nfs3_CREATE args error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	//fh := binary.BigEndian.Uint32(args.Where.Dir.Data)

	fh, err := operate.LocalGetFh(args.Where.Dir.Data)

	if err != nil {
		mlog.Error("Nfs3_CREATE GetFh error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	var file_mode uint32

	if args.How.Obj_attributes.Mode.Set_it == true {
		file_mode = uint32(args.How.Obj_attributes.Mode.Mode)
	}

	nfh, obj, aft, bef, err := operate.Resc.Create(operate.Fh(fh), args.Where.Name, int(args.How.Mode), file_mode, args.How.Verf, int(cargs.Auth_unix.Uid), int(cargs.Auth_unix.Gid))

	if err != nil && err.Error() != "file exist" {
		mlog.Error("Nfs3_CREATE operate.Resc.Create error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	if err != nil && err.Error() == "file exist" {
		mlog.Error("Nfs3_CREATE operate.Resc.Create file exist")
		reply.Status = NFS3ERR_EXIST
		return nil
	}

	aft_attr := GetPosAttr(aft)
	bef_attr := GetPreAttr(bef)
	obj_attr := GetPosAttr(obj)

	//binary.BigEndian.PutUint32(reply.Resok.Obj.Handle.Data, uint32(nfh))

	reply.Resok.Obj.Handle.Data, err = operate.LocalFhToByte(nfh)

	if err != nil {
		mlog.Error("Nfs3_CREATE FhToByte error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	reply.Status = NFS3_OK
	reply.Resok.Dir_wcc.After = aft_attr
	reply.Resok.Dir_wcc.Before = bef_attr
	//reply.Resok.Obj.Handle.Data = make([]byte, 4)

	reply.Resok.Obj.Handle_follows = true
	reply.Resok.Obj_attributes = obj_attr

	mlog.Debug("Nfs3_CREATE reply:", fmt.Sprintf("%+v", reply))

	return nil
}

func (t *Nfs) Nfs3_MKDIR(cargs *sunrpc.CallArags, reply *MKDIR3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_MKDIR Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args MKDIR3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_MKDIR xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_MKDIR args :", args)

	if len(args.Where.Dir.Data) == 0 || len(args.Where.Dir.Data) < 4 {
		mlog.Warning("Nfs3_MKDIR args error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	//fh := binary.BigEndian.Uint32(args.Where.Dir.Data)

	fh, err := operate.LocalGetFh(args.Where.Dir.Data)

	if err != nil {
		mlog.Error("Nfs3_MKDIR GetFh error:", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	var attr operate.Setttr

	attr.Gid = uint32(args.Attributes.Gid.Gid)
	attr.Mode = uint32(args.Attributes.Mode.Mode)
	attr.Uid = uint32(args.Attributes.Uid.Uid)

	nfh, nattr, bef, aft, err := operate.Resc.Mkdir(operate.Fh(fh), args.Where.Name, attr, int(cargs.Auth_unix.Uid), int(cargs.Auth_unix.Gid))

	if err != nil && err.Error() != operate.FILE_EXISTS {
		mlog.Error("Nfs3_MKDIR operate.Resc.Mkdir error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	if err != nil && err.Error() == operate.FILE_EXISTS {
		mlog.Warning("Nfs3_MKDIR operate.Resc.Mkdir file exists")
		reply.Status = NFS3ERR_EXIST
		reply.Resfail.Dir_wcc.After = GetPosAttr(aft)
		reply.Resfail.Dir_wcc.Before = GetPreAttr(bef)
		return nil
	}

	//reply.Resok.Obj.Handle.Data = make([]byte, 4)

	//binary.BigEndian.PutUint32(reply.Resok.Obj.Handle.Data, uint32(nfh))

	reply.Resok.Obj.Handle.Data, err = operate.LocalFhToByte(nfh)

	if err != nil {
		mlog.Error("Nfs3_MKDIR FhToByte error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	reply.Status = NFS3_OK

	reply.Resok.Dir_wcc.After = GetPosAttr(aft)
	reply.Resok.Dir_wcc.Before = GetPreAttr(bef)
	reply.Resok.Obj_attributes = GetPosAttr(nattr)

	reply.Resok.Obj.Handle_follows = true

	mlog.Debug("Nfs3_MKDIR reply :", fmt.Sprintf("%+v", reply))

	return nil
}

func (t *Nfs) Nfs3_SYMLINK(cargs *sunrpc.CallArags, reply *SYMLINK3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_SYMLINK Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args SYMLINK3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_SYMLINK xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_SYMLINK args :", args)

	if len(args.Where.Dir.Data) == 0 || len(args.Where.Dir.Data) < 4 {
		mlog.Warning("Nfs3_SYMLINK args error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	//fh := binary.BigEndian.Uint32(args.Where.Dir.Data)

	fh, err := operate.LocalGetFh(args.Where.Dir.Data)

	if err != nil {
		mlog.Error("Nfs3_SYMLINK GetFh error:", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	nfh, nattr, bef, aft, err := operate.Resc.Symlink(operate.Fh(fh), args.Where.Name, string(args.Symlink.Symlink_data), int(cargs.Auth_unix.Uid), int(cargs.Auth_unix.Gid))

	if err != nil && err.Error() != operate.FILE_EXISTS {
		mlog.Error("Nfs3_SYMLINK operate.Resc.Symlink error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	if err != nil && err.Error() == operate.FILE_EXISTS {
		mlog.Error("Nfs3_SYMLINK operate.Resc.Symlink error :", err)
		reply.Status = NFS3ERR_EXIST
		return nil
	}

	//reply.Resok.Obj.Handle.Data = make([]byte, 4)

	//	binary.BigEndian.PutUint32(reply.Resok.Obj.Handle.Data, uint32(nfh))

	reply.Resok.Obj.Handle.Data, err = operate.LocalFhToByte(nfh)

	if err != nil {
		mlog.Error("Nfs3_SYMLINK FhToByte error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	reply.Status = NFS3_OK
	reply.Resok.Obj.Handle_follows = true

	reply.Resok.Dir_wcc.Before = GetPreAttr(bef)
	reply.Resok.Dir_wcc.After = GetPosAttr(aft)
	reply.Resok.Obj_attributes = GetPosAttr(nattr)

	return nil
}

func (t *Nfs) Nfs3_MKNODE(cargs *sunrpc.CallArags, reply *MKNOD3res) error {

	defer PinacRevove()

	mlog.Error("Nfs3_MKNODE Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args MKNOD3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_MKNODE xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	return nil
}

func (t *Nfs) Nfs3_REMOVE(cargs *sunrpc.CallArags, reply *REMOVE3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_REMOVE Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args REMOVE3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_REMOVE xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_REMOVE args :", args)

	if len(args.Object.Dir.Data) == 0 || len(args.Object.Dir.Data) < 4 {
		mlog.Warning("Nfs3_REMOVE args error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	//fh := binary.BigEndian.Uint32(args.Object.Dir.Data)

	fh, err := operate.LocalGetFh(args.Object.Dir.Data)

	if err != nil {
		mlog.Error("Nfs3_REMOVE GetFh error:", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	bs, as, err := operate.Resc.Remove(operate.Fh(fh), args.Object.Name)

	if err != nil {
		mlog.Error("Nfs3_REMOVE operate.Resc.Remove error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	reply.Status = NFS3_OK
	reply.Resok.Dir_wcc.After = GetPosAttr(as)
	reply.Resok.Dir_wcc.Before = GetPreAttr(bs)

	return nil
}

func (t *Nfs) Nfs3_RMDIR(cargs *sunrpc.CallArags, reply *RMDIR3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_RMDIR Begin auth :", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args RMDIR3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_RMDIR xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_RMDIR args :", args)

	if len(args.Object.Dir.Data) == 0 || len(args.Object.Dir.Data) < 4 {
		mlog.Warning("Nfs3_RMDIR args error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	//fh := binary.BigEndian.Uint32(args.Object.Dir.Data)

	fh, err := operate.LocalGetFh(args.Object.Dir.Data)

	if err != nil {
		mlog.Error("Nfs3_RMDIR GetFh error:", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	bef, aft, err := operate.Resc.Remove(operate.Fh(fh), args.Object.Name)

	if err != nil {
		mlog.Error("Nfs3_RMDIR operate.Resc.Remove error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	reply.Status = NFS3_OK
	reply.Resok.Dir_wcc.Before = GetPreAttr(bef)
	reply.Resok.Dir_wcc.After = GetPosAttr(aft)

	return nil
}

func (t *Nfs) Nfs3_RENAME(cargs *sunrpc.CallArags, reply *RENAME3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_RENAME Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args RENAME3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_RENAME xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_RENAME args:", args)

	if len(args.From.Dir.Data) == 0 || len(args.From.Dir.Data) < 4 {
		mlog.Warning("Nfs3_RENAME args from dir error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	if len(args.To.Dir.Data) == 0 || len(args.To.Dir.Data) < 4 {
		mlog.Warning("Nfs3_RENAME args to dir error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	//ofh := binary.BigEndian.Uint32(args.From.Dir.Data)
	//nfh := binary.BigEndian.Uint32(args.To.Dir.Data)

	ofh, err := operate.LocalGetFh(args.From.Dir.Data)

	if err != nil {
		mlog.Error("Nfs3_RENAME GetFh from error:", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	nfh, err := operate.LocalGetFh(args.To.Dir.Data)

	if err != nil {
		mlog.Error("Nfs3_RENAME GetFh to error:", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	fbef, tbef, faft, taft, err := operate.Resc.Rename(operate.Fh(ofh), operate.Fh(nfh), string(args.From.Name), string(args.To.Name))

	if err != nil {
		mlog.Error("Nfs3_RENAME operate.Resc.Rename error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	reply.Status = NFS3_OK
	reply.Resok.Fromdir_wcc.After = GetPosAttr(faft)
	reply.Resok.Fromdir_wcc.Before = GetPreAttr(fbef)
	reply.Resok.Todir_wcc.After = GetPosAttr(taft)
	reply.Resok.Todir_wcc.Before = GetPreAttr(tbef)

	return nil
}

func (t *Nfs) Nfs3_LINK(cargs *sunrpc.CallArags, reply *LINK3res) error {

	defer PinacRevove()

	mlog.Error("Nfs3_LINK Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args LINK3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_LINK xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_LINK args :", fmt.Sprintf("%+v", args))

	return nil
}

func (t *Nfs) Nfs3_READDIR(cargs *sunrpc.CallArags, reply *READDIR3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_READDIR Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args READDIR3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_READDIR xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_READDIR args :", fmt.Sprintf("%+v", args))

	if len(args.Dir.Data) == 0 || len(args.Dir.Data) < 4 {
		mlog.Warning("Nfs3_READDIR args error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	//fh := binary.BigEndian.Uint32(args.Dir.Data)

	fh, err := operate.LocalGetFh(args.Dir.Data)

	if err != nil {
		mlog.Error("Nfs3_READDIR GetFh error:", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	infos, err := operate.Resc.ReadDir(operate.Fh(fh))

	if err != nil {
		mlog.Debug("Nfs3_READDIR operate.Resc.ReadDir error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	dir_stat, err := operate.Resc.Stat(operate.Fh(fh))

	if err != nil {
		mlog.Debug("Nfs3_READDIR operate.Resc.Stat error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	dir_attr := GetPosAttr(dir_stat)

	var cookie cookie3

	cookie = args.Cookie
	/*
		upper := cookie & 0xFFFFFFFF
		if cookie != 0 && upper != 0 {
			cookie = 0
		}
		cookie &= 0xFFFFFFFF
	*/

	reply.Status = NFS3_OK
	reply.Resok.Dir_attributes = dir_attr

	var cookievef []byte

	cookievef = args.Cookieverf[:]

	num := binary.BigEndian.Uint64(cookievef)

	mlog.Debug("Nfs3_READDIR num :", num, " cookies:", cookie)

	if len(infos) < 0 {
		mlog.Warning("Nfs3_READDIR dir infos empty")
	}

	if uint64(len(infos)) > uint64(cookie) {
		mlog.Warning("Nfs3_READDIR cookie > len infos")
	}

	if len(infos) > 0 && uint64(len(infos)) > uint64(cookie) {

		var i int
		var k int

		head := new(entry3)
		head.Name = filename3(infos[cookie].Name)
		//head.Cookie = (cookie + cookie3(2) + cookie3(num)) | 0
		head.Cookie = cookie3(cookie + 1)
		head.Fileid = fileid3(infos[cookie].Stat.Ino)
		/*
			if num != uint64(cookie) {
				reply.Status = NFS3ERR_BAD_COOKIE
				reply.Resfail2.Dir_attributes = dir_attr
				return nil
			}
		*/
		mlog.Debug("Nfs3_READDIR head item :", fmt.Sprintf("%+v", head))

		for i = int(cookie) + 1; i < len(infos) && k < 10; i++ {

			nne := new(entry3)
			nne.Name = filename3(infos[i].Name)
			//nne.Cookie = (cookie + cookie3(2) + cookie3(i)) | 0
			//nne.Cookie = (cookie3(num) + cookie3(2) + cookie3(j))
			nne.Fileid = fileid3(infos[i].Stat.Ino)
			nne.Cookie = cookie3(i)

			InsertEntry3(head, nne, k+1)

			mlog.Debug("Nfs3_READDIR item :", fmt.Sprintf("%+v", nne))
			k++
		}

		reply.Resok.Reply.Entries = head

		if len(infos) > i {
			reply.Resok.Reply.Eof = false
		} else {
			reply.Resok.Reply.Eof = true
		}
		/*
			var b []byte
			b = make([]byte, 8)
			binary.BigEndian.PutUint64(b, uint64(i))

			for i, _ := range reply.Resok.Cookieverf {
				reply.Resok.Cookieverf[i] = b[i]
			}
		*/
	} else {
		reply.Resok.Reply.Eof = true
		mlog.Debug("Nfs3_READDIR empty dir")
	}

	mlog.Debug("Nfs3_READDIR reply :", fmt.Sprintf("%+v", reply))

	var en *entry3

	en = reply.Resok.Reply.Entries

	for en != nil {
		mlog.Debug("debug ***** en ", fmt.Sprintf("%+v", en))
		en = en.Nextentry
	}

	return nil
}

func (t *Nfs) Nfs3_READDIRPLUS(cargs *sunrpc.CallArags, reply *READDIRPLUS3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_READDIRPLUS Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args READDIRPLUS3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_READDIRPLUS xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_READDIRPLUS args:", args)

	if len(args.Dir.Data) == 0 || len(args.Dir.Data) < 4 {
		mlog.Warning("Nfs3_READDIRPLUS args error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	reply.Status = NFS3ERR_NOTSUPP
	reply.Resfail.Dir_attributes.Attributes_follow = false

	return nil

	//fh := binary.BigEndian.Uint32(args.Dir.Data)
	/*
		fh, err := operate.LocalGetFh(args.Dir.Data)

		if err != nil {
			mlog.Error("Nfs3_READDIRPLUS GetFh error:", err)
			reply.Status = NFS3ERR_INVAL
			return nil
		}

		infos, err := operate.Resc.ReadDirPlus(operate.Fh(fh))

		if err != nil {
			mlog.Debug("Nfs3_READDIRPLUS operate.Resc.ReadDir error :", err)
			reply.Status = NFS3ERR_INVAL
			return nil
		}

		dir_stat, err := operate.Resc.Stat(operate.Fh(fh))

		if err != nil {
			mlog.Debug("Nfs3_READDIRPLUS operate.Resc.Stat error :", err)
			reply.Status = NFS3ERR_INVAL
			return nil
		}

		dir_attr := GetPosAttr(dir_stat)

		var cookie cookie3

		cookie = args.Cookie

		upper := cookie & 0xFFFFFFFF
		if cookie != 0 && upper != 0 {
			cookie = 0
		}
		cookie &= 0xFFFFFFFF

		reply.Status = NFS3_OK
		reply.Resok.Dir_attributes = dir_attr

		var cookievef []byte

		cookievef = args.Cookieverf[:]

		num := binary.BigEndian.Uint64(cookievef)

		mlog.Debug("Nfs3_READDIRPLUS num :", num)

		if len(infos) > 0 && uint64(len(infos)) > uint64(num) {

			var i int
			var k int

			head := new(entryplus3)
			head.Name = filename3(infos[num].Name)
			head.Cookie = (cookie + 1) | 0
			head.Fileid = fileid3(infos[num].Stat.Ino)
			head.Name_attributes = GetPosAttr(infos[num].Stat)
			head.Name_handle.Handle.Data, _ = operate.LocalFhToByte(infos[num].Fh)
			head.Name_handle.Handle_follows = true

			for i = int(num) + 1; i < len(infos) && k < 10; i++ {

				nne := new(entryplus3)
				nne.Name = filename3(infos[i].Name)
				nne.Cookie = (cookie + cookie3(1) + cookie3(i)) | 0
				nne.Fileid = fileid3(infos[i].Stat.Ino)
				nne.Name_attributes = GetPosAttr(infos[i].Stat)

				nne.Name_handle.Handle.Data, _ = operate.LocalFhToByte(infos[i].Fh)
				nne.Name_handle.Handle_follows = true

				InsertEntryPlus3(head, nne, k+1)

				mlog.Debug("Nfs3_READDIRPLUS item :", fmt.Sprintf("%+v", nne))
				k++
			}

			reply.Resok.Reply.Entries = head

			if len(infos) > i {
				reply.Resok.Reply.Eof = false
			} else {
				reply.Resok.Reply.Eof = true
			}

			var b []byte
			b = make([]byte, 8)
			binary.BigEndian.PutUint64(b, uint64(i))

			for i, _ := range reply.Resok.Cookieverf {
				reply.Resok.Cookieverf[i] = b[i]
			}

		} else {
			reply.Resok.Reply.Eof = true
			mlog.Debug("Nfs3_READDIRPLUS empty dir")
		}

		mlog.Debug("Nfs3_READDIRPLUS reply :", fmt.Sprintf("%+v", reply))

		return nil
	*/
}

func (t *Nfs) Nfs3_FSSTAT(cargs *sunrpc.CallArags, reply *FSSTAT3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_FSSTAT Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args FSSTAT3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_FSSTAT xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_FSSTAT args :", args)

	if len(args.Fsroot.Data) == 0 || len(args.Fsroot.Data) < 4 {
		mlog.Warning("Nfs3_FSSTAT args error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	//fh := binary.BigEndian.Uint32(args.Fsroot.Data)

	fh, err := operate.LocalGetFh(args.Fsroot.Data)

	if err != nil {
		mlog.Error("Nfs3_FSSTAT GetFh err :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	file, err := operate.Resc.GetFile(operate.Fh(fh))

	if err != nil {
		mlog.Error("Nfs3_FSSTAT operate.Resc.GetFile error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	f := file.(operate.LocalFileEx)

	dir := f.Name

	var buf syscall.Statfs_t

	err = syscall.Statfs(dir, &buf)

	if err != nil {
		mlog.Error("Nfs3_FSSTAT syscall.Statfs error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	reply.Resok.Abytes = size3(buf.Bavail * uint64(buf.Frsize))
	reply.Resok.Tbytes = size3(buf.Blocks * uint64(buf.Frsize))
	reply.Resok.Fbytes = size3(buf.Bfree * uint64(buf.Frsize))
	reply.Resok.Tfbytes = size3(buf.Files)
	reply.Resok.Ffbytes = size3(buf.Ffree)
	reply.Resok.Afbytes = size3(buf.Ffree)
	reply.Resok.Invarsec = 0

	return nil
}

func (t *Nfs) Nfs3_FSINFO(cargs *sunrpc.CallArags, reply *FSINFO3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_FSINFO Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args FSINFO3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_FSINFO xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_FSINFO args :", fmt.Sprintf("%+v", args))

	reply.Status = NFS3_OK

	//fh := binary.BigEndian.Uint32(args.Fsroot.Data)

	fh, err := operate.LocalGetFh(args.Fsroot.Data)

	if err != nil {
		mlog.Error("Nfs3_FSINFO GetFh error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_FSINFO fh :", fh)

	stat, err := operate.Resc.Stat(fh)

	if err != nil {
		mlog.Error("Nfs3_FSINFO operate.Resc.Stat error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}
	attr := GetPosAttr(stat)
	reply.Resok.Obj_attributes = attr

	reply.Resok.Maxfilesize64 = math.MaxUint64
	reply.Resok.Dtpref = 4096
	reply.Resok.Rtmax = 524288
	reply.Resok.Rtmult = 4096
	reply.Resok.Rtpref = 524288
	reply.Resok.Wtmax = 524288
	reply.Resok.Wtmult = 4096
	reply.Resok.Wtpref = 524288

	/*
		reply.Resok.Maxfilesize64 = math.MaxUint64
		reply.Resok.Dtpref = 512
		reply.Resok.Rtmax = 65535
		reply.Resok.Rtmult = 512
		reply.Resok.Rtpref = 65535
		reply.Resok.Wtmax = 65535
		reply.Resok.Wtmult = 512
		reply.Resok.Wtpref = 65535
	*/
	/*
		reply.Resok.Maxfilesize64 = math.MaxUint64
		reply.Resok.Dtpref = 12288
		reply.Resok.Rtmax = 524288
		reply.Resok.Rtmult = 12288
		reply.Resok.Rtpref = 524288
		reply.Resok.Wtmax = 524288
		reply.Resok.Wtmult = 12288
		reply.Resok.Wtpref = 524288
	*/
	reply.Resok.Time_delta.Seconds = 1
	reply.Resok.Properties = FSF3_LINK | FSF3_SYMLINK | FSF3_HOMOGENEOUS | FSF3_CANSETTIME

	return nil
}

func (t *Nfs) Nfs3_PATHCONF(cargs *sunrpc.CallArags, reply *PATHCONF3res) error {

	defer PinacRevove()

	mlog.Debug("Nfs3_PATHCONF Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args PATHCONF3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_PATHCONF xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfs3_PATHCONF")

	reply.Status = NFS3_OK
	reply.Resok.Namemax = 1024
	reply.Resok.Linkmax = 0xFFFFFFFF
	reply.Resok.No_trunc = true
	reply.Resok.Case_preserving = true

	return nil
}

func (t *Nfs) Nfs3_COMMIT(cargs *sunrpc.CallArags, reply *COMMIT3res) error {

	defer PinacRevove()

	mlog.Error("Nfs3_COMMIT Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args COMMIT3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfs3_COMMIT xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	reply.Status = NFS3_OK

	return nil
}
