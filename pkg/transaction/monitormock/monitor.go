// Copyright 2021 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package monitormock

import (
	"context"
	"errors"
	"math/big"

	"github.com/calmw/bee-tron/pkg/transaction"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type transactionMonitorMock struct {
	watchTransaction func(txHash common.Hash, nonce uint64) (<-chan types.Receipt, <-chan error, error)
	waitBlock        func(ctx context.Context, block *big.Int) (*types.Block, error)
}

func (m *transactionMonitorMock) WatchTransaction(txHash common.Hash, nonce uint64) (<-chan types.Receipt, <-chan error, error) {
	if m.watchTransaction != nil {
		return m.watchTransaction(txHash, nonce)
	}
	return nil, nil, errors.New("not implemented")
}

func (m *transactionMonitorMock) WaitBlock(ctx context.Context, block *big.Int) (*types.Block, error) {
	if m.watchTransaction != nil {
		return m.waitBlock(ctx, block)
	}
	return nil, errors.New("not implemented")
}

func (m *transactionMonitorMock) Close() error {
	return nil
}

// Option is the option passed to the mock Chequebook service
type Option interface {
	apply(*transactionMonitorMock)
}

type optionFunc func(*transactionMonitorMock)

func (f optionFunc) apply(r *transactionMonitorMock) { f(r) }

func WithWatchTransactionFunc(f func(txHash common.Hash, nonce uint64) (<-chan types.Receipt, <-chan error, error)) Option {
	return optionFunc(func(s *transactionMonitorMock) {
		s.watchTransaction = f
	})
}

func WithWaitBlockFunc(f func(ctx context.Context, block *big.Int) (*types.Block, error)) Option {
	return optionFunc(func(s *transactionMonitorMock) {
		s.waitBlock = f
	})
}

func New(opts ...Option) transaction.Monitor {
	mock := new(transactionMonitorMock)
	for _, o := range opts {
		o.apply(mock)
	}
	return mock
}
