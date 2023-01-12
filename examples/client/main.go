package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	pb "github.com/vine-io/raft/test/proto"
	"github.com/vine-io/vine/core/client"
	"github.com/vine-io/vine/core/client/grpc"
)

func main() {
	c := flag.String("company", "", "")

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

	fmt.Println(conn.GetProduct(ctx, &pb.GetProductRequest{Id: uuid.New().String()}))
	rsp, err := conn.CreateProduct(ctx, &pb.CreateProductRequest{Company: *c}, client.WithAddress("127.0.0.1:33379"))
	if err != nil {
		log.Fatal(err)
	}

	log.Println(rsp.String())
}
