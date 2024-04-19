package etcd

import (
	"context"
	"fmt"
	"go-learner/slice"
	"strconv"
	"sync/atomic"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	etcdV3 "go.etcd.io/etcd/client/v3"
)

type Service struct {
	state            atomic.Int64
	shardNumKey      string
	managerKeyPrefix string

	etcdClient *etcdV3.Client

	runCtx        context.Context
	cancelWatch   context.CancelFunc
	leaseId       etcdV3.LeaseID
	keepAliveChan <-chan *etcdV3.LeaseKeepAliveResponse
}

func (srv *Service) loadShardNum(ctx context.Context, retry int) (int64, error) {
	ctx1, cancel := context.WithTimeout(ctx, time.Second*3)
	rsp, err := srv.etcdClient.Get(ctx1, srv.shardNumKey)
	cancel()

	interval := time.Second
	for err != nil && retry > 0 {
		time.Sleep(interval)

		interval *= 2
		retry--

		ctx1, cancel := context.WithTimeout(ctx, time.Second*3)
		rsp, err = srv.etcdClient.Get(ctx1, srv.shardNumKey)
		cancel()
	}
	if err != nil {
		return 0, err
	}

	if len(rsp.Kvs) != 1 {
		return 0, fmt.Errorf("invalid shard size value")
	}

	sValue := slice.ByteSlice2String(rsp.Kvs[0].Value)
	_, err = strconv.ParseInt(sValue, 10, 64)
	if err != nil {
		return 0, err
	}

	return rsp.Header.Revision, nil
}

func (srv *Service) preemptShard(ctx context.Context, shardId int64, value string, ttl int64) (bool, error) {
	reqCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	rsp, err := srv.etcdClient.Grant(reqCtx, ttl)
	if err != nil {
		return false, err
	}
	leaseId := rsp.ID

	key := fmt.Sprintf("%s%d", srv.managerKeyPrefix, shardId)
	txn := srv.etcdClient.Txn(reqCtx).
		If(etcdV3.Compare(etcdV3.CreateRevision(key), "=", 0)).
		Then(etcdV3.OpPut(key, value, etcdV3.WithLease(leaseId))).
		Else(etcdV3.OpGet(key))
	txnRsp, err := txn.Commit()
	if err != nil {
		return false, err
	}

	if !txnRsp.Succeeded {
		return false, nil
	}

	ch, err := srv.etcdClient.KeepAlive(ctx, leaseId)
	if err != nil {
		return false, err
	}

	srv.leaseId = leaseId
	srv.keepAliveChan = ch

	return true, nil
}

func (srv *Service) managerLoop(shardNumRev, scaleRev, workerInfoRev int64) {
	ctx := srv.runCtx
	var watchCtx context.Context
	watchCtx, srv.cancelWatch = context.WithCancel(ctx)

	shardNumCh := srv.etcdClient.Watch(watchCtx, srv.shardNumKey,
		etcdV3.WithRev(shardNumRev),
	)

LOOP:
	for {
		select {
		case <-ctx.Done():
			return

		case _, ok := <-srv.keepAliveChan:
			if !ok {
				break LOOP
			}

		case wr, ok := <-shardNumCh:
			if !ok {
				shardNumCh = srv.etcdClient.Watch(watchCtx, srv.shardNumKey,
					etcdV3.WithRev(shardNumRev),
				)

				continue
			}

			revUpdated := GetLatestWatchRevision(&shardNumRev, &wr)
			if err := wr.Err(); err != nil {
				continue
			}

			if !revUpdated {
				continue
			}

			for _, ev := range wr.Events {
				if ev.Type == mvccpb.PUT {

				} else if ev.Type == mvccpb.DELETE {

				}
			}

		}
	}

	srv.cancelWatch()
}

func GetLatestWatchRevision(revision *int64, wr *etcdV3.WatchResponse) bool {
	var changed bool
	if wr.CompactRevision > *revision {
		*revision = wr.CompactRevision
		changed = true
	}

	if wr.Header.Revision > *revision {
		*revision = wr.Header.Revision
		changed = true
	}

	return changed
}
