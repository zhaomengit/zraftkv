package main

import (
	"github.com/coreos/etcd/raft/raftpb"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type httpKVAPI struct {
	store       *kvstore
	confChangeC chan<- raftpb.ConfChange // 发送raft confchange channel
}

// 实现的Handler接口
func (h *httpKVAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 获取key
	key := r.RequestURI
	switch {
	case r.Method == "PUT":
		v, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read on PUT (%v)\n", err)
			http.Error(w, "Failed on PUT", http.StatusBadRequest)
			return
		}

		// 向channel
		h.store.Propose(key, string(v))
		w.WriteHeader(http.StatusNoContent)

	case r.Method == "POST": // 增加节点

		url, err := ioutil.ReadAll(r.Body)

		if err != nil {
			log.Printf("Failed to read on POST (%v)\n", err)
			http.Error(w, "Failed on POST", http.StatusBadRequest)
			return
		}

		nodeId, err := strconv.ParseUint(key[1:], 0, 64)
		if err != nil {
			log.Printf("Failed to convert ID for conf change (%v)\n", err)
			http.Error(w, "Failed on POST", http.StatusBadRequest)
			return
		}

		cc := raftpb.ConfChange{
			Type:    raftpb.ConfChangeAddNode, // 增加节点
			NodeID:  nodeId,
			Context: url,
		}
		h.confChangeC <- cc

		w.WriteHeader(http.StatusNoContent)
	case r.Method == "GET":

		// get从store的map中查找key
		if v, ok := h.store.Lookup(key); ok {
			w.Write([]byte(v))
		} else {
			http.Error(w, "Failed to GET", http.StatusNotFound)
		}

	case r.Method == "DELETE": // 删除节点

		// 取nodeid
		nodeId, err := strconv.ParseUint(key[1:], 0, 64)

		if err != nil {
			log.Printf("Failed to convert ID for conf change (%v)\n %d", err, nodeId)
			http.Error(w, "Failed on DELETE", http.StatusBadRequest)
			return
		}

		// 构造raft confchange对象
		cc := raftpb.ConfChange{
			Type:   raftpb.ConfChangeRemoveNode, // 删除节点
			NodeID: nodeId,
		}
		// 发送
		h.confChangeC <- cc
		w.WriteHeader(http.StatusNoContent)

	default:
		// 设置Head头信息
		w.Header().Set("Allow", "PUT")
		w.Header().Add("Allow", "GET")
		w.Header().Add("Allow", "POST")
		w.Header().Add("Allow", "DELETE")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

// 开启api REST 服务
func serveHttpKVAPI(kv *kvstore, port int, confChangeC chan<- raftpb.ConfChange, errorC <-chan error) {
	srv := http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: &httpKVAPI{
			store:       kv,
			confChangeC: confChangeC,
		},
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
