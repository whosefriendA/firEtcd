package kvraft

import (
	"context"
	"time"

	"github.com/whosefriendA/firEtcd/proto/pb"
)

// LeaseService implements pb.LeaseServer using KVServer's lease manager
type LeaseService struct {
	kv *KVServer
	pb.UnimplementedLeaseServer
}

func NewLeaseService(kv *KVServer) *LeaseService {
	return &LeaseService{kv: kv}
}

func (s *LeaseService) LeaseGrant(ctx context.Context, req *pb.LeaseGrantRequest) (*pb.LeaseGrantResponse, error) {
	ttl := time.Duration(req.GetTTLMs()) * time.Millisecond
	l, err := s.kv.leaseMgr.Grant(ttl)
	if err != nil {
		return nil, err
	}
	return &pb.LeaseGrantResponse{ID: l.ID, TTLMs: int64(l.TTL / time.Millisecond)}, nil
}

func (s *LeaseService) LeaseRevoke(ctx context.Context, req *pb.LeaseRevokeRequest) (*pb.LeaseRevokeResponse, error) {
	s.kv.proposeLeaseRevoke(req.GetID())
	return &pb.LeaseRevokeResponse{}, nil
}

func (s *LeaseService) LeaseKeepAlive(stream pb.Lease_LeaseKeepAliveServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		l, err := s.kv.leaseMgr.KeepAlive(req.GetID())
		if err != nil || l == nil {
			if sendErr := stream.Send(&pb.LeaseKeepAliveResponse{ID: req.GetID(), TTLMs: 0}); sendErr != nil {
				return sendErr
			}
			continue
		}
		remaining := l.ExpiresAt.Sub(time.Now())
		if remaining < 0 {
			remaining = 0
		}
		if err := stream.Send(&pb.LeaseKeepAliveResponse{ID: l.ID, TTLMs: int64(remaining / time.Millisecond)}); err != nil {
			return err
		}
	}
}

func (s *LeaseService) LeaseTimeToLive(ctx context.Context, req *pb.LeaseTimeToLiveRequest) (*pb.LeaseTimeToLiveResponse, error) {
	l := s.kv.leaseMgr.GetLease(req.GetID())
	if l == nil {
		return &pb.LeaseTimeToLiveResponse{ID: req.GetID(), TTLMs: 0}, nil
	}
	remaining := l.ExpiresAt.Sub(time.Now())
	if remaining < 0 {
		remaining = 0
	}
	resp := &pb.LeaseTimeToLiveResponse{ID: l.ID, TTLMs: int64(remaining / time.Millisecond)}
	if req.GetKeys() {
		resp.Keys = s.kv.leaseMgr.GetLeaseKeys(l.ID)
	}
	return resp, nil
}
