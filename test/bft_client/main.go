package main

import (
	"fmt"
	"net/rpc"
)

func main() {
	type RequestMsg struct {
		Name   string
	}
	client, err := rpc.DialHTTP("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	defer client.Close()
	var reply string
	err = client.Call("RpcServer.Introduce", &RequestMsg{Name: "random_w"}, &reply)
	if err != nil {
		panic(err)
	}
	fmt.Println(reply)
}
