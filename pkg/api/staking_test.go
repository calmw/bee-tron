// Copyright 2022 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api_test

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"testing"

	"github.com/calmw/bee-tron/pkg/bigint"
	"github.com/ethereum/go-ethereum/common"

	"github.com/calmw/bee-tron/pkg/api"
	"github.com/calmw/bee-tron/pkg/jsonhttp"
	"github.com/calmw/bee-tron/pkg/jsonhttp/jsonhttptest"
	"github.com/calmw/bee-tron/pkg/sctx"
	"github.com/calmw/bee-tron/pkg/storageincentives/staking"
	stakingContractMock "github.com/calmw/bee-tron/pkg/storageincentives/staking/mock"
)

func TestDepositStake(t *testing.T) {
	t.Parallel()

	txHash := common.HexToHash("0x1234")
	minStake := big.NewInt(100000000000000000).String()
	depositStake := func(amount string) string {
		return fmt.Sprintf("/stake/%s", amount)
	}

	t.Run("ok", func(t *testing.T) {
		t.Parallel()

		contract := stakingContractMock.New(
			stakingContractMock.WithDepositStake(func(ctx context.Context, stakedAmount *big.Int) (common.Hash, error) {
				return txHash, nil
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{StakingContract: contract})
		jsonhttptest.Request(t, ts, http.MethodPost, depositStake(minStake), http.StatusOK)
	})

	t.Run("with invalid stake amount", func(t *testing.T) {
		t.Parallel()

		invalidMinStake := big.NewInt(0).String()
		contract := stakingContractMock.New(
			stakingContractMock.WithDepositStake(func(ctx context.Context, stakedAmount *big.Int) (common.Hash, error) {
				return common.Hash{}, staking.ErrInsufficientStakeAmount
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{StakingContract: contract})
		jsonhttptest.Request(t, ts, http.MethodPost, depositStake(invalidMinStake), http.StatusBadRequest,
			jsonhttptest.WithExpectedJSONResponse(&jsonhttp.StatusResponse{Code: http.StatusBadRequest, Message: "insufficient stake amount"}))
	})

	t.Run("out of funds", func(t *testing.T) {
		t.Parallel()

		contract := stakingContractMock.New(
			stakingContractMock.WithDepositStake(func(ctx context.Context, stakedAmount *big.Int) (common.Hash, error) {
				return common.Hash{}, staking.ErrInsufficientFunds
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{StakingContract: contract})
		jsonhttptest.Request(t, ts, http.MethodPost, depositStake(minStake), http.StatusBadRequest)
		jsonhttptest.WithExpectedJSONResponse(&jsonhttp.StatusResponse{Code: http.StatusBadRequest, Message: "out of funds"})
	})

	t.Run("internal error", func(t *testing.T) {
		t.Parallel()

		contract := stakingContractMock.New(
			stakingContractMock.WithDepositStake(func(ctx context.Context, stakedAmount *big.Int) (common.Hash, error) {
				return common.Hash{}, fmt.Errorf("some error")
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{StakingContract: contract})
		jsonhttptest.Request(t, ts, http.MethodPost, depositStake(minStake), http.StatusInternalServerError)
		jsonhttptest.WithExpectedJSONResponse(&jsonhttp.StatusResponse{Code: http.StatusInternalServerError, Message: "cannot stake"})
	})

	t.Run("gas limit header", func(t *testing.T) {
		t.Parallel()

		contract := stakingContractMock.New(
			stakingContractMock.WithDepositStake(func(ctx context.Context, stakedAmount *big.Int) (common.Hash, error) {
				gasLimit := sctx.GetGasLimit(ctx)
				if gasLimit != 2000000 {
					t.Fatalf("want 2000000, got %d", gasLimit)
				}
				return txHash, nil
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{
			StakingContract: contract,
		})

		jsonhttptest.Request(t, ts, http.MethodPost, depositStake(minStake), http.StatusOK,
			jsonhttptest.WithRequestHeader(api.GasLimitHeader, "2000000"),
		)
	})
}

func TestGetStakeCommitted(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		t.Parallel()

		contract := stakingContractMock.New(
			stakingContractMock.WithGetStake(func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(1), nil
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{StakingContract: contract})
		jsonhttptest.Request(t, ts, http.MethodGet, "/stake", http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(&api.GetStakeResponse{StakedAmount: bigint.Wrap(big.NewInt(1))}))
	})

	t.Run("with error", func(t *testing.T) {
		t.Parallel()

		contractWithError := stakingContractMock.New(
			stakingContractMock.WithGetStake(func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(0), fmt.Errorf("get stake failed")
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{StakingContract: contractWithError})
		jsonhttptest.Request(t, ts, http.MethodGet, "/stake", http.StatusInternalServerError,
			jsonhttptest.WithExpectedJSONResponse(&jsonhttp.StatusResponse{Code: http.StatusInternalServerError, Message: "get staked amount failed"}))
	})
}

func TestGetStakeWithdrawable(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		t.Parallel()

		contract := stakingContractMock.New(
			stakingContractMock.WithGetStake(func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(1), nil
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{StakingContract: contract})
		jsonhttptest.Request(t, ts, http.MethodGet, "/stake/withdrawable", http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(&api.GetWithdrawableResponse{WithdrawableAmount: bigint.Wrap(big.NewInt(1))}))
	})

	t.Run("with error", func(t *testing.T) {
		t.Parallel()

		contractWithError := stakingContractMock.New(
			stakingContractMock.WithGetStake(func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(0), fmt.Errorf("get stake failed")
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{StakingContract: contractWithError})
		jsonhttptest.Request(t, ts, http.MethodGet, "/stake/withdrawable", http.StatusInternalServerError,
			jsonhttptest.WithExpectedJSONResponse(&jsonhttp.StatusResponse{Code: http.StatusInternalServerError, Message: "get staked amount failed"}))
	})
}

func Test_stakingDepositHandler_invalidInputs(t *testing.T) {
	t.Parallel()

	client, _, _, _ := newTestServer(t, testServerOptions{})

	tests := []struct {
		name   string
		amount string
		want   jsonhttp.StatusResponse
	}{{
		name:   "amount - invalid value",
		amount: "a",
		want: jsonhttp.StatusResponse{
			Code:    http.StatusBadRequest,
			Message: "invalid path params",
			Reasons: []jsonhttp.Reason{
				{
					Field: "amount",
					Error: "invalid value",
				},
			},
		},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jsonhttptest.Request(t, client, http.MethodPost, "/stake/"+tc.amount, tc.want.Code,
				jsonhttptest.WithExpectedJSONResponse(tc.want),
			)
		})
	}
}

func TestWithdrawStake(t *testing.T) {
	t.Parallel()

	txHash := common.HexToHash("0x1234")

	t.Run("ok", func(t *testing.T) {
		t.Parallel()

		contract := stakingContractMock.New(
			stakingContractMock.WithWithdrawStake(func(ctx context.Context) (common.Hash, error) {
				return txHash, nil
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{StakingContract: contract})
		jsonhttptest.Request(t, ts, http.MethodDelete, "/stake/withdrawable", http.StatusOK, jsonhttptest.WithExpectedJSONResponse(
			&api.StakeTransactionReponse{TxHash: txHash.String()}))
	})

	t.Run("with invalid stake amount", func(t *testing.T) {
		t.Parallel()

		contract := stakingContractMock.New(
			stakingContractMock.WithWithdrawStake(func(ctx context.Context) (common.Hash, error) {
				return common.Hash{}, staking.ErrInsufficientStake
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{StakingContract: contract})
		jsonhttptest.Request(t, ts, http.MethodDelete, "/stake/withdrawable", http.StatusBadRequest,
			jsonhttptest.WithExpectedJSONResponse(&jsonhttp.StatusResponse{Code: http.StatusBadRequest, Message: "insufficient stake to withdraw"}))
	})

	t.Run("internal error", func(t *testing.T) {
		t.Parallel()

		contract := stakingContractMock.New(
			stakingContractMock.WithWithdrawStake(func(ctx context.Context) (common.Hash, error) {
				return common.Hash{}, fmt.Errorf("some error")
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{StakingContract: contract})
		jsonhttptest.Request(t, ts, http.MethodDelete, "/stake/withdrawable", http.StatusInternalServerError)
		jsonhttptest.WithExpectedJSONResponse(&jsonhttp.StatusResponse{Code: http.StatusInternalServerError, Message: "cannot withdraw stake"})
	})

	t.Run("gas limit header", func(t *testing.T) {
		t.Parallel()

		contract := stakingContractMock.New(
			stakingContractMock.WithWithdrawStake(func(ctx context.Context) (common.Hash, error) {
				gasLimit := sctx.GetGasLimit(ctx)
				if gasLimit != 2000000 {
					t.Fatalf("want 2000000, got %d", gasLimit)
				}
				return txHash, nil
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{
			StakingContract: contract,
		})

		jsonhttptest.Request(t, ts, http.MethodDelete, "/stake/withdrawable", http.StatusOK,
			jsonhttptest.WithRequestHeader(api.GasLimitHeader, "2000000"),
		)
	})
}

func TestMigrateStake(t *testing.T) {
	t.Parallel()

	txHash := common.HexToHash("0x1234")

	t.Run("ok", func(t *testing.T) {
		t.Parallel()

		contract := stakingContractMock.New(
			stakingContractMock.WithMigrateStake(func(ctx context.Context) (common.Hash, error) {
				return txHash, nil
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{StakingContract: contract})
		jsonhttptest.Request(t, ts, http.MethodDelete, "/stake", http.StatusOK, jsonhttptest.WithExpectedJSONResponse(
			&api.StakeTransactionReponse{TxHash: txHash.String()}))
	})

	t.Run("with invalid stake amount", func(t *testing.T) {
		t.Parallel()

		contract := stakingContractMock.New(
			stakingContractMock.WithMigrateStake(func(ctx context.Context) (common.Hash, error) {
				return common.Hash{}, staking.ErrInsufficientStake
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{StakingContract: contract})
		jsonhttptest.Request(t, ts, http.MethodDelete, "/stake", http.StatusBadRequest,
			jsonhttptest.WithExpectedJSONResponse(&jsonhttp.StatusResponse{Code: http.StatusBadRequest, Message: "insufficient stake to migrate"}))
	})

	t.Run("internal error", func(t *testing.T) {
		t.Parallel()

		contract := stakingContractMock.New(
			stakingContractMock.WithMigrateStake(func(ctx context.Context) (common.Hash, error) {
				return common.Hash{}, fmt.Errorf("some error")
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{StakingContract: contract})
		jsonhttptest.Request(t, ts, http.MethodDelete, "/stake", http.StatusInternalServerError)
		jsonhttptest.WithExpectedJSONResponse(&jsonhttp.StatusResponse{Code: http.StatusInternalServerError, Message: "cannot withdraw stake"})
	})

	t.Run("gas limit header", func(t *testing.T) {
		t.Parallel()

		contract := stakingContractMock.New(
			stakingContractMock.WithMigrateStake(func(ctx context.Context) (common.Hash, error) {
				gasLimit := sctx.GetGasLimit(ctx)
				if gasLimit != 2000000 {
					t.Fatalf("want 2000000, got %d", gasLimit)
				}
				return txHash, nil
			}),
		)
		ts, _, _, _ := newTestServer(t, testServerOptions{
			StakingContract: contract,
		})

		jsonhttptest.Request(t, ts, http.MethodDelete, "/stake", http.StatusOK,
			jsonhttptest.WithRequestHeader(api.GasLimitHeader, "2000000"),
		)
	})
}
