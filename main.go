package main

import (
    "flag"
    "fmt"
    "github.com/coreos/etcd/raft/raftpb"
    "strings"
)

func main()  {
    // 集群主机raft服务地址
    cluster := flag.String("cluster", "http://127.0.0.1:9091", "集群raft服务地址,多个使用 , 隔开")
    // 集群主机id编号
    id := flag.Int("id", 1, "节点ID")
    // RESTFul API 访问地址
    kvport := flag.Int("port", 9191, "RESTFul API访问端口")
    // 是否是加入集群
    join := flag.Bool("join", false, "是否加入一个新的集群")
    flag.Parse()
    fmt.Print(*cluster, *id, *kvport, *join)

    // 创建发送channel
    proposeC := make(chan string)
    defer close(proposeC)

    // 创建ConfChange channel
    confChangeC := make(chan raftpb.ConfChange)
    defer close(confChangeC)

    var kvs *kvstore

    // 包装一下函数,包外访问不到
    getSnapshot := func() ([]byte, error) { return kvs.getSnapshot() }

    // 创建raft节点
    commitC, errorC, snapshotterReady := newRaftNode(*id, strings.Split(*cluster, ","), *join, getSnapshot, proposeC, confChangeC)

    kvs = newKVStore(<-snapshotterReady, proposeC, commitC, errorC)

    // 启动动http api服务器,处理发送到的raft请求
    serveHttpKVAPI(kvs, *kvport, confChangeC, errorC)
}

