// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package swap

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/calmw/bee-tron/pkg/log"
	"github.com/calmw/bee-tron/pkg/postage/postagecontract"
	"github.com/calmw/bee-tron/pkg/settlement"
	"github.com/calmw/bee-tron/pkg/settlement/swap/chequebook"
	"github.com/calmw/bee-tron/pkg/settlement/swap/swapprotocol"
	"github.com/calmw/bee-tron/pkg/storage"
	"github.com/calmw/bee-tron/pkg/swarm"
	"github.com/ethereum/go-ethereum/common"
)

// loggerName is the tree path name of the logger for this package.
const loggerName = "swap"

var (
	// ErrWrongChequebook is the error if a peer uses a different chequebook from before.
	ErrWrongChequebook = errors.New("wrong chequebook")
	// ErrUnknownBeneficary is the error if a peer has never announced a beneficiary.
	ErrUnknownBeneficary = errors.New("unknown beneficiary for peer")
	// ErrChequeValueTooLow is the error a peer issued a cheque not covering 1 accounting credit
	ErrChequeValueTooLow = errors.New("cheque value too low")
	ErrNoChequebook      = errors.New("no chequebook")
)

type Interface interface {
	settlement.Interface
	// LastSentCheque returns the last sent cheque for the peer
	LastSentCheque(peer swarm.Address) (*chequebook.SignedCheque, error)
	// LastSentCheques returns the list of last sent cheques for all peers
	LastSentCheques() (map[string]*chequebook.SignedCheque, error)
	// LastReceivedCheque returns the last received cheque for the peer
	LastReceivedCheque(peer swarm.Address) (*chequebook.SignedCheque, error)
	// LastReceivedCheques returns the list of last received cheques for all peers
	LastReceivedCheques() (map[string]*chequebook.SignedCheque, error)
	// CashCheque sends a cashing transaction for the last cheque of the peer
	CashCheque(ctx context.Context, peer swarm.Address) (common.Hash, error)
	// CashoutStatus gets the status of the latest cashout transaction for the peers chequebook
	CashoutStatus(ctx context.Context, peer swarm.Address) (*chequebook.CashoutStatus, error)
}

// Service is the implementation of the swap settlement layer.
type Service struct {
	proto          swapprotocol.Interface
	logger         log.Logger
	store          storage.StateStorer
	accounting     settlement.Accounting
	metrics        metrics
	chequebook     chequebook.Service
	chequeStore    chequebook.ChequeStore
	cashout        chequebook.CashoutService
	addressbook    Addressbook
	networkID      uint64
	cashoutAddress common.Address
}

// New creates a new swap Service.
func New(proto swapprotocol.Interface, logger log.Logger, store storage.StateStorer, chequebook chequebook.Service, chequeStore chequebook.ChequeStore, addressbook Addressbook, networkID uint64, cashout chequebook.CashoutService, accounting settlement.Accounting, cashoutAddress common.Address) *Service {
	return &Service{
		proto:          proto,
		logger:         logger.WithName(loggerName).Register(),
		store:          store,
		metrics:        newMetrics(),
		chequebook:     chequebook,
		chequeStore:    chequeStore,
		addressbook:    addressbook,
		networkID:      networkID,
		cashout:        cashout,
		accounting:     accounting,
		cashoutAddress: cashoutAddress,
	}
}

// ReceiveCheque is called by the swap protocol if a cheque is received.
func (s *Service) ReceiveCheque(ctx context.Context, peer swarm.Address, cheque *chequebook.SignedCheque, exchangeRate, deduction *big.Int) (err error) {
	// check this is the same chequebook for this peer as previously
	expectedChequebook, known, err := s.addressbook.Chequebook(peer)
	if err != nil {
		return err
	}
	if known && expectedChequebook != cheque.Chequebook {
		return ErrWrongChequebook
	}

	receivedAmount, err := s.chequeStore.ReceiveCheque(ctx, cheque, exchangeRate, deduction)
	if err != nil {
		s.metrics.ChequesRejected.Inc()
		return fmt.Errorf("rejecting cheque: %w", err)
	}

	if deduction.Cmp(big.NewInt(0)) > 0 {
		err = s.addressbook.AddDeductionFor(peer)
		if err != nil {
			return err
		}
	}

	decreasedAmount := new(big.Int).Sub(receivedAmount, deduction)
	amount := new(big.Int).Div(decreasedAmount, exchangeRate)

	if !known {
		err = s.addressbook.PutChequebook(peer, cheque.Chequebook)
		if err != nil {
			return err
		}
	}

	tot, _ := big.NewFloat(0).SetInt(receivedAmount).Float64()
	s.metrics.TotalReceived.Add(tot)
	s.metrics.ChequesReceived.Inc()

	return s.accounting.NotifyPaymentReceived(peer, amount)
}

// Pay initiates a payment to the given peer
func (s *Service) Pay(ctx context.Context, peer swarm.Address, amount *big.Int) {
	var err error
	defer func() {
		if err != nil {
			s.accounting.NotifyPaymentSent(peer, amount, err)
		}
	}()
	if s.chequebook == nil {
		err = ErrNoChequebook
		return
	}
	beneficiary, known, err := s.addressbook.Beneficiary(peer)
	if err != nil {
		return
	}
	if !known {
		err = ErrUnknownBeneficary
		return
	}

	balance, err := s.proto.EmitCheque(ctx, peer, beneficiary, amount, s.chequebook.Issue)

	if err != nil {
		return
	}

	bal, _ := big.NewFloat(0).SetInt(balance).Float64()
	s.metrics.AvailableBalance.Set(bal)
	s.accounting.NotifyPaymentSent(peer, amount, nil)
	amountFloat, _ := big.NewFloat(0).SetInt(amount).Float64()
	s.metrics.TotalSent.Add(amountFloat)
	s.metrics.ChequesSent.Inc()
}

func (s *Service) SetAccounting(accounting settlement.Accounting) {
	s.accounting = accounting
}

// TotalSent returns the total amount sent to a peer
func (s *Service) TotalSent(peer swarm.Address) (totalSent *big.Int, err error) {
	beneficiary, known, err := s.addressbook.Beneficiary(peer)
	if err != nil {
		return nil, err
	}
	if !known {
		return nil, settlement.ErrPeerNoSettlements
	}
	if s.chequebook == nil {
		return big.NewInt(0), nil
	}
	cheque, err := s.chequebook.LastCheque(beneficiary)
	if err != nil {
		if errors.Is(err, chequebook.ErrNoCheque) {
			return nil, settlement.ErrPeerNoSettlements
		}
		return nil, err
	}
	return cheque.CumulativePayout, nil
}

// TotalReceived returns the total amount received from a peer
func (s *Service) TotalReceived(peer swarm.Address) (totalReceived *big.Int, err error) {
	chequebookAddress, known, err := s.addressbook.Chequebook(peer)
	if err != nil {
		return nil, err
	}
	if !known {
		return nil, settlement.ErrPeerNoSettlements
	}

	cheque, err := s.chequeStore.LastCheque(chequebookAddress)
	if err != nil {
		if errors.Is(err, chequebook.ErrNoCheque) {
			return nil, settlement.ErrPeerNoSettlements
		}
		return nil, err
	}
	return cheque.CumulativePayout, nil
}

// SettlementsSent returns sent settlements for each individual known peer
func (s *Service) SettlementsSent() (map[string]*big.Int, error) {
	result := make(map[string]*big.Int)
	if s.chequebook == nil {
		return result, nil
	}
	cheques, err := s.chequebook.LastCheques()
	if err != nil {
		return nil, err
	}

	for beneficiary, cheque := range cheques {
		peer, known, err := s.addressbook.BeneficiaryPeer(beneficiary)
		if err != nil {
			return nil, err
		}
		if !known {
			continue
		}
		result[peer.String()] = cheque.CumulativePayout
	}

	return result, nil
}

// SettlementsReceived returns received settlements for each individual known peer.
func (s *Service) SettlementsReceived() (map[string]*big.Int, error) {
	result := make(map[string]*big.Int)
	cheques, err := s.chequeStore.LastCheques()
	if err != nil {
		return nil, err
	}

	for chequebook, cheque := range cheques {
		peer, known, err := s.addressbook.ChequebookPeer(chequebook)
		if err != nil {
			return nil, err
		}
		if !known {
			continue
		}
		result[peer.String()] = cheque.CumulativePayout
	}
	return result, err
}

// Handshake is called by the swap protocol when a handshake is received.
func (s *Service) Handshake(peer swarm.Address, beneficiary common.Address) error {
	loggerV1 := s.logger.V(1).Register()

	oldPeer, known, err := s.addressbook.BeneficiaryPeer(beneficiary)
	if err != nil {
		return err
	}
	if known && !peer.Equal(oldPeer) {
		s.logger.Debug("migrating swap addresses", "old_peer_address", oldPeer, "new_peer_address", peer)
		return s.addressbook.MigratePeer(oldPeer, peer)
	}

	_, known, err = s.addressbook.Beneficiary(peer)
	if err != nil {
		return err
	}
	if !known {
		loggerV1.Debug("initial swap handshake", "peer_address", peer, "beneficiary_address", beneficiary)
		return s.addressbook.PutBeneficiary(peer, beneficiary)
	}

	return nil
}

// LastSentCheque returns the last sent cheque for the peer
func (s *Service) LastSentCheque(peer swarm.Address) (*chequebook.SignedCheque, error) {

	common, known, err := s.addressbook.Beneficiary(peer)

	if err != nil {
		return nil, err
	}

	if !known {
		return nil, chequebook.ErrNoCheque
	}

	if s.chequebook == nil {
		return nil, ErrNoChequebook
	}

	return s.chequebook.LastCheque(common)
}

// LastReceivedCheque returns the last received cheque for the peer
func (s *Service) LastReceivedCheque(peer swarm.Address) (*chequebook.SignedCheque, error) {

	common, known, err := s.addressbook.Chequebook(peer)

	if err != nil {
		return nil, err
	}

	if !known {
		return nil, chequebook.ErrNoCheque
	}

	return s.chequeStore.LastCheque(common)
}

// LastSentCheques returns the list of last sent cheques for all peers
func (s *Service) LastSentCheques() (map[string]*chequebook.SignedCheque, error) {
	if s.chequebook == nil {
		return nil, ErrNoChequebook
	}
	lastcheques, err := s.chequebook.LastCheques()
	if err != nil {
		return nil, err
	}

	resultmap := make(map[string]*chequebook.SignedCheque, len(lastcheques))

	for i, j := range lastcheques {
		addr, known, err := s.addressbook.BeneficiaryPeer(i)
		if err == nil && known {
			resultmap[addr.String()] = j
		}
	}

	return resultmap, nil
}

// LastReceivedCheques returns the list of last received cheques for all peers
func (s *Service) LastReceivedCheques() (map[string]*chequebook.SignedCheque, error) {
	lastcheques, err := s.chequeStore.LastCheques()
	if err != nil {
		return nil, err
	}

	resultmap := make(map[string]*chequebook.SignedCheque, len(lastcheques))

	for i, j := range lastcheques {
		addr, known, err := s.addressbook.ChequebookPeer(i)
		if err == nil && known {
			resultmap[addr.String()] = j
		}
	}

	return resultmap, nil
}

// CashCheque sends a cashing transaction for the last cheque of the peer
func (s *Service) CashCheque(ctx context.Context, peer swarm.Address) (common.Hash, error) {
	chequebookAddress, known, err := s.addressbook.Chequebook(peer)
	if err != nil {
		return common.Hash{}, err
	}
	if !known {
		return common.Hash{}, chequebook.ErrNoCheque
	}
	return s.cashout.CashCheque(ctx, chequebookAddress, s.cashoutAddress)
}

// CashoutStatus gets the status of the latest cashout transaction for the peers chequebook
func (s *Service) CashoutStatus(ctx context.Context, peer swarm.Address) (*chequebook.CashoutStatus, error) {
	chequebookAddress, known, err := s.addressbook.Chequebook(peer)
	if err != nil {
		return nil, err
	}
	if !known {
		return nil, chequebook.ErrNoCheque
	}
	return s.cashout.CashoutStatus(ctx, chequebookAddress)
}

func (s *Service) GetDeductionForPeer(peer swarm.Address) (bool, error) {
	return s.addressbook.GetDeductionFor(peer)
}

func (s *Service) GetDeductionByPeer(peer swarm.Address) (bool, error) {
	return s.addressbook.GetDeductionBy(peer)
}

func (s *Service) AddDeductionByPeer(peer swarm.Address) error {
	return s.addressbook.AddDeductionBy(peer)
}

type NoOpSwap struct {
}

func (*NoOpSwap) TotalSent(peer swarm.Address) (totalSent *big.Int, err error) {
	return nil, postagecontract.ErrChainDisabled
}

// TotalReceived returns the total amount received from a peer
func (*NoOpSwap) TotalReceived(peer swarm.Address) (totalSent *big.Int, err error) {
	return nil, postagecontract.ErrChainDisabled
}

// SettlementsSent returns sent settlements for each individual known peer
func (*NoOpSwap) SettlementsSent() (map[string]*big.Int, error) {
	return nil, postagecontract.ErrChainDisabled
}

// SettlementsReceived returns received settlements for each individual known peer
func (*NoOpSwap) SettlementsReceived() (map[string]*big.Int, error) {
	return nil, postagecontract.ErrChainDisabled
}

func (*NoOpSwap) LastSentCheque(peer swarm.Address) (*chequebook.SignedCheque, error) {
	return nil, postagecontract.ErrChainDisabled
}

// LastSentCheques returns the list of last sent cheques for all peers
func (*NoOpSwap) LastSentCheques() (map[string]*chequebook.SignedCheque, error) {
	return nil, postagecontract.ErrChainDisabled
}

// LastReceivedCheque returns the last received cheque for the peer
func (*NoOpSwap) LastReceivedCheque(peer swarm.Address) (*chequebook.SignedCheque, error) {
	return nil, postagecontract.ErrChainDisabled
}

// LastReceivedCheques returns the list of last received cheques for all peers
func (*NoOpSwap) LastReceivedCheques() (map[string]*chequebook.SignedCheque, error) {
	return nil, postagecontract.ErrChainDisabled
}

// CashCheque sends a cashing transaction for the last cheque of the peer
func (*NoOpSwap) CashCheque(ctx context.Context, peer swarm.Address) (common.Hash, error) {
	return common.Hash{}, postagecontract.ErrChainDisabled
}

// CashoutStatus gets the status of the latest cashout transaction for the peers chequebook
func (*NoOpSwap) CashoutStatus(ctx context.Context, peer swarm.Address) (*chequebook.CashoutStatus, error) {
	return nil, postagecontract.ErrChainDisabled
}
