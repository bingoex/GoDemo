package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
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
	newServer := rpc.NewServer()
	newServer.Register(new(Arith))
	l, e := net.Listen("tcp", "127.0.0.1:9280") // any available address
	if e != nil {
		log.Fatalf("net.Listen tcp :0: %v", e)
	}

	fmt.Println("go go go")

	for {

		conn, err := l.Accept()
		if err != nil {
			continue
		}

		go newServer.ServeConn(conn)
	}
}
