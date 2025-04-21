// Copyright 2021 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package priceoracle_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/calmw/bee-tron/pkg/log"
	"github.com/calmw/bee-tron/pkg/settlement/swap/priceoracle"
	transactionmock "github.com/calmw/bee-tron/pkg/transaction/mock"
	"github.com/calmw/bee-tron/pkg/util/abiutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethersphere/go-price-oracle-abi/priceoracleabi"
)

var (
	priceOracleABI = abiutil.MustParseABI(priceoracleabi.PriceOracleABIv0_2_0)
)

func TestExchangeGetPrice(t *testing.T) {
	t.Parallel()

	priceOracleAddress := common.HexToAddress("0xabcd")

	expectedPrice := big.NewInt(100)
	expectedDeduce := big.NewInt(200)

	result := make([]byte, 64)
	expectedPrice.FillBytes(result[0:32])
	expectedDeduce.FillBytes(result[32:64])

	ex := priceoracle.New(
		log.Noop,
		priceOracleAddress,
		transactionmock.New(
			transactionmock.WithABICall(
				&priceOracleABI,
				priceOracleAddress,
				result,
				"getPrice",
			),
		),
		1,
	)

	price, deduce, err := ex.GetPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if expectedPrice.Cmp(price) != 0 {
		t.Fatalf("got wrong price. wanted %d, got %d", expectedPrice, price)
	}

	if expectedDeduce.Cmp(deduce) != 0 {
		t.Fatalf("got wrong deduce. wanted %d, got %d", expectedDeduce, deduce)
	}
}
