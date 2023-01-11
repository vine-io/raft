package main

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vine-io/raft"
	pb "github.com/vine-io/raft/test/proto"
	"github.com/vine-io/vine"
)

var _ pb.TestHandler = (*HelloServer)(nil)

type HelloServer struct {
	rnode raft.RaftNode
	ps    raft.PersistentStorage
}

func NewServer(rnode raft.RaftNode, ps raft.PersistentStorage) *HelloServer {
	return &HelloServer{rnode: rnode, ps: ps}
}

func (s *HelloServer) GetProduct(ctx *vine.Context, in *pb.GetProductRequest, out *pb.GetProductResponse) error {
	product := &pb.Product{}
	err := s.ps.Get(ctx, product, in.Id)
	if err != nil {
		return err
	}

	out = &pb.GetProductResponse{Product: product}
	return nil
}

func (s *HelloServer) CreateProduct(ctx *vine.Context, in *pb.CreateProductRequest, out *pb.CreateProductResponse) error {

	product := &pb.Product{
		Id:      uuid.New().String(),
		Date:    time.Now().Unix(),
		Company: in.Company,
	}
	op := &raft.Operation{Op: raft.Create, Target: product}

	err := s.rnode.Propose(context.TODO(), op)
	if err != nil {
		return err
	}

	out = &pb.CreateProductResponse{Product: product}
	return nil
}
