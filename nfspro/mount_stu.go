package nfspro

const MNTPATHLEN = 1024 /* Maximum bytes in a path name */
const MNTNAMLEN = 255   /* Maximum bytes in a name */
const FHSIZE3 = 64      /* Maximum bytes in a V3 file handle */

type fhandle3 []byte
type dirpath string
type name string

type Void struct {
}

type mountstat3 int32

const (
	MNT3_OK mountstat3 = iota
	MNT3ERR_PERM
	MNT3ERR_NOENT
	MNT3ERR_IO
	MNT3ERR_ACCES
	MNT3ERR_NOTDIR
	MNT3ERR_INVAL
	MNT3ERR_NAMETOOLONG
	MNT3ERR_NOTSUPP
	MNT3ERR_SERVERFAULT
)

type Dirpath3Arg struct {
	Dirpath dirpath
}

//MNT
type mountres3_ok struct {
	Fhandle       fhandle3
	Auth_floavors []int
}

type Mountres3 struct {
	Fhs_status mountstat3   `xdr:"union"`
	Mountinfo  mountres3_ok `xdr:"unioncase=0"`
}

//DUMP
type Mountlist struct {
	Mountbody *mountbody `xdr:"optional"`
}

type mountbody struct {
	Ml_hostname  name
	Ml_directory dirpath
	Ml_next      Mountlist
}

//EXPORT

type groupnode struct {
	Name   name
	Groups *groupnode `xdr:"optional"`
}

type exportnode struct {
	Ex_dir  dirpath
	Groups  *groupnode  `xdr:"optional"`
	Exports *exportnode `xdr:"optional"`
}

type Groups struct {
	Groups *groupnode `xdr:"optional"`
}

type Exports struct {
	Exports *exportnode `xdr:"optional"`
}
