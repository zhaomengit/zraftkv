# zraftkv

这是一个使用[etcd]的raft和httraft库实现的分布式key-value存储系统,提供简单的RESTful API来进行key-value存储,使用了[Raft]分布式一致性算法

[etcd]: https://github.com/coreos/etcd
[raft]: http://raftconsensus.github.io/


## 使用

### 参数

```
Usage of ./main:
  -cluster string
        集群raft服务地址,多个使用 , 隔开 (default "http://127.0.0.1:9091")
  -id int
        节点ID (default 1)
  -join
        是否加入一个新的集群
  -port int
        RESTFul API访问端口 (default 9191)
```



