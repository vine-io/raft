package hello

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vine-io/apimachinery/runtime"
	"github.com/vine-io/raft"
	pb "github.com/vine-io/raft/test/proto"
	"github.com/vine-io/vine"
)

var _ pb.TestHandler = (*HelloServer)(nil)

type HelloServer struct {
	rnode  raft.RaftNode
	scheme runtime.Scheme
}

func NewServer(rnode raft.RaftNode, scheme runtime.Scheme) *HelloServer {
	return &HelloServer{rnode: rnode, scheme: scheme}
}

func (s *HelloServer) GetProduct(ctx *vine.Context, in *pb.GetProductRequest, out *pb.GetProductResponse) error {

	obj := s.scheme.Default(&pb.Product{})

	result, err := s.rnode.Get(ctx, &raft.GetOption{
		GVK:  obj.GetObjectKind().GroupVersionKind(),
		Id:   in.Id,
		List: false,
	})
	if err != nil {
		return err
	}

	out.Product = result.Out.(*pb.Product)
	return nil
}

func (s *HelloServer) CreateProduct(ctx *vine.Context, in *pb.CreateProductRequest, out *pb.CreateProductResponse) error {

	product := &pb.Product{
		Id:      uuid.New().String(),
		Date:    time.Now().Unix(),
		Company: in.Company,
	}
	body, _ := product.Marshal()
	frame := &pb.Frame{
		Op:   pb.Op_Create,
		Gvk:  s.scheme.Default(product).GetObjectKind().GroupVersionKind().String(),
		Body: body,
	}
	data, _ := frame.Marshal()

	err := s.rnode.Propose(context.TODO(), data)
	if err != nil {
		return err
	}

	out.Product = product
	return nil
}
