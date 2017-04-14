package main

import (
    "net/http"
    "strconv"
    "log"
)

type httpKVAPI struct {

}

// 实现的Handler接口
func (h *httpKVAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    key := r.RequestURI
    switch {
    case r.Method == "PUT":
        w.WriteHeader(http.StatusNoContent)
    case r.Method == "POST":
        w.WriteHeader(http.StatusNoContent)
    case r.Method == "GET":
        w.Write([]byte(key))
    case r.Method == "DELETE":
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

func serveHttpKVAPI(kvport int)  {
    srv := http.Server{
        Addr: ":" + strconv.Itoa(kvport),
        Handler: &httpKVAPI{

        },
    }

    if err := srv.ListenAndServe(); err != nil {
        log.Fatal(err)
    }

}