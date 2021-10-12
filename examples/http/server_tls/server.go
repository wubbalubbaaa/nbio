package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/lesismal/nbio/nbhttp"
)

var (
	qps   uint64 = 0
	total uint64 = 0
)

// visit: https://localhost:8888/echo
func onEcho(w http.ResponseWriter, r *http.Request) {
	data, _ := io.ReadAll(r.Body)
	if len(data) > 0 {
		w.Write(data)
	} else {
		w.Write([]byte(time.Now().Format("20060102 15:04:05")))
	}
	atomic.AddUint64(&qps, 1)
}

func main() {
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalf("tls.X509KeyPair failed: %v", err)
	}
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	tlsConfig.BuildNameToCertificate()

	mux := &http.ServeMux{}
	mux.HandleFunc("/echo", onEcho)

	confTLS := nbhttp.Config{
		Network: "tcp",
		Addrs:   []string{"localhost:8888"},
	}
	// create common executor pool
	serverExecutorPool := nbhttp.NewServerExecutorPool(&confTLS)
	defer serverExecutorPool.Stop()

	// start tls server
	svrTLS := nbhttp.NewServerTLS(confTLS, mux, serverExecutorPool.Go, tlsConfig)

	err = svrTLS.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}
	defer svrTLS.Stop()

	confNonTLS := nbhttp.Config{
		Network: "tcp",
		Addrs:   []string{"localhost:8887"},
	}
	// start non tls server
	svrNonTLS := nbhttp.NewServer(confNonTLS, mux, serverExecutorPool.Go)
	err = svrNonTLS.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}
	defer svrNonTLS.Stop()

	// stats
	ticker := time.NewTicker(time.Second)
	for i := 1; true; i++ {
		<-ticker.C
		n := atomic.SwapUint64(&qps, 0)
		total += n
		fmt.Printf("running for %v seconds, NumGoroutine: %v, qps: %v, total: %v\n", i, runtime.NumGoroutine(), n, total)
	}
}
