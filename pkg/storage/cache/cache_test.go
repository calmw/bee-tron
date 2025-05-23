// Copyright 2023 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache_test

import (
	"testing"

	"github.com/calmw/bee-tron/pkg/storage/cache"
	"github.com/calmw/bee-tron/pkg/storage/leveldbstore"
	"github.com/calmw/bee-tron/pkg/storage/storagetest"
	"github.com/calmw/bee-tron/pkg/util/testutil"
)

func TestCache(t *testing.T) {
	t.Parallel()

	store, err := leveldbstore.New(t.TempDir(), nil)
	if err != nil {
		t.Fatalf("create store failed: %v", err)
	}
	testutil.CleanupCloser(t, store)

	cache, err := cache.Wrap(store, 100_000)
	if err != nil {
		t.Fatalf("create cache failed: %v", err)
	}

	storagetest.TestStore(t, cache)
}
