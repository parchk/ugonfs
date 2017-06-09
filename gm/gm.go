package gm

import (
	"NfsAgent/conf"
	"NfsAgent/mlog"
	"net/http"
)

type gm_handle struct {
}

func (h *gm_handle) LoadExports(w http.ResponseWriter, req *http.Request) {
	conf.LoadExports()
}

func init() {
	handle := gm_handle{}

	mux := http.NewServeMux()

	mux.HandleFunc("/load_exports", handle.LoadExports)

	go func() {

		err := http.ListenAndServe("localhost:9922", mux)

		if err != nil {
			mlog.Error("gm server ListenAndServe error :", err)
		}

	}()
}
