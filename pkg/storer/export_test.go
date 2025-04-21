// Copyright 2023 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package storer

import (
	"github.com/calmw/bee-tron/pkg/storer/internal/events"
	"github.com/calmw/bee-tron/pkg/storer/internal/reserve"
)

func (db *DB) Reserve() *reserve.Reserve {
	return db.reserve
}

func (db *DB) Events() *events.Subscriber {
	return db.events
}

func ReplaceSharkyShardLimit(val int) {
	sharkyNoOfShards = val
}

func (db *DB) WaitForBgCacheWorkers() (unblock func()) {
	for i := 0; i < defaultBgCacheWorkers; i++ {
		db.cacheLimiter.sem <- struct{}{}
	}
	return func() {
		for i := 0; i < defaultBgCacheWorkers; i++ {
			<-db.cacheLimiter.sem
		}
	}
}

func DefaultOptions() *Options {
	return defaultOptions()
}
