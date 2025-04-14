package kvraft

import (
	"github.com/whosefriendA/firEtcd/proto/pb"
	"github.com/whosefriendA/firEtcd/src/firlog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type KVClient struct {
	Valid    bool
	Conn     pb.KvserverClient
	Realconn *grpc.ClientConn
}

func NewKvClient(addr string) *KVClient {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		firlog.Logger.Infoln("Dail faild ", err.Error())
		return nil
	}
	client := pb.NewKvserverClient(conn)

	ret := &KVClient{
		Valid: true,
		Conn:  client,
	}
	return ret
}
