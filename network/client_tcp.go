package main

import (
	"fmt"
	"net"
)

func main() {
	sendMessage()
	select {}
}

func sendMessage() {
	conn, err := net.Dial("tcp", "10.139.11.42:5006")

	if err != nil {
		panic("error")
	}

	header := "GET / HTTP/1.0\r\n\r\n"
	fmt.Fprintf(conn, header)
	println("done")
}
