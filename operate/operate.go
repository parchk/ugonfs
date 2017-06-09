package operate

import (
	"syscall"
	"time"
)

const (
	FILE_EXISTS = "file exists"
)

var Resc Resouce

func init() {
	Resc = NewLocalFileResourceEx()
}

func init() {
	go TimerStop()
}

func TimerStop() {
	for {
		t1 := time.NewTimer(time.Second * 10)

		for {
			select {
			case <-t1.C:
				println("10s timer")
				t1.Reset(time.Second * 10)
				Resc.FileDump()
			}
		}
	}
}

type Setttr struct {
	Mode uint32
	Uid  uint32
	Gid  uint32
	Size uint64
}

type Fh interface{}

type DirItemInfo struct {
	Name string
	Stat *syscall.Stat_t
	Fh   Fh
}

type Resouce interface {
	Mnt(dirpath string, clientid string) (Fh, error)
	Unmnt(dirpath string) error
	GetFile(fh Fh) (FileBase, error)
	Look(fh Fh, name string) (Fh, *syscall.Stat_t, *syscall.Stat_t, error)
	Stat(fh Fh) (*syscall.Stat_t, error)
	ReadLink(fh Fh, b []byte) (*syscall.Stat_t, error)
	Read(fh Fh, p []byte, count int64, offset int64) (int, *syscall.Stat_t, error)
	ReadDir(fh Fh) ([]DirItemInfo, error)
	ReadDirPlus(fh Fh) ([]DirItemInfo, error)
	Write(fh Fh, p []byte, count int64, offset int64) (n int, bef *syscall.Stat_t, aft *syscall.Stat_t, err error)
	Create(fh Fh, name string, mode int, file_mode uint32, verf [8]byte, uid int, gid int) (obj_fh Fh, obj_s *syscall.Stat_t, bef *syscall.Stat_t, aft *syscall.Stat_t, err error)
	Remove(fh Fh, name string) (bef *syscall.Stat_t, aft *syscall.Stat_t, err error)
	Rename(ofh Fh, nfh Fh, old, new_name string) (fbef *syscall.Stat_t, faft *syscall.Stat_t, tbef *syscall.Stat_t, taft *syscall.Stat_t, err error)
	Mkdir(fh Fh, name string, attr Setttr, uid int, gid int) (nfh Fh, objs *syscall.Stat_t, bef *syscall.Stat_t, aft *syscall.Stat_t, err error)
	Symlink(fh Fh, name string, path string, uid int, gid int) (nfh Fh, objs *syscall.Stat_t, bef *syscall.Stat_t, aft *syscall.Stat_t, err error)
	Link(fh Fh, name string, path string, uid int, gid int) (nfh Fh, objs *syscall.Stat_t, bef *syscall.Stat_t, aft *syscall.Stat_t, err error)
	FileDump()
}
