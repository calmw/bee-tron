// Copyright 2021 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package stewardess provides convenience methods
// for reseeding content on Swarm.
package steward

import (
	"context"
	"errors"
	"fmt"

	"github.com/calmw/bee-tron/pkg/file/redundancy"
	"github.com/calmw/bee-tron/pkg/postage"
	"github.com/calmw/bee-tron/pkg/retrieval"
	"github.com/calmw/bee-tron/pkg/storage"
	"github.com/calmw/bee-tron/pkg/storer"
	"github.com/calmw/bee-tron/pkg/swarm"
	"github.com/calmw/bee-tron/pkg/topology"
	"github.com/calmw/bee-tron/pkg/traversal"
)

type Interface interface {
	// Reupload root hash and all of its underlying
	// associated chunks to the network.
	Reupload(context.Context, swarm.Address, postage.Stamper) error

	// IsRetrievable checks whether the content
	// on the given address is retrievable.
	IsRetrievable(context.Context, swarm.Address) (bool, error)
}

type steward struct {
	netStore     storer.NetStore
	traverser    traversal.Traverser
	netTraverser traversal.Traverser
	netGetter    retrieval.Interface
}

func New(ns storer.NetStore, r retrieval.Interface, joinerPutter storage.Putter) Interface {
	return &steward{
		netStore:     ns,
		traverser:    traversal.New(ns.Download(true), joinerPutter, redundancy.DefaultLevel),
		netTraverser: traversal.New(&netGetter{r}, joinerPutter, redundancy.DefaultLevel),
		netGetter:    r,
	}
}

// Reupload content with the given root hash to the network.
// The service will automatically dereference and traverse all
// addresses and push every chunk individually to the network.
// It assumes all chunks are available locally. It is therefore
// advisable to pin the content locally before trying to reupload it.
func (s *steward) Reupload(ctx context.Context, root swarm.Address, stamper postage.Stamper) error {
	uploaderSession := s.netStore.DirectUpload()
	getter := s.netStore.Download(false)

	fn := func(addr swarm.Address) error {
		c, err := getter.Get(ctx, addr)
		if err != nil {
			return err
		}

		stamp, err := stamper.Stamp(c.Address(), c.Address())
		if err != nil {
			return fmt.Errorf("stamping chunk %s: %w", c.Address(), err)
		}

		return uploaderSession.Put(ctx, c.WithStamp(stamp))
	}

	if err := s.traverser.Traverse(ctx, root, fn); err != nil {
		return errors.Join(
			fmt.Errorf("traversal of %s failed: %w", root.String(), err),
			uploaderSession.Cleanup(),
		)
	}

	return uploaderSession.Done(root)
}

// IsRetrievable implements Interface.IsRetrievable method.
func (s *steward) IsRetrievable(ctx context.Context, root swarm.Address) (bool, error) {
	fn := func(a swarm.Address) error {
		_, err := s.netGetter.RetrieveChunk(ctx, a, swarm.ZeroAddress)
		return err
	}
	switch err := s.netTraverser.Traverse(ctx, root, fn); {
	case errors.Is(err, storage.ErrNotFound):
		return false, nil
	case errors.Is(err, topology.ErrNotFound):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("traversal of %q failed: %w", root, err)
	default:
		return true, nil
	}
}

// netGetter implements the storage Getter.Get method in a way
// that it will try to retrieve the chunk only from the network.
type netGetter struct {
	retrieval retrieval.Interface
}

// Get implements the storage Getter.Get interface.
func (ng *netGetter) Get(ctx context.Context, addr swarm.Address) (swarm.Chunk, error) {
	return ng.retrieval.RetrieveChunk(ctx, addr, swarm.ZeroAddress)
}
