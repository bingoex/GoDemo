package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

type Args struct {
	A, B int
}

type Reply struct {
	C int
}

type Arith int

func (t *Arith) Add(args Args, reply *Reply) error {
	reply.C = args.A + args.B
	return nil
}

func main() {
	client, err := rpc.Dial("tcp", "127.0.0.1:9280")
	if err != nil {
		log.Fatal(err)
	}
	client.Close()

	cnt, _ := strconv.Atoi(os.Args[1])

	args := Args{17, 8}

	var reply Reply

	for i := 0; i < cnt; i++ {
		if err := client.Call("Arith.Add", args, &reply); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("add(%v) = '%v'\n", args, reply)
	}
}
