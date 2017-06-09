package nfspro

import (
	"NfsAgent/conf"
	"NfsAgent/mlog"
	"NfsAgent/operate"
	"bytes"
	"fmt"
	"strings"

	"github.com/prashanthpai/sunrpc"
	"github.com/rasky/go-xdr/xdr2"
)

func InsertGroupnode(h, d *groupnode, p int) bool {
	if h.Groups == nil {
		h.Groups = d
		return true
	}
	i := 0
	n := h
	for n.Groups != nil {
		i++
		if i == p {
			if n.Groups.Groups == nil {
				n.Groups = d
				return true
			} else {
				d.Groups = n.Groups
				n.Groups = d.Groups
				return true
			}
		}
		n = n.Groups
		if n.Groups == nil {
			n.Groups = d
			return true
		}
	}
	return false
}

func InsertExportnode(h, d *exportnode, p int) bool {
	if h.Exports == nil {
		h.Exports = d
		return true
	}
	i := 0
	n := h
	for n.Exports != nil {
		i++
		if i == p {
			if n.Exports.Exports == nil {
				n.Exports = d
				return true
			} else {
				d.Exports = n.Exports
				n.Exports = d.Exports
				return true
			}
		}
		n = n.Exports
		if n.Exports == nil {
			n.Exports = d
			return true
		}
	}
	return false
}

type Mountd struct {
}

func (t *Mountd) Mountd3_NULL(cargs *sunrpc.CallArags, reply *Void) error {

	defer PinacRevove()

	mlog.Debug("Mountd3_NULL Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args Void

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Mountd3_NULL xdr Unmarshal args error :", err)
		return nil
	}

	mlog.Debug("Mountd3_NULL")

	return nil
}

func (t *Mountd) Mountd3_MNT(cargs *sunrpc.CallArags, reply *Mountres3) error {

	defer PinacRevove()

	mlog.Debug("Mountd3_MNT Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args Dirpath3Arg

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Mountd3_MNT xdr Unmarshal args error :", err)
		reply.Fhs_status = MNT3ERR_INVAL
		return nil
	}

	mlog.Debug("Mountd3_MNT args :", fmt.Sprintf("%+v", args))

	var argspath string
	var userpwd string

	tmp := strings.SplitN(string(args.Dirpath), "/", 2)

	if len(tmp) < 2 {
		mlog.Error("Mountd3_MNT args error")
		reply.Fhs_status = MNT3ERR_NOTDIR
		return nil
	}

	mlog.Debug("Mountd3_MNT tmp :", tmp)
	argspath = tmp[1]
	userpwd = tmp[0]

	mlog.Debug("Mountd3_MNT argspath:", argspath, "userpwd:", userpwd)

	path := "/" + argspath

	var pwd string
	var user string

	if userpwd != "" {

		up_tmp := strings.SplitN(userpwd, ":", 2)

		if len(up_tmp) < 2 {
			mlog.Error("Mountd3_MNT userpwd args error")
			reply.Fhs_status = MNT3ERR_NOTDIR
			return nil
		}

		user_tmp := up_tmp[0]

		pwd = up_tmp[1]

		user = user_tmp[1:]

	}

	mlog.Debug("Mountd3_MNT pwd:", pwd, "user:", user)
	/*

		ip_port_str := cargs.Auth_unix.Addr.String()

		ip_port := strings.Split(ip_port_str, ":")

		ip := ip_port[0]

	*/
	/*

		flag := conf.IsExport(path, ip)

		if flag == false {

			mlog.Error("Mountd3_MNT Export error path :", path, " ip:", ip)

			reply.Fhs_status = MNT3ERR_ACCES

			return nil
		}

	*/
	fh, err := operate.Resc.Mnt(path, user)

	if err != nil {
		mlog.Error("Mountd3_MNT operate.Resc.Mnt error :", err, " args:", string(args.Dirpath))
		reply.Fhs_status = MNT3ERR_INVAL
		return nil
	}

	reply.Fhs_status = MNT3_OK
	reply.Mountinfo.Auth_floavors = append(reply.Mountinfo.Auth_floavors, 1)

	//reply.Mountinfo.Fhandle = make([]byte, 4)
	//binary.BigEndian.PutUint32(reply.Mountinfo.Fhandle, uint32(fh))

	mlog.Debug("Mountd3_MNT fh :", fh)

	reply.Mountinfo.Fhandle, err = operate.LocalFhToByte(fh)

	if err != nil {
		mlog.Error("Mountd3_MNT  operate.LocalFhToByte error :", err)
		reply.Fhs_status = MNT3ERR_INVAL
		return nil
	}

	mlog.Debug("Mountd3_MNT reply :", fmt.Sprintf("%+v", reply))

	return nil
}

func (t *Mountd) Mountd3_DUMP(cargs *sunrpc.CallArags, reply *Mountlist) error {

	defer PinacRevove()

	mlog.Debug("Mountd3_DUMP Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args Void

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Mountd3_DUMP xdr Unmarshal args error :", err)
		return nil
	}

	mlog.Debug("Mountd3_DUMP")

	return nil
}

func (t *Mountd) Mountd3_UMNT(cargs *sunrpc.CallArags, reply *Void) error {

	defer PinacRevove()

	mlog.Debug("Mountd3_UMNT Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args Dirpath3Arg

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Mountd3_UMNT xdr Unmarshal args error :", err)
		return nil
	}

	/*	err = operate.Resc.Unmnt(string(args.Dirpath))

		if err != nil {
			mlog.Error("Mountd3_UMNT operate.Resc.Unmnt error :", err)
			return nil
		}
	*/
	mlog.Debug("Mountd3_UMNT")
	return nil

}

func (t *Mountd) Mountd3_EXPORT(cargs *sunrpc.CallArags, reply *Exports) error {

	defer PinacRevove()

	mlog.Debug("Mountd3_EXPORT Begin ", fmt.Sprintf("%+v", cargs.Auth_unix))

	var args Void

	args_buff := bytes.NewBuffer(cargs.Arags)

	_, err := xdr.Unmarshal(args_buff, &args)

	if err != nil {
		mlog.Error("Mountd3_EXPORT xdr Unmarshal args error :", err)
		return nil
	}

	if len(conf.Exports) <= 0 {
		mlog.Debug("Mountd3_EXPORT conf exports empty ")
		return nil
	}

	var head exportnode
	var ghead groupnode

	ghead.Name = name(conf.Exports[0].Info[0].GroupName)

	head.Ex_dir = dirpath(conf.Exports[0].ExportPath)
	head.Groups = &ghead

	for i := 1; i < len(conf.Exports[0].Info); i++ {
		var gp groupnode
		gp.Name = name(conf.Exports[0].Info[i].GroupName)

		InsertGroupnode(&ghead, &gp, i)
	}

	for i := 1; i < len(conf.Exports); i++ {

		var en exportnode
		en.Ex_dir = dirpath(conf.Exports[i].ExportPath)

		var gph groupnode
		gph.Name = name(conf.Exports[i].Info[0].GroupName)

		for k := 1; k < len(conf.Exports[i].Info); k++ {
			var gp groupnode
			gp.Name = name(conf.Exports[i].Info[k].GroupName)

			InsertGroupnode(&gph, &gp, k)
		}

		en.Groups = &gph
	}

	reply.Exports = &head

	return nil
}
