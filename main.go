package main

import (
    "flag"
    "fmt"
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

    //开启RESTFull API Server
    serveHttpKVAPI(*kvport)
}

