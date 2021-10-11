package main

import (
	"fmt"
	"github.com/Grivn/phalanx/common/mocks"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	phalanx "github.com/Grivn/phalanx/core"
	"log"
	"net/http"
	_ "net/http/pprof"
	"net/rpc"
	"os"
	"strconv"
)

type Student struct {
	Name   string
	School string
}
type RpcServer struct{}

func (r *RpcServer) Introduce(student Student, words *string) error {
	fmt.Println("student: ", student)


	n := 4

	async := false

	nc := make(map[uint64]chan *protos.ConsensusMessage)

	cc := make(map[uint64]chan *protos.Command)

	closeC := make(chan bool)

	phx := make(map[uint64]phalanx.Provider)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		nc[id] = make(chan *protos.ConsensusMessage)
		cc[id] = make(chan *protos.Command)
	}
	net := mocks.NewSimpleNetwork(nc, cc, types.NewRawLogger(), async)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		exec := mocks.NewSimpleExecutor(id, types.NewRawLogger())
		phx[id] = phalanx.NewPhalanxProvider(n, id, exec, net, types.NewRawLogger())
	}

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		go phalanxListener(phx[id], nc[id], cc[id], closeC)
	}

	replicas := make(map[uint64]*replica)
	bftCs := make(map[uint64]chan *bftMessage)
	sendC := make(chan *bftMessage)
	logDir := "bft_nodes"
	_ = os.Mkdir(logDir, os.ModePerm)
	for i:=0; i<n; i++ {
		id := uint64(i+1)
		bftCs[id] = make(chan *bftMessage)
		replicas[id] = newReplica(n, id, phx[id], sendC, bftCs[id], closeC, types.NewRawLoggerFile(logDir+"/bft-node-"+strconv.Itoa(i+1)+".log"))
		replicas[id].run()
	}
	go cluster(sendC, bftCs, closeC)

	num := 1000
	client := 4
	//transactionSendInstance(num, client, phx)
	commandSendInstance(num, client, phx)

	*words = fmt.Sprintf("Hello everyone, my name is %s, and I am from %s", student.Name, student.School)
	return nil
}

func main() {
	rpcServer := new(RpcServer)
	// 注册rpc服务
	_ = rpc.Register(rpcServer)
	//把服务处理绑定到http协议上
	rpc.HandleHTTP()
	log.Println("http rpc service start success addr:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
