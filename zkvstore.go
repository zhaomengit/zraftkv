// 用于zkv后端存储

package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/coreos/etcd/snap"
	"log"
	"sync"
)

type kvstore struct {
	kvStore     map[string]string // 键值存储
	mu          sync.RWMutex      // 同步,读写互斥量
	proposeC    chan<- string     // channel 提交更新
	snapshotter *snap.Snapshotter // 序列化
}

type kv struct {
	Key   string
	Value string
}

// 创建kvstore函数
func newKVStore(snapshotter *snap.Snapshotter, proposeC chan<- string, commitC <-chan *string, errorC <-chan error) *kvstore {
    s := &kvstore{proposeC: proposeC, kvStore: make(map[string]string), snapshotter: snapshotter}
    // replay log into key-value map
    s.readCommits(commitC, errorC)
    // 从raft中读取提交信息,存储到kvStore map, 出错退出
    go s.readCommits(commitC, errorC)
    return s
}

//
func (k *kvstore) readCommits(commitC <-chan *string, errorC <-chan error) {
    // 循环遍历commitC
    for data := range commitC {
        if data == nil {
            // done replaying log; new data incoming
            // OR signaled to load snapshot
            snapshot, err := k.snapshotter.Load()
            if err == snap.ErrNoSnapshot {
                return
            }
            if err != nil && err != snap.ErrNoSnapshot {
                log.Panic(err)
            }
            log.Printf("loading snapshot at term %d and index %d", snapshot.Metadata.Term, snapshot.Metadata.Index)
            if err := k.recoverFromSnapshot(snapshot.Data); err != nil {
                log.Panic(err)
            }
            continue
        }

        var dataKv kv
        dec := gob.NewDecoder(bytes.NewBufferString(*data))
        if err := dec.Decode(&dataKv); err != nil {
            log.Fatalf("raftexample: could not decode message (%v)", err)
        }
        k.mu.Lock()
        k.kvStore[dataKv.Key] = dataKv.Value
        k.mu.Unlock()
    }
    if err, ok := <-errorC; ok {
        log.Fatal(err)
    }
}

// 查找key
func (k *kvstore) Lookup(key string) (string, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	val, ok := k.kvStore[key]
	return val, ok
}

// 向proposeC channel中发送key,value值
func (k *kvstore) Propose(key string, val string) {
	var buff bytes.Buffer
	// 使用gob编码
	if err := gob.NewEncoder(&buff).Encode(kv{key, val}); err != nil {
		log.Fatal(err)
	}

	k.proposeC <- string(buff.Bytes())
}

// 对kvstore的kvStore进行json化
func (k *kvstore) getSnapshot() ([]byte, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	return json.Marshal(k.kvStore)
}

// 根据byte数组恢复kvstore中的kvStore
func (k *kvstore) recoverFromSnapshot(snapshot []byte) error {
	var store map[string]string
	if err := json.Unmarshal(snapshot, &store); err != nil {
		return err
	}
	k.mu.Lock()
	k.kvStore = store
	k.mu.Unlock()
	return nil
}
