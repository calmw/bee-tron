// Copyright 2023 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package replicas

import "github.com/calmw/bee-tron/pkg/storage"

var (
	Signer = signer
)

func Wait(g storage.Getter) {
	g.(*getter).wg.Wait()
}
