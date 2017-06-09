package server

import (
	"NfsAgent/conf"
	"NfsAgent/mlog"
	"NfsAgent/nfspro"
	"io"
	"net"
	"net/rpc"
	"os"
	"time"

	"github.com/prashanthpai/sunrpc"
)

const (
	programNumber  = uint32(100005)
	programVersion = uint32(3)

	programNumber_Nfs  = uint32(100003)
	programVersion_Nfs = uint32(3)

	programNumber_Nfs_ACL  = uint32(100227)
	programVersion_Nfs_ACL = uint32(3)
)

func init() {

	Instance = NewAgentSvr()
	err := Instance.reigisterNfs()
	err = Instance.reigisterMound()
	err = Instance.reigisterNfsacl()
	err = Instance.setMountPort()
	err = Instance.setNfsPort()
	err = Instance.setNfsaclPort()
	//err = Instance.setTestPort()

	if err != nil {
		mlog.Error("init server error")
		os.Exit(1)
	}
}

func Run() {
	Instance.Run()
}

func NewAgentSvr() *AgentSvr {

	svr := AgentSvr{
		rpcsvr:      rpc.NewServer(),
		notifyClose: make(chan io.ReadWriteCloser, 5),
	}

	return &svr
}

var Instance *AgentSvr

type AgentSvr struct {
	rpcsvr      *rpc.Server
	notifyClose chan io.ReadWriteCloser
}

func (c *AgentSvr) Run() {

	listener, err := net.Listen("tcp", conf.AppConf.Host)
	if err != nil {
		mlog.Error("AgentSvr net.Listen() failed: ", err)
	}

	go func() {
		for rwc := range c.notifyClose {
			conn := rwc.(net.Conn)
			mlog.Debug("Client %s disconnected", conn.RemoteAddr().String())
		}
	}()

	for {
		conn, err := listener.Accept()

		if err != nil {
			mlog.Error("listener.Accept() failed: ", err)
			time.Sleep(time.Second * 10)
			continue
		} else {
			mlog.Debug("Client %s connected", conn.RemoteAddr().String())
		}

		go c.rpcsvr.ServeCodec(sunrpc.NewServerCodec(conn, c.notifyClose))
	}

}

func (c *AgentSvr) setTestPort() error {
	ret, err := sunrpc.PmapUnset(12345, 1)
	if err != nil {
		mlog.Error("AgentSvr setNfsaclPort unset error:", err)
		return err
	}

	if ret != true {
		mlog.Warning("AgentSvr setNfsaclPort unset faile")
	}

	ret, err = sunrpc.PmapSet(12345, 1, sunrpc.IPProtoTCP, uint32(2049))
	if err != nil {
		mlog.Error("AgentSvr setNfsaclPort set error:", err)
		return err
	}

	if ret != true {
		mlog.Warning("AgentSvr setNfsaclPort set faile")
	}

	return nil
}

func (c *AgentSvr) setNfsaclPort() error {

	ret, err := sunrpc.PmapUnset(programNumber_Nfs_ACL, programVersion_Nfs_ACL)
	if err != nil {
		mlog.Error("AgentSvr setNfsaclPort unset error:", err)
		return err
	}

	if ret != true {
		mlog.Warning("AgentSvr setNfsaclPort unset faile")
	}

	ret, err = sunrpc.PmapSet(programNumber_Nfs_ACL, programVersion_Nfs_ACL, sunrpc.IPProtoTCP, uint32(conf.AppConf.NfsaclPort))
	if err != nil {
		mlog.Error("AgentSvr setNfsaclPort set error:", err)
		return err
	}

	if ret != true {
		mlog.Warning("AgentSvr setNfsaclPort set faile")
	}

	return nil
}

func (c *AgentSvr) setNfsPort() error {

	ret, err := sunrpc.PmapUnset(programNumber_Nfs, programVersion_Nfs)
	if err != nil {
		mlog.Error("AgentSvr setNfsPort unset error:", err)
		return err
	}

	if ret != true {
		mlog.Warning("AgentSvr setNfsPort unset faile")
	}

	ret, err = sunrpc.PmapSet(programNumber_Nfs, programVersion_Nfs, sunrpc.IPProtoTCP, uint32(conf.AppConf.NfsPort))
	if err != nil {
		mlog.Error("AgentSvr setNfsPort set error:", err)
		return err
	}

	if ret != true {
		mlog.Warning("AgentSvr setNfsPort set faile")
	}

	return nil
}

func (c *AgentSvr) setMountPort() error {

	ret, err := sunrpc.PmapUnset(programNumber, programVersion)
	if err != nil {
		mlog.Error("AgentSvr setMountPort unset error:", err)
		return err
	}

	if ret != true {
		mlog.Warning("AgentSvr setMountPort unset faile")
	}

	ret, err = sunrpc.PmapSet(programNumber, programVersion, sunrpc.IPProtoTCP, uint32(conf.AppConf.MountPort))
	if err != nil {
		mlog.Error("AgentSvr setMountPort set error:", err)
		return err
	}

	if ret != true {
		mlog.Warning("AgentSvr setMountPort set faile")
	}

	return nil
}

func (c *AgentSvr) reigisterNfs() error {
	nfs := new(nfspro.Nfs)

	err := c.rpcsvr.Register(nfs)

	////Nfs
	{
		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(0)},
			Name: "Nfs.Nfs3_NULL"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(1)},
			Name: "Nfs.Nfs3_GETATTR"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(2)},
			Name: "Nfs.Nfs3_SETATTR"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(3)},
			Name: "Nfs.Nfs3_LOOKUP"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(4)},
			Name: "Nfs.Nfs3_ACCESS"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(5)},
			Name: "Nfs.Nfs3_READLINK"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(6)},
			Name: "Nfs.Nfs3_READ"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(7)},
			Name: "Nfs.Nfs3_WRITE"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(8)},
			Name: "Nfs.Nfs3_CREATE"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(9)},
			Name: "Nfs.Nfs3_MKDIR"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(10)},
			Name: "Nfs.Nfs3_SYMLINK"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(11)},
			Name: "Nfs.Nfs3_MKNODE"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(12)},
			Name: "Nfs.Nfs3_REMOVE"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(13)},
			Name: "Nfs.Nfs3_RMDIR"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(14)},
			Name: "Nfs.Nfs3_RENAME"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(15)},
			Name: "Nfs.Nfs3_LINK"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(16)},
			Name: "Nfs.Nfs3_READDIR"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(17)},
			Name: "Nfs.Nfs3_READDIRPLUS"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(18)},
			Name: "Nfs.Nfs3_FSSTAT"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(19)},
			Name: "Nfs.Nfs3_FSINFO"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(20)},
			Name: "Nfs.Nfs3_PATHCONF"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs, programVersion_Nfs, uint32(21)},
			Name: "Nfs.Nfs3_COMMIT"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfs error :", err)
			return err
		}

	}

	return nil
}

func (c *AgentSvr) reigisterMound() error {

	mount := new(nfspro.Mountd)

	err := c.rpcsvr.Register(mount)
	////mount
	{

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber, programVersion, uint32(0)},
			Name: "Mountd.Mountd3_NULL"})

		if err != nil {
			mlog.Error("AgentSvr reigisterMound error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber, programVersion, uint32(1)},
			Name: "Mountd.Mountd3_MNT"})

		if err != nil {
			mlog.Error("AgentSvr reigisterMound error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber, programVersion, uint32(3)},
			Name: "Mountd.Mountd3_UMNT"})

		if err != nil {
			mlog.Error("AgentSvr reigisterMound error :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber, programVersion, uint32(5)},
			Name: "Mountd.Mountd3_EXPORT"})

		if err != nil {
			mlog.Error("AgentSvr reigisterMound err :", err)
			return err
		}
	}

	return nil
}

func (c *AgentSvr) reigisterNfsacl() error {
	nfsacl := new(nfspro.Nfsacl)

	err := c.rpcsvr.Register(nfsacl)

	///nfsacl
	{

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs_ACL, programVersion_Nfs_ACL, uint32(0)},
			Name: "Nfsacl.Nfsacl3_NULL"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfsacl err :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs_ACL, programVersion_Nfs_ACL, uint32(1)},
			Name: "Nfsacl.Nfsacl3_GETACL"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfsacl err :", err)
			return err
		}

		err = sunrpc.RegisterProcedure(sunrpc.Procedure{
			ID:   sunrpc.ProcedureID{programNumber_Nfs_ACL, programVersion_Nfs_ACL, uint32(2)},
			Name: "Nfsacl.Nfsacl3_SETACL"})

		if err != nil {
			mlog.Error("AgentSvr reigisterNfsacl err :", err)
			return err
		}

	}

	return nil
}
