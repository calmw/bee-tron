// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package intervalstore

import (
	"errors"
	"testing"

	"github.com/calmw/bee-tron/pkg/log"
	"github.com/calmw/bee-tron/pkg/statestore/leveldb"
	"github.com/calmw/bee-tron/pkg/statestore/mock"
	"github.com/calmw/bee-tron/pkg/storage"
	"github.com/calmw/bee-tron/pkg/util/testutil"
)

// TestInmemoryStore tests basic functionality of InmemoryStore.
func TestInmemoryStore(t *testing.T) {
	t.Parallel()

	testStore(t, mock.NewStateStore())
}

// TestDBStore tests basic functionality of DBStore.
func TestDBStore(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	store, err := leveldb.NewStateStore(dir, log.Noop)
	if err != nil {
		t.Fatal(err)
	}
	testutil.CleanupCloser(t, store)

	testStore(t, store)
}

// testStore is a helper function to test various Store implementations.
func testStore(t *testing.T, s storage.StateStorer) {
	t.Helper()

	key1 := "key1"
	i1 := NewIntervals(0)
	i1.Add(10, 20)
	if err := s.Put(key1, i1); err != nil {
		t.Fatal(err)
	}
	i := &Intervals{}
	err := s.Get(key1, i)
	if err != nil {
		t.Fatal(err)
	}
	if i.String() != i1.String() {
		t.Errorf("expected interval %s, got %s", i1, i)
	}

	key2 := "key2"
	i2 := NewIntervals(0)
	i2.Add(10, 20)
	if err := s.Put(key2, i2); err != nil {
		t.Fatal(err)
	}
	err = s.Get(key2, i)
	if err != nil {
		t.Fatal(err)
	}
	if i.String() != i2.String() {
		t.Errorf("expected interval %s, got %s", i2, i)
	}

	if err := s.Delete(key1); err != nil {
		t.Fatal(err)
	}
	if err := s.Get(key1, i); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("expected error %v, got %s", storage.ErrNotFound, err)
	}
	if err := s.Get(key2, i); err != nil {
		t.Errorf("expected error %v, got %s", nil, err)
	}

	if err := s.Delete(key2); err != nil {
		t.Fatal(err)
	}
	if err := s.Get(key2, i); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("expected error %v, got %s", storage.ErrNotFound, err)
	}
}
