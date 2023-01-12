package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/vine-io/raft/test/proto"
	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/core/client/grpc"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:12379", "server address")
	c := flag.String("company", "c1", "")

	flag.Parse()

	if *c == "" {
		fmt.Println("missing company")
		os.Exit(1)
	}

	defaultTimeout := time.Second * 30
	client.DefaultDialTimeout = defaultTimeout
	client.DefaultRequestTimeout = defaultTimeout

	conn := pb.NewTestService("go.vine.server", grpc.NewClient())

	ctx := context.TODO()

	rsp, err := conn.CreateProduct(ctx, &pb.CreateProductRequest{Company: *c}, client.WithAddress(*addr))
	if err != nil {
		log.Fatal(err)
	}

	log.Println(rsp.String())
}
