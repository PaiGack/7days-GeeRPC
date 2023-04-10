package main

import (
	geerpc "GeeRPC/codec"
	"GeeRPC/codec/codec"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

func startServer(addr chan string) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("network error: %v", err)
	}

	log.Printf("start rpc server on: %v", lis.Addr())
	addr <- lis.Addr().String()
	geerpc.Accept(lis)
}

func main() {
	addr := make(chan string)
	go startServer(addr)

	conn, _ := net.Dial("tcp", <-addr)
	defer func() {
		_ = conn.Close()
	}()

	time.Sleep(time.Second)

	_ = json.NewEncoder(conn).Encode(geerpc.DefaultOption)
	cc := codec.NewGobCodec(conn)

	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}

		_ = cc.Write(h, fmt.Sprintf("geerpc req %d", h.Seq))
		_ = cc.ReadHeader(h)

		var reply string
		_ = cc.ReadBody(&reply)

		log.Printf("reply: %v", reply)
	}
}
