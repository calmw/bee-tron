// Copyright 2023 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package migration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/calmw/bee-tron/pkg/log"
	"github.com/calmw/bee-tron/pkg/storage/inmemstore"
	"github.com/calmw/bee-tron/pkg/storer/internal"
	"github.com/calmw/bee-tron/pkg/storer/internal/transaction"

	"github.com/calmw/bee-tron/pkg/storage/migration"
	localmigration "github.com/calmw/bee-tron/pkg/storer/migration"
)

func TestPreSteps(t *testing.T) {
	t.Parallel()

	store := internal.NewInmemStorage()

	assert.NotEmpty(t, localmigration.AfterInitSteps("", 0, store, log.Noop))

	t.Run("version numbers", func(t *testing.T) {
		t.Parallel()

		err := migration.ValidateVersions(localmigration.AfterInitSteps("", 0, store, log.Noop))
		assert.NoError(t, err)
	})

	t.Run("zero store migration", func(t *testing.T) {
		t.Parallel()

		store := internal.NewInmemStorage()
		err := store.Run(context.Background(), func(s transaction.Store) error {
			return migration.Migrate(s.IndexStore(), "migration", localmigration.AfterInitSteps("", 4, store, log.Noop))
		})
		assert.NoError(t, err)
	})
}

func TestPostSteps(t *testing.T) {
	t.Parallel()

	st := inmemstore.New()

	assert.NotEmpty(t, localmigration.BeforeInitSteps(st, log.Noop))

	t.Run("version numbers", func(t *testing.T) {
		t.Parallel()

		err := migration.ValidateVersions(localmigration.BeforeInitSteps(st, log.Noop))
		assert.NoError(t, err)
	})

	t.Run("zero store migration", func(t *testing.T) {
		t.Parallel()

		store := inmemstore.New()

		err := migration.Migrate(store, "migration", localmigration.BeforeInitSteps(store, log.Noop))
		assert.NoError(t, err)
	})
}
