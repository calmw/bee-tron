// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mem_test

import (
	"testing"

	"github.com/calmw/bee-tron/pkg/crypto"
	"github.com/calmw/bee-tron/pkg/keystore/mem"
	"github.com/calmw/bee-tron/pkg/keystore/test"
)

func TestService(t *testing.T) {
	t.Parallel()

	t.Run("EDGSecp256_K1", func(t *testing.T) {
		test.Service(t, mem.New(), crypto.EDGSecp256_K1)
	})

	t.Run("EDGSecp256_R1", func(t *testing.T) {
		test.Service(t, mem.New(), crypto.EDGSecp256_R1)
	})
}
