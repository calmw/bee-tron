// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package addresses_test

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/calmw/bee-tron/pkg/file"
	"github.com/calmw/bee-tron/pkg/file/addresses"
	"github.com/calmw/bee-tron/pkg/file/joiner"
	"github.com/calmw/bee-tron/pkg/file/redundancy"
	filetest "github.com/calmw/bee-tron/pkg/file/testing"
	"github.com/calmw/bee-tron/pkg/storage/inmemchunkstore"
	"github.com/calmw/bee-tron/pkg/swarm"
)

func TestAddressesGetterIterateChunkAddresses(t *testing.T) {
	t.Parallel()

	store := inmemchunkstore.New()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// create root chunk with 2 references and the referenced data chunks
	rootChunk := filetest.GenerateTestRandomFileChunk(swarm.ZeroAddress, swarm.ChunkSize*2, swarm.SectionSize*2)
	err := store.Put(ctx, rootChunk)
	if err != nil {
		t.Fatal(err)
	}

	firstAddress := swarm.NewAddress(rootChunk.Data()[8 : swarm.SectionSize+8])
	firstChunk := filetest.GenerateTestRandomFileChunk(firstAddress, swarm.ChunkSize, swarm.ChunkSize)
	err = store.Put(ctx, firstChunk)
	if err != nil {
		t.Fatal(err)
	}

	secondAddress := swarm.NewAddress(rootChunk.Data()[swarm.SectionSize+8:])
	secondChunk := filetest.GenerateTestRandomFileChunk(secondAddress, swarm.ChunkSize, swarm.ChunkSize)
	err = store.Put(ctx, secondChunk)
	if err != nil {
		t.Fatal(err)
	}

	createdAddresses := []swarm.Address{rootChunk.Address(), firstAddress, secondAddress}

	foundAddresses := make(map[string]struct{})
	var foundAddressesMu sync.Mutex

	addressIterFunc := func(addr swarm.Address) error {
		foundAddressesMu.Lock()
		defer foundAddressesMu.Unlock()

		foundAddresses[addr.String()] = struct{}{}
		return nil
	}

	addressesGetter := addresses.NewGetter(store, addressIterFunc)

	j, _, err := joiner.New(ctx, addressesGetter, store, rootChunk.Address(), redundancy.DefaultLevel)
	if err != nil {
		t.Fatal(err)
	}

	_, err = file.JoinReadAll(ctx, j, io.Discard)
	if err != nil {
		t.Fatal(err)
	}

	if len(createdAddresses) != len(foundAddresses) {
		t.Fatalf("expected to find %d addresses, got %d", len(createdAddresses), len(foundAddresses))
	}

	checkAddressFound := func(t *testing.T, foundAddresses map[string]struct{}, address swarm.Address) {
		t.Helper()

		if _, ok := foundAddresses[address.String()]; !ok {
			t.Fatalf("expected address %s not found", address.String())
		}
	}

	for _, createdAddress := range createdAddresses {
		checkAddressFound(t, foundAddresses, createdAddress)
	}
}
