// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package addressbook_test

import (
	"errors"
	"testing"

	"github.com/calmw/bee-tron/pkg/addressbook"
	"github.com/calmw/bee-tron/pkg/bzz"
	"github.com/calmw/bee-tron/pkg/crypto"
	"github.com/calmw/bee-tron/pkg/statestore/mock"
	"github.com/calmw/bee-tron/pkg/swarm"
	"github.com/ethereum/go-ethereum/common"

	ma "github.com/multiformats/go-multiaddr"
)

type bookFunc func() (book addressbook.Interface)

func TestInMem(t *testing.T) {
	t.Parallel()

	run(t, func() addressbook.Interface {
		store := mock.NewStateStore()
		book := addressbook.New(store)
		return book
	})
}

func run(t *testing.T, f bookFunc) {
	t.Helper()

	store := f()
	addr1 := swarm.NewAddress([]byte{0, 1, 2, 3})
	addr2 := swarm.NewAddress([]byte{0, 1, 2, 4})
	trxHash := common.HexToHash("0x1").Bytes()
	multiaddr, err := ma.NewMultiaddr("/ip4/1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}

	pk, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}

	bzzAddr, err := bzz.NewAddress(crypto.NewDefaultSigner(pk), multiaddr, addr1, 1, trxHash)
	if err != nil {
		t.Fatal(err)
	}

	err = store.Put(addr1, *bzzAddr)
	if err != nil {
		t.Fatal(err)
	}

	v, err := store.Get(addr1)
	if err != nil {
		t.Fatal(err)
	}

	if !bzzAddr.Equal(v) {
		t.Fatalf("expectted: %s, want %s", v, multiaddr)
	}

	notFound, err := store.Get(addr2)
	if !errors.Is(err, addressbook.ErrNotFound) {
		t.Fatal(err)
	}

	if notFound != nil {
		t.Fatalf("expected nil got %s", v)
	}

	overlays, err := store.Overlays()
	if err != nil {
		t.Fatal(err)
	}

	if len(overlays) != 1 {
		t.Fatalf("expected overlay len %v, got %v", 1, len(overlays))
	}

	addresses, err := store.Addresses()
	if err != nil {
		t.Fatal(err)
	}

	if len(addresses) != 1 {
		t.Fatalf("expected addresses len %v, got %v", 1, len(addresses))
	}
}
