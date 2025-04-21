// Copyright 2021 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pricing

import (
	"context"

	"github.com/calmw/bee-tron/pkg/p2p"
)

func (s *Service) Init(ctx context.Context, p p2p.Peer) error {
	return s.init(ctx, p)
}
