package main

import (
	"fmt"
    "log"
    "os"
    "strconv"

	"github.com/coreos/etcd/etcdserver/stats"
	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/rafthttp"
	"github.com/coreos/etcd/snap"
	"github.com/coreos/etcd/wal"
    "golang.org/x/net/context"
)

//  用户raft后端存储的 key-value 结构
type raftNode struct {
	proposeC    <-chan string            // 提交 messages (k,v) 信道
	confChangeC <-chan raftpb.ConfChange // 提交集群 config changes 信道
	commitC     chan<- *string           // 提交到日志信道
	errorC      chan<- error             // 出错信道

	id          int      // 用于raft回话的client id
	peers       []string // raft 服务url
	join        bool     // 节点时候要加入一个集群中
	waldir      string   // WAL目录路径
	snapdir     string   // snapshot路径
	getSnapshot func() ([]byte, error)
	lastIndex   uint64 // 开始日志的索引

	confState     raftpb.ConfState
	snapshotIndex uint64
	appliedIndex  uint64

	node        raft.Node
	raftStorage *raft.MemoryStorage
	wal         *wal.WAL

	snapshotter      *snap.Snapshotter       // 存储node节点的状态, 只有一个属性: 存放节点状态的目录
	snapshotterReady chan *snap.Snapshotter  // snapshotter准备好的信号

	snapCount uint64
	transport *rafthttp.Transport
	stopc     chan struct{} // 提交信道关闭的信号
	httpstopc chan struct{} // 关闭http server的信号
	httpdonec chan struct{} // http server 关闭完成的信号
}

var defaultSnapCount uint64 = 10000

// 启动raft
func (rc *raftNode) startRaft() {

	// snapdir不存在创建
	if !fileutil.Exist(rc.snapdir) {
		if err := os.Mkdir(rc.snapdir, 0750); err != nil {
			log.Fatalf("raftexample: cannot create dir for snapshot (%v)", err)
		}
	}
	rc.snapshotter = snap.New(rc.snapdir)
	rc.snapshotterReady <- rc.snapshotter

	oldwal := wal.Exist(rc.waldir)
	rc.wal = rc.replayWAL()

	// raft.Peer 包含两个字段, ID: 节点ID Context 数据
	rpeers := make([]raft.Peer, len(rc.peers))
	for i := range rpeers {
		rpeers[i] = raft.Peer{ID: uint64(i + 1)}
	}
	c := &raft.Config{
		ID:              uint64(rc.id),
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         rc.raftStorage,
		MaxSizePerMsg:   1024 * 1024,
		MaxInflightMsgs: 256,
	}

	if oldwal {
		rc.node = raft.RestartNode(c)
	} else {
		startPeers := rpeers
		if rc.join {
			startPeers = nil
		}
		rc.node = raft.StartNode(c, startPeers)
	}

	ss := &stats.ServerStats{}
	ss.Initialize()

	rc.transport = &rafthttp.Transport{
		ID:          types.ID(rc.id),
		ClusterID:   0x1000,
		Raft:        rc,
		ServerStats: ss,
		LeaderStats: stats.NewLeaderStats(strconv.Itoa(rc.id)),
		ErrorC:      make(chan error),
	}

	rc.transport.Start()
	for i := range rc.peers {
		if i+1 != rc.id {
			rc.transport.AddPeer(types.ID(i+1), []string{rc.peers[i]})
		}
	}

	go rc.serveRaft()
	go rc.serveChannels()
}

func (rc *raftNode) serveRaft()  {

}

func (rc *raftNode) serveChannels()  {

}

func (rc *raftNode) replayWAL()  *wal.WAL {
	return nil
}

func (rc *raftNode) Process(ctx context.Context, m raftpb.Message) error {
    return rc.node.Step(ctx, m)
}
func (rc *raftNode) IsIDRemoved(id uint64) bool                           { return false }
func (rc *raftNode) ReportUnreachable(id uint64)                          {}
func (rc *raftNode) ReportSnapshot(id uint64, status raft.SnapshotStatus) {}


// 创建raft节点
// newRaftNode初始化一个raft实例,返回一个日志记录channel和错误channel,
// 日志更新通过proposeC,日志回播通过日志记录channel,后面跟随一个nil信息
// 停止的话关闭proposeC和errorC
// id : 节点ID
// peers : raft服务 集群url地址
// getSnapshot: 获取kvstore中Snapshot的函数
//
func newRaftNode(id int,
	peers []string,
	join bool,
	getSnapshot func() ([]byte, error),
	proposeC <-chan string,
	confChangeC <-chan raftpb.ConfChange,
) (<-chan *string, <-chan error, <-chan *snap.Snapshotter) {

	commitC := make(chan *string)
	errorC := make(chan error)

	rc := &raftNode{
		proposeC:    proposeC,
		confChangeC: confChangeC,
		commitC:     commitC,
		errorC:      errorC,
		id:          id,
		peers:       peers,
		join:        join,
		waldir:      fmt.Sprintf("zraftexample-%d", id),
		snapdir:     fmt.Sprintf("zraftexample-%d-snap", id),
		getSnapshot: getSnapshot,
		snapCount:   defaultSnapCount,
		stopc:       make(chan struct{}),
		httpstopc:   make(chan struct{}),
		httpdonec:   make(chan struct{}),

		snapshotterReady: make(chan *snap.Snapshotter, 1),
		// rest of structure populated after WAL replay
	}
	go rc.startRaft()
	return commitC, errorC, rc.snapshotterReady
}
