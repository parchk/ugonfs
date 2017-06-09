package nfspro

//GETACL
type nfsacl_type int32

const (
	NFSACL_TYPE_USER_OBJ          nfsacl_type = 0x0001
	NFSACL_TYPE_USER                          = 0x0002
	NFSACL_TYPE_GROUP_OBJ                     = 0x0004
	NFSACL_TYPE_GROUP                         = 0x0008
	NFSACL_TYPE_CLASS_OBJ                     = 0x0010
	NFSACL_TYPE_CLASS                         = 0x0020
	NFSACL_TYPE_DEFAULT                       = 0x1000
	NFSACL_TYPE_DEFAULT_USER_OBJ              = 0x1001
	NFSACL_TYPE_DEFAULT_USER                  = 0x1002
	NFSACL_TYPE_DEFAULT_GROUP_OBJ             = 0x1004
	NFSACL_TYPE_DEFAULT_GROUP                 = 0x1008
	NFSACL_TYPE_DEFAULT_CLASS_OBJ             = 0x1010
	NFSACL_TYPE_DEFAULT_OTHER_OBJ             = 0x1020
)

type nfsacl_ace struct {
	Type nfsacl_type
	Id   uint
	Perm uint
}

type GETACL3resok struct {
	Attr              post_op_attr
	Mask              uint
	Ace_count         uint
	Ace               []nfsacl_type
	Default_ace_count uint
	Default_ace       []nfsacl_type
}

type GETACL3res struct {
	Status nfsstat3     `xdr:"union"`
	Resok  GETACL3resok `xdr:"unioncase=0"`
}

type GETACL3args struct {
	Dir  nfs_fh3
	Mask uint
}

//SETACL
type SETACL3args struct {
	Dir               nfs_fh3
	Mask              uint32
	Ace_count         uint32
	Ace               []nfsacl_ace
	Default_ace_count uint32
	Default_ace       []nfsacl_ace
}

type SETACL3resok struct {
	Attr post_op_attr
}

type SETACL3res struct {
	Status nfsstat3     `xdr:"union"`
	Resok  SETACL3resok `xdr:"unioncase=0"`
}
