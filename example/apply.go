package main

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

func (a *DaoApplier) List(ctx context.Context, option raft.ListOption, iterFunc raft.IterFunc) (int64, error) {
	sc, _, err := a.selectStorage(option.In)
	if err != nil {
		return 0, err
	}

	var list []runtime.Object
	var n int64
	if option.In != nil && option.Page != 0 && option.Size != 0 {
		orderBy := clause.OrderBy{
			Columns: []clause.OrderByColumn{{Column: clause.Column{Name: "creation_timestamp"}, Desc: true}},
		}
		list, n, err = sc.Cond(orderBy).Cond(option.Exprs...).FindPage(ctx, option.Page, option.Size)
	} else {
		list, err = sc.Cond(option.Exprs...).FindAll(ctx)
		n = int64(len(list))
	}

	if err != nil {
		return 0, err
	}

	for idx, item := range list {
		iterFunc(item, idx)
	}

	return n, nil
}

func (a *DaoApplier) Get(ctx context.Context, in runtime.Object, id string) error {
	sc, _, err := a.selectStorage(in)
	if err != nil {
		return err
	}

	out, err := sc.Cond(clause.Cond().Op(clause.EqOp).Build("uid", id)).FindOne(ctx)
	if err != nil {
		return err
	}
	in.DeepFrom(out)
	in = a.Schema.Default(in)
	return nil
}

func (a *DaoApplier) Find(ctx context.Context, in runtime.Object) error {
	sc, _, err := a.selectStorage(in)
	if err != nil {
		return err
	}

	out, err := sc.FindOne(ctx)
	if err != nil {
		return err
	}

	in.DeepFrom(out)
	in = a.Schema.Default(in)
	return nil
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

func (a *DaoApplier) SetEpoch(ctx context.Context, term, index uint64) error {
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
