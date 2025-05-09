// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chequebook_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/calmw/bee-tron/pkg/settlement/swap/chequebook"
	storemock "github.com/calmw/bee-tron/pkg/statestore/mock"
	transactionmock "github.com/calmw/bee-tron/pkg/transaction/mock"
	"github.com/ethereum/go-ethereum/common"
)

func TestReceiveCheque(t *testing.T) {
	t.Parallel()

	store := storemock.NewStateStore()
	beneficiary := common.HexToAddress("0xffff")
	issuer := common.HexToAddress("0xbeee")
	cumulativePayout := big.NewInt(101)
	cumulativePayout2 := big.NewInt(201)
	chequebookAddress := common.HexToAddress("0xeeee")
	sig := make([]byte, 65)
	chainID := int64(1)
	exchangeRate := big.NewInt(10)
	deduction := big.NewInt(1)

	cheque := &chequebook.SignedCheque{
		Cheque: chequebook.Cheque{
			Beneficiary:      beneficiary,
			CumulativePayout: cumulativePayout,
			Chequebook:       chequebookAddress,
		},
		Signature: sig,
	}

	var verifiedWithFactory bool
	factory := &factoryMock{
		verifyChequebook: func(ctx context.Context, address common.Address) error {
			if address != chequebookAddress {
				t.Fatal("verifying wrong chequebook")
			}
			verifiedWithFactory = true
			return nil
		},
	}

	chequestore := chequebook.NewChequeStore(
		store,
		factory,
		chainID,
		beneficiary,
		transactionmock.New(
			transactionmock.WithABICallSequence(
				transactionmock.ABICall(&chequebookABI, chequebookAddress, common.BytesToHash(issuer.Bytes()).Bytes(), "issuer"),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, cumulativePayout2.FillBytes(make([]byte, 32)), "balance"),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, big.NewInt(0).FillBytes(make([]byte, 32)), "paidOut", beneficiary),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, common.BytesToHash(issuer.Bytes()).Bytes(), "issuer"),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, cumulativePayout2.FillBytes(make([]byte, 32)), "balance"),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, big.NewInt(0).FillBytes(make([]byte, 32)), "paidOut", beneficiary),
			),
		),
		func(c *chequebook.SignedCheque, cid int64) (common.Address, error) {
			if cid != chainID {
				t.Fatalf("recovery with wrong chain id. wanted %d, got %d", chainID, cid)
			}
			if !cheque.Equal(c) {
				t.Fatalf("recovery with wrong cheque. wanted %v, got %v", cheque, c)
			}
			return issuer, nil
		})

	received, err := chequestore.ReceiveCheque(context.Background(), cheque, exchangeRate, deduction)
	if err != nil {
		t.Fatal(err)
	}

	if !verifiedWithFactory {
		t.Fatal("did not verify with factory")
	}

	if received.Cmp(cumulativePayout) != 0 {
		t.Fatalf("calculated wrong received cumulativePayout. wanted %d, got %d", cumulativePayout, received)
	}

	lastCheque, err := chequestore.LastCheque(chequebookAddress)
	if err != nil {
		t.Fatal(err)
	}

	if !cheque.Equal(lastCheque) {
		t.Fatalf("stored wrong cheque. wanted %v, got %v", cheque, lastCheque)
	}

	cheque = &chequebook.SignedCheque{
		Cheque: chequebook.Cheque{
			Beneficiary:      beneficiary,
			CumulativePayout: cumulativePayout2,
			Chequebook:       chequebookAddress,
		},
		Signature: sig,
	}

	verifiedWithFactory = false
	received, err = chequestore.ReceiveCheque(context.Background(), cheque, exchangeRate, deduction)
	if err != nil {
		t.Fatal(err)
	}

	if verifiedWithFactory {
		t.Fatal("needlessly verify with factory")
	}

	expectedReceived := big.NewInt(0).Sub(cumulativePayout2, cumulativePayout)
	if received.Cmp(expectedReceived) != 0 {
		t.Fatalf("calculated wrong received cumulativePayout. wanted %d, got %d", expectedReceived, received)
	}
}

func TestReceiveChequeInvalidBeneficiary(t *testing.T) {
	t.Parallel()

	store := storemock.NewStateStore()
	beneficiary := common.HexToAddress("0xffff")
	issuer := common.HexToAddress("0xbeee")
	cumulativePayout := big.NewInt(10)
	chequebookAddress := common.HexToAddress("0xeeee")
	sig := make([]byte, 65)
	chainID := int64(1)

	cheque := &chequebook.SignedCheque{
		Cheque: chequebook.Cheque{
			Beneficiary:      issuer,
			CumulativePayout: cumulativePayout,
			Chequebook:       chequebookAddress,
		},
		Signature: sig,
	}

	chequestore := chequebook.NewChequeStore(
		store,
		&factoryMock{},
		chainID,
		beneficiary,
		transactionmock.New(),
		nil,
	)

	_, err := chequestore.ReceiveCheque(context.Background(), cheque, cumulativePayout, big.NewInt(0))
	if err == nil {
		t.Fatal("accepted cheque with wrong beneficiary")
	}
	if !errors.Is(err, chequebook.ErrWrongBeneficiary) {
		t.Fatalf("wrong error. wanted %v, got %v", chequebook.ErrWrongBeneficiary, err)
	}
}

func TestReceiveChequeInvalidAmount(t *testing.T) {
	t.Parallel()

	store := storemock.NewStateStore()
	beneficiary := common.HexToAddress("0xffff")
	issuer := common.HexToAddress("0xbeee")
	cumulativePayout := big.NewInt(10)
	cumulativePayoutLower := big.NewInt(5)
	chequebookAddress := common.HexToAddress("0xeeee")
	sig := make([]byte, 65)
	chainID := int64(1)

	chequestore := chequebook.NewChequeStore(
		store,
		&factoryMock{
			verifyChequebook: func(ctx context.Context, address common.Address) error {
				return nil
			},
		},
		chainID,
		beneficiary,
		transactionmock.New(
			transactionmock.WithABICallSequence(
				transactionmock.ABICall(&chequebookABI, chequebookAddress, common.BytesToHash(issuer.Bytes()).Bytes(), "issuer"),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, cumulativePayout.FillBytes(make([]byte, 32)), "balance"),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, big.NewInt(0).FillBytes(make([]byte, 32)), "paidOut", beneficiary),
			),
		),
		func(c *chequebook.SignedCheque, cid int64) (common.Address, error) {
			return issuer, nil
		})

	_, err := chequestore.ReceiveCheque(context.Background(), &chequebook.SignedCheque{
		Cheque: chequebook.Cheque{
			Beneficiary:      beneficiary,
			CumulativePayout: cumulativePayout,
			Chequebook:       chequebookAddress,
		},
		Signature: sig,
	}, cumulativePayout, big.NewInt(0))
	if err != nil {
		t.Fatal(err)
	}

	_, err = chequestore.ReceiveCheque(context.Background(), &chequebook.SignedCheque{
		Cheque: chequebook.Cheque{
			Beneficiary:      beneficiary,
			CumulativePayout: cumulativePayoutLower,
			Chequebook:       chequebookAddress,
		},
		Signature: sig,
	}, cumulativePayout, big.NewInt(0))
	if err == nil {
		t.Fatal("accepted lower amount cheque")
	}
	if !errors.Is(err, chequebook.ErrChequeNotIncreasing) {
		t.Fatalf("wrong error. wanted %v, got %v", chequebook.ErrChequeNotIncreasing, err)
	}
}

func TestReceiveChequeInvalidChequebook(t *testing.T) {
	t.Parallel()

	store := storemock.NewStateStore()
	beneficiary := common.HexToAddress("0xffff")
	issuer := common.HexToAddress("0xbeee")
	cumulativePayout := big.NewInt(10)
	chequebookAddress := common.HexToAddress("0xeeee")
	sig := make([]byte, 65)
	chainID := int64(1)

	chequestore := chequebook.NewChequeStore(
		store,
		&factoryMock{
			verifyChequebook: func(ctx context.Context, address common.Address) error {
				return chequebook.ErrNotDeployedByFactory
			},
		},
		chainID,
		beneficiary,
		transactionmock.New(
			transactionmock.WithABICallSequence(
				transactionmock.ABICall(&chequebookABI, chequebookAddress, issuer.Bytes(), "issuer"),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, cumulativePayout.FillBytes(make([]byte, 32)), "balance"),
			),
		),
		func(c *chequebook.SignedCheque, cid int64) (common.Address, error) {
			return issuer, nil
		})

	_, err := chequestore.ReceiveCheque(context.Background(), &chequebook.SignedCheque{
		Cheque: chequebook.Cheque{
			Beneficiary:      beneficiary,
			CumulativePayout: cumulativePayout,
			Chequebook:       chequebookAddress,
		},
		Signature: sig,
	}, cumulativePayout, big.NewInt(0))
	if !errors.Is(err, chequebook.ErrNotDeployedByFactory) {
		t.Fatalf("wrong error. wanted %v, got %v", chequebook.ErrNotDeployedByFactory, err)
	}
}

func TestReceiveChequeInvalidSignature(t *testing.T) {
	t.Parallel()

	store := storemock.NewStateStore()
	beneficiary := common.HexToAddress("0xffff")
	issuer := common.HexToAddress("0xbeee")
	cumulativePayout := big.NewInt(10)
	chequebookAddress := common.HexToAddress("0xeeee")
	sig := make([]byte, 65)
	chainID := int64(1)

	chequestore := chequebook.NewChequeStore(
		store,
		&factoryMock{
			verifyChequebook: func(ctx context.Context, address common.Address) error {
				return nil
			},
		},
		chainID,
		beneficiary,
		transactionmock.New(
			transactionmock.WithABICallSequence(
				transactionmock.ABICall(&chequebookABI, chequebookAddress, common.BytesToHash(issuer.Bytes()).Bytes(), "issuer"),
			),
		),
		func(c *chequebook.SignedCheque, cid int64) (common.Address, error) {
			return common.Address{}, nil
		})

	_, err := chequestore.ReceiveCheque(context.Background(), &chequebook.SignedCheque{
		Cheque: chequebook.Cheque{
			Beneficiary:      beneficiary,
			CumulativePayout: cumulativePayout,
			Chequebook:       chequebookAddress,
		},
		Signature: sig,
	}, cumulativePayout, big.NewInt(0))
	if !errors.Is(err, chequebook.ErrChequeInvalid) {
		t.Fatalf("wrong error. wanted %v, got %v", chequebook.ErrChequeInvalid, err)
	}
}

func TestReceiveChequeInsufficientBalance(t *testing.T) {
	t.Parallel()

	store := storemock.NewStateStore()
	beneficiary := common.HexToAddress("0xffff")
	issuer := common.HexToAddress("0xbeee")
	cumulativePayout := big.NewInt(10)
	chequebookAddress := common.HexToAddress("0xeeee")
	sig := make([]byte, 65)
	chainID := int64(1)

	chequestore := chequebook.NewChequeStore(
		store,
		&factoryMock{
			verifyChequebook: func(ctx context.Context, address common.Address) error {
				return nil
			},
		},
		chainID,
		beneficiary,
		transactionmock.New(
			transactionmock.WithABICallSequence(
				transactionmock.ABICall(&chequebookABI, chequebookAddress, common.BytesToHash(issuer.Bytes()).Bytes(), "issuer"),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, new(big.Int).Sub(cumulativePayout, big.NewInt(1)).FillBytes(make([]byte, 32)), "balance"),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, big.NewInt(0).FillBytes(make([]byte, 32)), "paidOut", beneficiary),
			),
		),
		func(c *chequebook.SignedCheque, cid int64) (common.Address, error) {
			return issuer, nil
		})

	_, err := chequestore.ReceiveCheque(context.Background(), &chequebook.SignedCheque{
		Cheque: chequebook.Cheque{
			Beneficiary:      beneficiary,
			CumulativePayout: cumulativePayout,
			Chequebook:       chequebookAddress,
		},
		Signature: sig,
	}, cumulativePayout, big.NewInt(0))
	if !errors.Is(err, chequebook.ErrBouncingCheque) {
		t.Fatalf("wrong error. wanted %v, got %v", chequebook.ErrBouncingCheque, err)
	}
}

func TestReceiveChequeSufficientBalancePaidOut(t *testing.T) {
	t.Parallel()

	store := storemock.NewStateStore()
	beneficiary := common.HexToAddress("0xffff")
	issuer := common.HexToAddress("0xbeee")
	cumulativePayout := big.NewInt(10)
	chequebookAddress := common.HexToAddress("0xeeee")
	sig := make([]byte, 65)
	chainID := int64(1)

	chequestore := chequebook.NewChequeStore(
		store,
		&factoryMock{
			verifyChequebook: func(ctx context.Context, address common.Address) error {
				return nil
			},
		},
		chainID,
		beneficiary,
		transactionmock.New(
			transactionmock.WithABICallSequence(
				transactionmock.ABICall(&chequebookABI, chequebookAddress, common.BytesToHash(issuer.Bytes()).Bytes(), "issuer"),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, new(big.Int).Sub(cumulativePayout, big.NewInt(100)).FillBytes(make([]byte, 32)), "balance"),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, big.NewInt(0).FillBytes(make([]byte, 32)), "paidOut", beneficiary),
			),
		),
		func(c *chequebook.SignedCheque, cid int64) (common.Address, error) {
			return issuer, nil
		})

	_, err := chequestore.ReceiveCheque(context.Background(), &chequebook.SignedCheque{
		Cheque: chequebook.Cheque{
			Beneficiary:      beneficiary,
			CumulativePayout: cumulativePayout,
			Chequebook:       chequebookAddress,
		},
		Signature: sig,
	}, cumulativePayout, big.NewInt(0))
	if err != nil {
		t.Fatal(err)
	}
}

func TestReceiveChequeNotEnoughValue(t *testing.T) {
	t.Parallel()

	store := storemock.NewStateStore()
	beneficiary := common.HexToAddress("0xffff")
	issuer := common.HexToAddress("0xbeee")
	cumulativePayout := big.NewInt(100)
	chequebookAddress := common.HexToAddress("0xeeee")
	sig := make([]byte, 65)
	chainID := int64(1)
	exchangeRate := big.NewInt(101)
	deduction := big.NewInt(0)

	cheque := &chequebook.SignedCheque{
		Cheque: chequebook.Cheque{
			Beneficiary:      beneficiary,
			CumulativePayout: cumulativePayout,
			Chequebook:       chequebookAddress,
		},
		Signature: sig,
	}

	factory := &factoryMock{
		verifyChequebook: func(ctx context.Context, address common.Address) error {
			if address != chequebookAddress {
				t.Fatal("verifying wrong chequebook")
			}
			return nil
		},
	}

	chequestore := chequebook.NewChequeStore(
		store,
		factory,
		chainID,
		beneficiary,
		transactionmock.New(
			transactionmock.WithABICallSequence(
				transactionmock.ABICall(&chequebookABI, chequebookAddress, common.BytesToHash(issuer.Bytes()).Bytes(), "issuer"),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, cumulativePayout.FillBytes(make([]byte, 32)), "balance"),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, big.NewInt(0).FillBytes(make([]byte, 32)), "paidOut", beneficiary),
			),
		),
		func(c *chequebook.SignedCheque, cid int64) (common.Address, error) {
			if cid != chainID {
				t.Fatalf("recovery with wrong chain id. wanted %d, got %d", chainID, cid)
			}
			if !cheque.Equal(c) {
				t.Fatalf("recovery with wrong cheque. wanted %v, got %v", cheque, c)
			}
			return issuer, nil
		})

	_, err := chequestore.ReceiveCheque(context.Background(), cheque, exchangeRate, deduction)
	if !errors.Is(err, chequebook.ErrChequeValueTooLow) {
		t.Fatalf("got wrong error. wanted %v, got %v", chequebook.ErrChequeValueTooLow, err)
	}
}

func TestReceiveChequeNotEnoughValueAfterDeduction(t *testing.T) {
	t.Parallel()

	store := storemock.NewStateStore()
	beneficiary := common.HexToAddress("0xffff")
	issuer := common.HexToAddress("0xbeee")
	cumulativePayout := big.NewInt(100)
	chequebookAddress := common.HexToAddress("0xeeee")
	sig := make([]byte, 65)
	chainID := int64(1)

	// cheque needs to cover initial deduction (if applicable) plus one times the exchange rate
	// in order to amount to at least 1 accounting credit and be accepted
	// in this test cheque amount is just not enough to cover that therefore we expect

	exchangeRate := big.NewInt(100)
	deduction := big.NewInt(1)

	cheque := &chequebook.SignedCheque{
		Cheque: chequebook.Cheque{
			Beneficiary:      beneficiary,
			CumulativePayout: cumulativePayout,
			Chequebook:       chequebookAddress,
		},
		Signature: sig,
	}

	factory := &factoryMock{
		verifyChequebook: func(ctx context.Context, address common.Address) error {
			if address != chequebookAddress {
				t.Fatal("verifying wrong chequebook")
			}
			return nil
		},
	}

	chequestore := chequebook.NewChequeStore(
		store,
		factory,
		chainID,
		beneficiary,
		transactionmock.New(
			transactionmock.WithABICallSequence(
				transactionmock.ABICall(&chequebookABI, chequebookAddress, common.BytesToHash(issuer.Bytes()).Bytes(), "issuer"),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, cumulativePayout.FillBytes(make([]byte, 32)), "balance"),
				transactionmock.ABICall(&chequebookABI, chequebookAddress, big.NewInt(0).FillBytes(make([]byte, 32)), "paidOut", beneficiary),
			),
		),
		func(c *chequebook.SignedCheque, cid int64) (common.Address, error) {
			if cid != chainID {
				t.Fatalf("recovery with wrong chain id. wanted %d, got %d", chainID, cid)
			}
			if !cheque.Equal(c) {
				t.Fatalf("recovery with wrong cheque. wanted %v, got %v", cheque, c)
			}
			return issuer, nil
		})

	_, err := chequestore.ReceiveCheque(context.Background(), cheque, exchangeRate, deduction)
	if !errors.Is(err, chequebook.ErrChequeValueTooLow) {
		t.Fatalf("got wrong error. wanted %v, got %v", chequebook.ErrChequeValueTooLow, err)
	}
}
