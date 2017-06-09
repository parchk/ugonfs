package nfspro

import (
	"NfsAgent/mlog"
	"NfsAgent/operate"
	"bytes"
	//	"encoding/binary"
	"fmt"

	"github.com/prashanthpai/sunrpc"
	"github.com/rasky/go-xdr/xdr2"
)

type Nfsacl struct {
}

func (c *Nfsacl) Nfsacl3_NULL(cargs *sunrpc.CallArags, reply *Void) error {

	defer PinacRevove()

	mlog.Debug("Nfsacl3_NULL Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args Void

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfsacl3_NULL xdr Unmarshal args error :", err)
		return nil
	}

	mlog.Debug("Nfsacl3_NULL")
	return nil
}

func (c *Nfsacl) Nfsacl3_GETACL(cargs *sunrpc.CallArags, reply *GETATTR3res) error {

	defer PinacRevove()

	mlog.Debug("Nfsacl3_GETACL Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args GETACL3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfsacl3_GETACL xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfsacl3_GETACL args :", args)

	if len(args.Dir.Data) == 0 || len(args.Dir.Data) < 4 {
		mlog.Warning("Nfsacl3_GETACL args object data error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	reply.Status = NFS3ERR_NOTSUPP
	/*
		stat, err := operate.Resc.Stat(operate.Fh(fh))

		if err != nil {
			mlog.Error("Nfsacl3_GETACL operate.Resc.Stat err :", err)
			reply.Status = NFS3ERR_INVAL
			return nil
		}

		attr := GetPosAttr(stat)

		reply.Resok.Obj_attributes = attr.Attributes
	*/
	return nil
}

func (c *Nfsacl) Nfsacl3_SETACL(cargs *sunrpc.CallArags, reply *SETACL3res) error {

	defer PinacRevove()

	mlog.Error("Nfsacl3_SETACL Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args SETACL3args

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Nfsacl3_SETACL xdr Unmarshal args error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	mlog.Debug("Nfsacl3_SETACL args :", fmt.Sprintf("%+v", args))

	if len(args.Dir.Data) == 0 || len(args.Dir.Data) < 4 {
		mlog.Warning("Nfsacl3_SETACL args object data error")
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	//fh := binary.BigEndian.Uint32(args.Dir.Data)
	fh, err := operate.LocalGetFh(args.Dir.Data)

	if err != nil {
		mlog.Error("Nfsacl3_SETACL GetFh error :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	stat, err := operate.Resc.Stat(operate.Fh(fh))

	if err != nil {
		mlog.Error("Nfsacl3_SETACL operate.Resc.Stat err :", err)
		reply.Status = NFS3ERR_INVAL
		return nil
	}

	attr := GetPosAttr(stat)

	reply.Resok.Attr = attr

	return nil
}
