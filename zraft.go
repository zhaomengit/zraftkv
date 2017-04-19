package main

import (
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/snap"
    "github.com/coreos/etcd/raft"
    "github.com/coreos/etcd/wal"
    "github.com/coreos/etcd/rafthttp"
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

    snapshotter      *snap.Snapshotter
    snapshotterReady chan *snap.Snapshotter // snapshotter准备好的信号

    snapCount uint64
    transport *rafthttp.Transport
    stopc     chan struct{} // 提交信道关闭的信号
    httpstopc chan struct{} // 关闭http server的信号
    httpdonec chan struct{} // http server 关闭完成的信号
}

// 创建raft节点
// id : 节点ID
// peers : raft服务 集群url地址
// getSnapshot: 获取kvstore中Snapshot的函数
//
func newRaftNode(id int,
	peers []string,
	join bool,
	getSnapshot func() ([]byte, error),
	proposeC chan<- string,
	confChangeC chan<- raftpb.ConfChange,
) (<-chan *string, <-chan error, <-chan *snap.Snapshotter) {
	commitC := make(chan *string)
	errorC := make(chan error)
	snapshotterReady := make(chan *snap.Snapshotter, 1)
	return commitC, errorC, snapshotterReady
}
