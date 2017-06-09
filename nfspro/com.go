package nfspro

import (
	"NfsAgent/mlog"
	"runtime/debug"
)

func PinacRevove() {
	if err := recover(); err != nil {
		mlog.Error(err)
		mlog.Error(string(debug.Stack()))
	}
}
