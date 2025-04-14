package raft

import (
	"github.com/whosefriendA/firEtcd/pb"
	"github.com/whosefriendA/firEtcd/src/config"
	"github.com/whosefriendA/firEtcd/src/firlog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RaftEnd struct {
	conf laneConfig.RaftEnd
	conn pb.RaftClient
}

func NewRaftClient(conf laneConfig.RaftEnd) *RaftEnd {
	conn, err := grpc.NewClient(conf.Addr+conf.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		laneLog.Logger.Infoln("Dail faild ", err.Error())
		return nil
	}
	client := pb.NewRaftClient(conn)
	ret := &RaftEnd{
		conn: client,
		conf: conf,
	}
	return ret
}
