package raft

import "github.com/whosefriendA/firEtcd/pkg/firlog"

// Debugging
const Debug = true

func DPrintf(format string, a ...interface{}) {
	if Debug {
		firlog.Logger.Infof(format, a...)
	}
}
