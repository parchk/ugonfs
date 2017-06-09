package nfspro

const NFS3_FHSIZE = 64
const NFS3_WRITEVERFSIZE = 8
const NFS3_CREATEVERFSIZE = 8
const NFS3_COOKIEVERFSIZE = 8

type offset3 uint64
type cookie3 uint64
type cookieverf3 [NFS3_COOKIEVERFSIZE]byte
type count3 uint32
type fileid3 uint64
type filename3 string
type mode3 uint32
type uid3 uint32
type gid3 uint32
type createverf3 [NFS3_CREATEVERFSIZE]byte
type nfspath3 string
type writeverf3 [NFS3_WRITEVERFSIZE]byte

type nfsstat3 int32

const (
	NFS3_OK             nfsstat3 = 0
	NFS3ERR_PERM                 = 1
	NFS3ERR_NOENT                = 2
	NFS3ERR_IO                   = 5
	NFS3ERR_NXIO                 = 6
	NFS3ERR_ACCES                = 13
	NFS3ERR_EXIST                = 17
	NFS3ERR_XDEV                 = 18
	NFS3ERR_NODEV                = 19
	NFS3ERR_NOTDIR               = 20
	NFS3ERR_ISDIR                = 21
	NFS3ERR_INVAL                = 22
	NFS3ERR_FBIG                 = 27
	NFS3ERR_NOSPC                = 28
	NFS3ERR_ROFS                 = 30
	NFS3ERR_MLINK                = 31
	NFS3ERR_NAMETOOLONG          = 63
	NFS3ERR_NOTEMPTY             = 66
	NFS3ERR_DQUOT                = 69
	NFS3ERR_STALE                = 70
	NFS3ERR_REMOTE               = 71
	NFS3ERR_BADHANDLE            = 10001
	NFS3ERR_NOT_SYNC             = 10002
	NFS3ERR_BAD_COOKIE           = 10003
	NFS3ERR_NOTSUPP              = 10004
	NFS3ERR_TOOSMALL             = 10005
	NFS3ERR_SERVERFAULT          = 10006
	NFS3ERR_BADTYPE              = 10007
	NFS3ERR_JUKEBOX              = 10008
)

type nfs_fh3 struct {
	Data []byte
}

type specdata3 struct {
	Specdata1 uint32
	Specdata2 uint32
}

type nfstime3 struct {
	Seconds  uint32
	Nseconds uint32
}

type ftype3 int32

const (
	NF3REG  ftype3 = 1
	NF3DIR         = 2
	NF3BLK         = 3
	NF3CHR         = 4
	NF3LNK         = 5
	NF3SOCK        = 6
	NF3FIFO        = 7
)

type time_how int32

const (
	DONT_CHANGE        time_how = 0
	SET_TO_SERVER_TIME          = 1
	SET_TO_CLIENT_TIME          = 2
)

const (
	ACCESS3_READ    = 0x0001
	ACCESS3_LOOKUP  = 0x0002
	ACCESS3_MODIFY  = 0x0004
	ACCESS3_EXTEND  = 0x0008
	ACCESS3_DELETE  = 0x0010
	ACCESS3_EXECUTE = 0x0020
)

const (
	FSF3_LINK        = 0x0001
	FSF3_SYMLINK     = 0x0002
	FSF3_HOMOGENEOUS = 0x0008
	FSF3_CANSETTIME  = 0x0010
)

type stable_how int32

const (
	UNSTABLE  stable_how = 0
	DATA_SYNC            = 1
	FILE_SYNC            = 2
)

type createmode3 int32

const (
	UNCHECKED = 0
	GUARDED   = 1
	EXCLUSIVE = 2
)

type fattr3 struct {
	Type   ftype3
	Mode   mode3
	Nlink  uint32
	Uid    uid3
	Gid    gid3
	Size   size3
	Used   size3
	Rdev   specdata3
	Fsid   uint64
	Fileid fileid3
	Atime  nfstime3
	Mtime  nfstime3
	Ctime  nfstime3
}

type set_mode3 struct {
	Set_it bool  `xdr:"union"`
	Mode   mode3 `xdr:"unioncase=1"`
}

type set_uid3 struct {
	Set_it bool `xdr:"union"`
	Uid    uid3 `xdr:"unioncase=1"`
}

type set_gid3 struct {
	Set_it bool `xdr:"union"`
	Gid    gid3 `xdr:"unioncase=1"`
}

type set_size3 struct {
	Set_it bool  `xdr:"union"`
	Size   size3 `xdr:"unioncase=1"`
}

type set_atime struct {
	Set_it time_how `xdr:"union"`
	Atime  nfstime3 `xdr:"unioncase=2"`
}

type set_mtime struct {
	Set_it time_how `xdr:"union"`
	Mtime  nfstime3 `xdr:"unioncase=2"`
}

type sattr3 struct {
	Mode  set_mode3
	Uid   set_uid3
	Gid   set_gid3
	Size  set_size3
	Atime set_atime
	Mtime set_mtime
}

type size3 uint64

type wcc_attr struct {
	Size  size3
	Mtime nfstime3
	Ctime nfstime3
}

type pre_op_attr struct {
	Attributes_follow bool     `xdr:"union"`
	Attributes        wcc_attr `xdr:"unioncase=1"`
}

type wcc_data struct {
	Before pre_op_attr
	After  post_op_attr
}

type diropargs3 struct {
	Dir  nfs_fh3
	Name string
}

type post_op_attr struct {
	Attributes_follow bool   `xdr:"union"` //hsh
	Attributes        fattr3 `xdr:"unioncase=1"`
}

//////GETATTR
type GETATTR3resok struct {
	Obj_attributes fattr3
}

type GETATTR3res struct {
	Status nfsstat3      `xdr:"union"`
	Resok  GETATTR3resok `xdr:"unioncase=0"`
}

type GETATTR3args struct {
	Object nfs_fh3
}

////SETATTR
type SETATTR3resok struct {
	Obj_wcc wcc_data
}

type SETATTR3resfail struct {
	Obj_wcc wcc_data
}

type SETATTR3res struct {
	Status  nfsstat3        `xdr:"union"`
	Resok   SETATTR3resok   `xdr:"unioncase=0"`
	Resfail SETATTR3resfail `xdr:"unioncase=-1"`
}

type sattrguard3 struct {
	Check     bool     `xdr:"union"`
	Obj_ctime nfstime3 `xdr:"unioncase=1"`
}

type SETATTR3args struct {
	Object         nfs_fh3
	New_attributes sattr3
	Guard          sattrguard3
}

//LOOKUP
type LOOKUP3resok struct {
	Object         nfs_fh3
	Obj_attributes post_op_attr
	Dir_attributes post_op_attr
}

type LOOKUP3resfail struct {
	Dir_attributes post_op_attr
}

type LOOKUP3args struct {
	What diropargs3
}

type LOOKUP3res struct {
	Status  nfsstat3       `xdr:"union"`
	Resok   LOOKUP3resok   `xdr:"unioncase=0"`
	Resfail LOOKUP3resfail `xdr:"unioncase=2"`
}

//ACCESS

type ACCESS3args struct {
	Object nfs_fh3
	Access uint
}

type ACCESS3resok struct {
	Obj_attributes post_op_attr
	Access         uint
}

type ACCESS3resfail struct {
	Obj_attributes post_op_attr
}

type ACCESS3res struct {
	Status  nfsstat3       `xdr:"union"`
	Resok   ACCESS3resok   `xdr:"unioncase=0"`
	Resfail ACCESS3resfail `xdr:"union=default"`
}

//READLINK
type READLINK3args struct {
	Symlink nfs_fh3
}

type READLINK3resok struct {
	Symlink_attributes post_op_attr
	Data               nfspath3
}

type READLINK3resfail struct {
	Symlink_attributes post_op_attr
}

type READLINK3res struct {
	Status  nfsstat3         `xdr:"union"`
	Resok   READLINK3resok   `xdr:"unioncase=0"`
	Resfail READLINK3resfail `xdr:"unioncase=-1"`
}

//READ
type READ3args struct {
	File   nfs_fh3
	Offset offset3
	Count  count3
}

type READ3reasok struct {
	File_attributes post_op_attr
	Count           count3
	Eof             bool
	Data            []byte
}

type READ3resfail struct {
	File_attributes post_op_attr
}

type READ3res struct {
	Status  nfsstat3     `xdr:"union"`
	Resok   READ3reasok  `xdr:"unioncase=0"`
	Resfail READ3resfail `xdr:"unioncase=-1"`
}

//WRITE
type WRITE3args struct {
	File   nfs_fh3
	Offset offset3
	Count  count3
	Stable stable_how
	Data   []byte
}

type WRITE3resok struct {
	File_wcc  wcc_data
	Count     count3
	Committed stable_how
	Verf      createverf3
}

type WRITE3resfail struct {
	File_wcc wcc_data
}

type WRITE3res struct {
	Status  nfsstat3      `xdr:"union"`
	Resok   WRITE3resok   `xdr:"unioncase=0"`
	Resfail WRITE3resfail `xdr:"unioncase=-1"`
}

//CREAT
type createhow3 struct {
	Mode           createmode3 `xdr:"union"`
	Obj_attributes sattr3      `xdr:"unioncase=1"`
	Verf           createverf3 `xdr:"unioncase=2"`
}

type CREATE3args struct {
	Where diropargs3
	How   createhow3
}

type CREATE3resok struct {
	Obj            post_op_fh3
	Obj_attributes post_op_attr
	Dir_wcc        wcc_data
}

type CREATE3resfail struct {
	Dir_wcc wcc_data
}

type CREATE3res struct {
	Status  nfsstat3       `xdr:"union"`
	Resok   CREATE3resok   `xdr:"unioncase=0"`
	Resfail CREATE3resfail `xdr:"unioncase=-1"`
}

//MKDIR
type MKDIR3args struct {
	Where      diropargs3
	Attributes sattr3
}

type MKDIR3resok struct {
	Obj            post_op_fh3
	Obj_attributes post_op_attr
	Dir_wcc        wcc_data
}

type MKDIR3resfail struct {
	Dir_wcc wcc_data
}

type MKDIR3res struct {
	Status  nfsstat3      `xdr:"union"`
	Resok   MKDIR3resok   `xdr:"unioncase=0"`
	Resfail MKDIR3resfail `xdr:"unioncase=17"`
	//Resfail2 MKDIR3resfail `xdr:"unioncase=22"`
}

//SYMLINK
type symlinkdata3 struct {
	Symlink_attributes sattr3
	Symlink_data       nfspath3
}

type SYMLINK3args struct {
	Where   diropargs3
	Symlink symlinkdata3
}

type SYMLINK3resok struct {
	Obj            post_op_fh3
	Obj_attributes post_op_attr
	Dir_wcc        wcc_data
}

type SYMLINK3resfail struct {
	Dir_wcc wcc_data
}

type SYMLINK3res struct {
	Status  nfsstat3        `xdr:"union"`
	Resok   SYMLINK3resok   `xdr:"unioncase=0"`
	Resfail SYMLINK3resfail `xdr:"unioncase=-1"`
}

//MKNOD
type devicedata3 struct {
	Dev_attributes sattr3
	Spec           specdata3
}

type mknoddata3 struct {
	Type            ftype3      `xdr:"union"`
	Chr_device      devicedata3 `xdr:"unioncase=4"`
	Blk_device      devicedata3 `xdr:"unioncase=3"`
	Sock_attributes sattr3      `xdr:"unioncase=6"`
	Pipe_attributes sattr3      `xdr:"unioncase=7"`
}

type MKNOD3args struct {
	Where diropargs3
	What  mknoddata3
}

type MKNOD3resok struct {
	Obj            post_op_fh3
	Obj_attributes post_op_attr
	Dir_wcc        wcc_data
}

type MKNOD3resfail struct {
	Dir_wcc wcc_data
}

type MKNOD3res struct {
	Status  nfsstat3      `xdr:"union"`
	Resok   MKDIR3resok   `xdr:"unioncase=0"`
	Resfail MKDIR3resfail `xdr:"unioncase=-1"`
}

//REMOVE
type REMOVE3args struct {
	Object diropargs3
}

type REMOVE3resok struct {
	Dir_wcc wcc_data
}

type REMOVE3resfail struct {
	Dir_wcc wcc_data
}

type REMOVE3res struct {
	Status  nfsstat3       `xdr:"union"`
	Resok   REMOVE3resok   `xdr:"unioncase=0"`
	Resfail REMOVE3resfail `xdr:"unioncase=-1"`
}

//RMDIR
type RMDIR3args struct {
	Object diropargs3
}

type RMDIR3resok struct {
	Dir_wcc wcc_data
}

type RMDIR3resfail struct {
	Dir_wcc wcc_data
}

type RMDIR3res struct {
	Status  nfsstat3      `xdr:"union"`
	Resok   RMDIR3resok   `xdr:"unioncase=0"`
	Resfail RMDIR3resfail `xdr:"unioncase=-1"`
}

//RENAME
type RENAME3args struct {
	From diropargs3
	To   diropargs3
}

type RENAME3resok struct {
	Fromdir_wcc wcc_data
	Todir_wcc   wcc_data
}

type RENAME3resfail struct {
	Fromdir_wcc wcc_data
	Todir_wcc   wcc_data
}

type RENAME3res struct {
	Status  nfsstat3       `xdr:"union"`
	Resok   RENAME3resok   `xdr:"unioncase=0"`
	Resfail RENAME3resfail `xdr:"unioncase=-1"`
}

//LINK
type LINK3args struct {
	File nfs_fh3
	Link diropargs3
}

type LINK3resok struct {
	File_attributes post_op_attr
	Linkdir_wcc     wcc_data
}

type LINK3resfail struct {
	File_attributes post_op_attr
	Linkdir_wcc     wcc_data
}

type LINK3res struct {
	Status  nfsstat3     `xdr:"union"`
	Resok   LINK3resok   `xdr:"unioncase=0"`
	Resfail LINK3resfail `xdr:"unioncase=-1"`
}

//READDIR
type entry3 struct {
	Fileid    fileid3
	Name      filename3
	Cookie    cookie3
	Nextentry *entry3 `xdr:"optional"`
}

type dirlist3 struct {
	Entries *entry3 `xdr:"optional"`
	Eof     bool
}

type READDIR3args struct {
	Dir        nfs_fh3
	Cookie     cookie3
	Cookieverf cookieverf3
	Count      count3
}

type READDIR3resok struct {
	Dir_attributes post_op_attr
	Cookieverf     cookieverf3
	Reply          dirlist3
}

type READDIR3resfail struct {
	Dir_attributes post_op_attr
}

type READDIR3res struct {
	Status   nfsstat3        `xdr:"union"`
	Resok    READDIR3resok   `xdr:"unioncase=0"`
	Resfail  READDIR3resfail `xdr:"unioncase=20"`
	Resfail2 READDIR3resfail `xdr:"unioncase=10003"`
}

//READDIRPLUS
type READDIRPLUS3args struct {
	Dir        nfs_fh3
	Cookie     cookie3
	Cookieverf cookieverf3
	Dircount   uint
	Maxcount   uint
}

type post_op_fh3 struct {
	Handle_follows bool    `xdr:"union"`
	Handle         nfs_fh3 `xdr:"unioncase=1"`
}

type entryplus3 struct {
	Fileid          fileid3
	Name            filename3
	Cookie          cookie3
	Name_attributes post_op_attr
	Name_handle     post_op_fh3
	Nextentry       *entryplus3 `xdr:"optional"`
}

type dirlistplus3 struct {
	Entries *entryplus3 `xdr:"optional"`
	Eof     bool
}

type READDIRPLUS3resok struct {
	Dir_attributes post_op_attr
	Cookieverf     cookieverf3
	Reply          dirlistplus3
}

type READDIRPLUS3resfail struct {
	Dir_attributes post_op_attr
}

type READDIRPLUS3res struct {
	Status  nfsstat3            `xdr:"union"`
	Resok   READDIRPLUS3resok   `xdr:"unioncase=0"`
	Resfail READDIRPLUS3resfail `xdr:"unioncase=10004"`
}

//FSSTAT
type FSSTAT3args struct {
	Fsroot nfs_fh3
}

type FSSTAT3resok struct {
	Obj_attributes post_op_attr
	Tbytes         size3
	Fbytes         size3
	Abytes         size3
	Tfbytes        size3
	Ffbytes        size3
	Afbytes        size3
	Invarsec       uint32
}

type FSSTAT3resfail struct {
	Obj_attributes post_op_attr
}

type FSSTAT3res struct {
	Status  nfsstat3       `xdr:"union"`
	Resok   FSSTAT3resok   `xdr:"unioncase=0"`
	Resfail FSSTAT3resfail `xdr:"unioncase=-1"`
}

//FSINFO
type FSINFO3resfail struct {
	Obj_attributes post_op_attr
}

type FSINFO3resok struct {
	Obj_attributes post_op_attr
	Rtmax          uint32
	Rtpref         uint32
	Rtmult         uint32
	Wtmax          uint32
	Wtpref         uint32
	Wtmult         uint32
	Dtpref         uint32
	Maxfilesize64  uint64
	Time_delta     nfstime3
	Properties     uint32
}

type FSINFO3res struct {
	Status  nfsstat3       `xdr:"union"`
	Resok   FSINFO3resok   `xdr:"unioncase=0"`
	Resfail FSINFO3resfail `xdr:"unioncase=default"`
}

type FSINFO3args struct {
	Fsroot nfs_fh3
}

//PATHCONF
type PATHCONF3args struct {
	Object nfs_fh3
}

type PATHCONF3resok struct {
	Obj_attributes   post_op_attr
	Linkmax          uint32
	Namemax          uint32
	No_trunc         bool
	Chown_restricted bool
	Case_insensitive bool
	Case_preserving  bool
}

type PATHCONF3resfail struct {
	Obj_attributes post_op_attr
}

type PATHCONF3res struct {
	Status  nfsstat3         `xdr:"union"`
	Resok   PATHCONF3resok   `xdr:"unioncase=0"`
	Resfail PATHCONF3resfail `xdr:"unioncase=-1"`
}

//COMMIT
type COMMIT3args struct {
	File   nfs_fh3
	Offset offset3
	Count  count3
}

type COMMIT3resok struct {
	File_wcc wcc_data
	Verf     writeverf3
}

type COMMIT3resfail struct {
	File_wcc wcc_data
}

type COMMIT3res struct {
	Status  nfsstat3       `xdr:"union"`
	Resok   COMMIT3resok   `xdr:"unioncase=0"`
	Resfail COMMIT3resfail `xdr:"unioncase=-1"`
}
