package raft

import (
	"github.com/whosefriendA/firEtcd/proto/pb"
	"github.com/whosefriendA/firEtcd/src/pkg/firconfig"
	"github.com/whosefriendA/firEtcd/src/pkg/firlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RaftEnd struct {
	conf firconfig.RaftEnd
	conn pb.RaftClient
}

func NewRaftClient(conf firconfig.RaftEnd) *RaftEnd {
	conn, err := grpc.NewClient(conf.Addr+conf.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		firlog.Logger.Infoln("Dail faild ", err.Error())
		return nil
	}
	client := pb.NewRaftClient(conn)
	ret := &RaftEnd{
		conn: client,
		conf: conf,
	}
	return ret
}
