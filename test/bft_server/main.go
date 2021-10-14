package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"net/rpc"
)

type RequestMsg struct {
	Name   string
}
type RpcServer struct{}

func (r *RpcServer) Introduce(req RequestMsg, words *string) error {
	fmt.Println("request: ", req)
	phalanxRunner()
	*words = fmt.Sprintf("hello from %s", req.Name)
	return nil
}

func main() {
	rpcServer := new(RpcServer)
	_ = rpc.Register(rpcServer)
	rpc.HandleHTTP()
	log.Println("http rpc service start success addr:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
