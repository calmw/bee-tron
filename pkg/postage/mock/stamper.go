// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mock

import (
	"github.com/calmw/bee-tron/pkg/postage"
	"github.com/calmw/bee-tron/pkg/swarm"
)

type mockStamper struct{}

// NewStamper returns anew new mock stamper.
func NewStamper() postage.Stamper {
	return &mockStamper{}
}

// Stamp implements the Stamper interface. It returns an empty postage stamp.
func (mockStamper) Stamp(_, _ swarm.Address) (*postage.Stamp, error) {
	return &postage.Stamp{}, nil
}

// Stamp implements the Stamper interface. It returns an empty postage stamp.
func (mockStamper) BatchId() []byte {
	return nil
}
