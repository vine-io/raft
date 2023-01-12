package hello

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	metav1 "github.com/vine-io/apimachinery/apis/meta/v1"
	"github.com/vine-io/apimachinery/runtime"
	"github.com/vine-io/apimachinery/schema"
	"github.com/vine-io/apimachinery/storage"
	"github.com/vine-io/raft"
	pb "github.com/vine-io/raft/test/proto"
	"github.com/vine-io/vine/lib/dao/clause"
)

var _ raft.Applier = (*DaoApplier)(nil)

type DaoApplier struct {
	Schema  runtime.Scheme
	Factory storage.Factory
}

func NewApplier(schema runtime.Scheme, factory storage.Factory) (raft.Applier, error) {
	applier := &DaoApplier{schema, factory}
	return applier, nil
}

func (a *DaoApplier) Get(ctx context.Context, option *raft.GetOption) (*raft.GetResult, error) {

	in, err := a.Schema.New(option.GVK)
	if err != nil {
		return nil, err
	}

	if !option.List && option.Id != "" {
		err = a.getById(ctx, in, option.Id)
		if err != nil {
			return nil, err
		}
		return &raft.GetResult{Out: in}, nil
	}

	if !option.List && option.Id == "" {
		err = a.Find(ctx, in, option.Exprs...)
		if err != nil {
			return nil, err
		}
		return &raft.GetResult{Out: in}, nil
	}

	list, count, err := a.List(ctx, option.Page, option.Size, in, option.Exprs...)
	if err != nil {
		return nil, err
	}

	return &raft.GetResult{List: true, Out: list, Total: &count}, nil
}

func (a *DaoApplier) List(ctx context.Context, page, size int32, in runtime.Object, exprs ...clause.Expression) ([]runtime.Object, int64, error) {
	sc, _, err := a.selectStorage(in)
	if err != nil {
		return nil, 0, err
	}

	var list []runtime.Object
	var n int64
	if in != nil && page != 0 && size != 0 {
		orderBy := clause.OrderBy{
			Columns: []clause.OrderByColumn{{Column: clause.Column{Name: "creation_timestamp"}, Desc: true}},
		}
		list, n, err = sc.Cond(orderBy).Cond(exprs...).FindPage(ctx, page, size)
	} else {
		list, err = sc.Cond(exprs...).FindAll(ctx)
		n = int64(len(list))
	}

	if err != nil {
		return nil, 0, err
	}

	return list, n, nil
}

func (a *DaoApplier) getById(ctx context.Context, in runtime.Object, id string) error {
	sc, _, err := a.selectStorage(in)
	if err != nil {
		return err
	}

	out, err := sc.Cond(clause.Cond().Op(clause.EqOp).Build("id", id)).FindOne(ctx)
	if err != nil {
		return err
	}
	in.DeepFrom(out)
	in = a.Schema.Default(in)
	return nil
}

func (a *DaoApplier) Find(ctx context.Context, in runtime.Object, exprs ...clause.Expression) error {
	sc, _, err := a.selectStorage(in)
	if err != nil {
		return err
	}

	out, err := sc.Cond(exprs...).FindOne(ctx)
	if err != nil {
		return err
	}

	in.DeepFrom(out)
	in = a.Schema.Default(in)
	return nil
}

func (a *DaoApplier) Put(ctx context.Context, data []byte) error {

	frame := &pb.Frame{}
	err := frame.Unmarshal(data)
	if err != nil {
		return err
	}
	gvk := schema.FromGVK(frame.Gvk)
	obj, err := a.Schema.New(gvk)
	if err != nil {
		return err
	}

	v, ok := obj.(interface {
		runtime.Object
		Unmarshal([]byte) error
	})
	if !ok {
		return fmt.Errorf("bad body")
	}

	err = v.Unmarshal(frame.Body)
	if err != nil {
		return err
	}

	switch frame.Op {
	case pb.Op_Create:
		err = a.Create(ctx, v)
	case pb.Op_Update:
		err = a.Update(ctx, v)
	case pb.Op_Delete:
		err = a.Delete(ctx, v, true)
	}

	return err
}

func (a *DaoApplier) Create(ctx context.Context, in runtime.Object) error {
	sc, _, err := a.selectStorage(in)
	if err != nil {
		return err
	}
	if v, ok := in.(metav1.Meta); ok {
		if v.GetUID() == "" {
			v.SetUID(uuid.New().String())
		}
		if v.GetCreationTimestamp() == 0 {
			v.SetCreationTimestamp(time.Now().Unix())
		}
	}

	out, err := sc.Create(ctx)
	if err != nil {
		return err
	}
	in.DeepFrom(out)

	return nil
}

func (a *DaoApplier) Update(ctx context.Context, in runtime.Object) error {
	sc, _, err := a.selectStorage(in)
	if err != nil {
		return err
	}

	out, err := sc.Updates(ctx)
	if err != nil {
		return err
	}
	in.DeepFrom(out)

	return nil
}

func (a *DaoApplier) Delete(ctx context.Context, in runtime.Object, soft bool) error {
	sc, _, err := a.selectStorage(in)
	if err != nil {
		return err
	}

	err = sc.Delete(ctx, soft)
	if err != nil {
		return err
	}

	return nil
}

func (a *DaoApplier) GetEpoch(ctx context.Context) (uint64, uint64, error) {
	epoch := &pb.Epoch{}

	err := a.Find(ctx, epoch)
	if err != nil {
		return 0, 0, nil
	}
	return uint64(epoch.Term), uint64(epoch.Index), nil
}

func (a *DaoApplier) ApplyEpoch(ctx context.Context, term, index uint64) error {
	epoch := &pb.Epoch{}

	_ = a.Find(ctx, epoch)
	epoch.Term = int64(term)
	epoch.Index = int64(index)

	if epoch.Id == "" {
		epoch.Id = uuid.New().String()
		return a.Create(ctx, epoch)
	}
	return a.Update(ctx, epoch)
}

func (a *DaoApplier) GetSnapshot(ctx context.Context) ([]byte, error) {

	snapshot := &pb.SnapshotMarshaller{}
	snapshot.Products, _ = pb.ProductStorageBuilder().FindAllEntities(ctx)

	return snapshot.Marshal()
}

func (a *DaoApplier) RecoverFromSnapshot(ctx context.Context, data []byte) error {

	snapshot := &pb.SnapshotMarshaller{}
	if err := snapshot.Unmarshal(data); err != nil {
		return err
	}

	_ = pb.ProductStorageBuilder().BatchDelete(ctx, true)
	for _, item := range snapshot.Products {
		pb.FromProduct(item).Create(ctx)
	}

	return nil
}

func (a *DaoApplier) selectStorage(in runtime.Object) (storage.Storage, schema.GroupVersionKind, error) {
	in = a.Schema.Default(in)
	gvk := in.GetObjectKind().GroupVersionKind()
	sc, err := a.Factory.NewStorage(in)
	if err != nil {
		return nil, gvk, fmt.Errorf("don't support '%v' schema", gvk.String())
	}
	return sc, gvk, nil
}
