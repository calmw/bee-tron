// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libp2p_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/calmw/bee-tron/pkg/p2p"
	"github.com/calmw/bee-tron/pkg/p2p/libp2p"
	"github.com/calmw/bee-tron/pkg/spinlock"
	libp2pm "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	protocol "github.com/libp2p/go-libp2p/core/protocol"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	swarmt "github.com/libp2p/go-libp2p/p2p/net/swarm/testing"
	"github.com/multiformats/go-multistream"
)

func TestNewStream(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s1, overlay1 := newService(t, 1, libp2pServiceOpts{libp2pOpts: libp2p.Options{
		FullNode: true,
	}})

	s2, _ := newService(t, 1, libp2pServiceOpts{})

	if err := s1.AddProtocol(newTestProtocol(func(_ context.Context, p p2p.Peer, _ p2p.Stream) error {
		return nil
	})); err != nil {
		t.Fatal(err)
	}

	addr := serviceUnderlayAddress(t, s1)

	if _, err := s2.Connect(ctx, addr); err != nil {
		t.Fatal(err)
	}

	stream, err := s2.NewStream(ctx, overlay1, nil, testProtocolName, testProtocolVersion, testStreamName)
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestNewStream_OnlyFull tests that the handler gets the full
// node information communicated correctly.
func TestNewStream_OnlyFull(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s1, overlay1 := newService(t, 1, libp2pServiceOpts{libp2pOpts: libp2p.Options{
		FullNode: true,
	}})

	s2, _ := newService(t, 1, libp2pServiceOpts{libp2pOpts: libp2p.Options{
		FullNode: true,
	}})

	if err := s1.AddProtocol(newTestProtocol(func(_ context.Context, p p2p.Peer, _ p2p.Stream) error {
		if !p.FullNode {
			t.Error("expected full node")
		}
		return nil
	})); err != nil {
		t.Fatal(err)
	}

	addr := serviceUnderlayAddress(t, s1)

	if _, err := s2.Connect(ctx, addr); err != nil {
		t.Fatal(err)
	}

	stream, err := s2.NewStream(ctx, overlay1, nil, testProtocolName, testProtocolVersion, testStreamName)
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestNewStream_Mixed tests that the handler gets the full
// node information communicated correctly for light node
func TestNewStream_Mixed(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s1, overlay1 := newService(t, 1, libp2pServiceOpts{libp2pOpts: libp2p.Options{
		FullNode: true,
	}})

	s2, _ := newService(t, 1, libp2pServiceOpts{})

	if err := s1.AddProtocol(newTestProtocol(func(_ context.Context, p p2p.Peer, _ p2p.Stream) error {
		if p.FullNode {
			t.Error("expected light node")
		}
		return nil
	})); err != nil {
		t.Fatal(err)
	}

	addr := serviceUnderlayAddress(t, s1)

	if _, err := s2.Connect(ctx, addr); err != nil {
		t.Fatal(err)
	}

	stream, err := s2.NewStream(ctx, overlay1, nil, testProtocolName, testProtocolVersion, testStreamName)
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestNewStreamMulti is a regression test to see that we trigger
// the right handler when multiple streams are registered under
// a single protocol.
func TestNewStreamMulti(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s1, overlay1 := newService(t, 1, libp2pServiceOpts{libp2pOpts: libp2p.Options{
		FullNode: true,
	}})

	var (
		h1calls, h2calls int32
		h1               = func(_ context.Context, p p2p.Peer, s p2p.Stream) error {
			defer s.Close()
			_ = atomic.AddInt32(&h1calls, 1)
			return nil
		}
		h2 = func(_ context.Context, p p2p.Peer, s p2p.Stream) error {
			defer s.Close()
			_ = atomic.AddInt32(&h2calls, 1)
			return nil
		}
	)
	s2, _ := newService(t, 1, libp2pServiceOpts{})

	if err := s1.AddProtocol(newTestMultiProtocol(h1, h2)); err != nil {
		t.Fatal(err)
	}

	addr := serviceUnderlayAddress(t, s1)

	if _, err := s2.Connect(ctx, addr); err != nil {
		t.Fatal(err)
	}

	stream, err := s2.NewStream(ctx, overlay1, nil, testProtocolName, testProtocolVersion, testStreamName)
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.FullClose(); err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&h1calls) != 1 {
		t.Fatal("handler should have been called but wasn't")
	}
	if atomic.LoadInt32(&h2calls) > 0 {
		t.Fatal("handler should not have been called")
	}
}

func TestNewStream_errNotSupported(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s1, overlay1 := newService(t, 1, libp2pServiceOpts{libp2pOpts: libp2p.Options{
		FullNode: true,
	}})

	s2, _ := newService(t, 1, libp2pServiceOpts{})

	addr := serviceUnderlayAddress(t, s1)

	// connect nodes
	if _, err := s2.Connect(ctx, addr); err != nil {
		t.Fatal(err)
	}

	// test for missing protocol
	_, err := s2.NewStream(ctx, overlay1, nil, testProtocolName, testProtocolVersion, testStreamName)
	expectErrNotSupported(t, err)

	// add protocol
	if err := s1.AddProtocol(newTestProtocol(func(_ context.Context, _ p2p.Peer, _ p2p.Stream) error {
		return nil
	})); err != nil {
		t.Fatal(err)
	}

	// test for incorrect protocol name
	_, err = s2.NewStream(ctx, overlay1, nil, testProtocolName+"invalid", testProtocolVersion, testStreamName)
	expectErrNotSupported(t, err)

	// test for incorrect stream name
	_, err = s2.NewStream(ctx, overlay1, nil, testProtocolName, testProtocolVersion, testStreamName+"invalid")
	expectErrNotSupported(t, err)
}

func TestNewStream_semanticVersioning(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s1, overlay1 := newService(t, 1, libp2pServiceOpts{libp2pOpts: libp2p.Options{
		FullNode: true,
	}})

	s2, _ := newService(t, 1, libp2pServiceOpts{})

	addr := serviceUnderlayAddress(t, s1)

	if _, err := s2.Connect(ctx, addr); err != nil {
		t.Fatal(err)
	}

	if err := s1.AddProtocol(newTestProtocol(func(_ context.Context, _ p2p.Peer, _ p2p.Stream) error {
		return nil
	})); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		version   string
		supported bool
	}{
		{version: "0", supported: false},
		{version: "1", supported: false},
		{version: "2", supported: false},
		{version: "3", supported: false},
		{version: "4", supported: false},
		{version: "a", supported: false},
		{version: "invalid", supported: false},
		{version: "0.0.0", supported: false},
		{version: "0.1.0", supported: false},
		{version: "1.0.0", supported: false},
		{version: "2.0.0", supported: true},
		{version: "2.2.0", supported: true},
		{version: "2.3.0", supported: true},
		{version: "2.3.1", supported: true},
		{version: "2.3.4", supported: true},
		{version: "2.3.5", supported: true},
		{version: "2.3.5-beta", supported: true},
		{version: "2.3.5+beta", supported: true},
		{version: "2.3.6", supported: true},
		{version: "2.3.6-beta", supported: true},
		{version: "2.3.6+beta", supported: true},
		{version: "2.4.0", supported: false},
		{version: "3.0.0", supported: false},
	} {
		_, err := s2.NewStream(ctx, overlay1, nil, testProtocolName, tc.version, testStreamName)
		if tc.supported {
			if err != nil {
				t.Fatal(err)
			}
		} else {
			expectErrNotSupported(t, err)
		}
	}
}

func TestDisconnectError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s1, overlay1 := newService(t, 1, libp2pServiceOpts{libp2pOpts: libp2p.Options{
		FullNode: true,
	}})

	s2, overlay2 := newService(t, 1, libp2pServiceOpts{})

	if err := s1.AddProtocol(newTestProtocol(func(_ context.Context, _ p2p.Peer, _ p2p.Stream) error {
		return p2p.NewDisconnectError(errors.New("test error"))
	})); err != nil {
		t.Fatal(err)
	}

	addr := serviceUnderlayAddress(t, s1)

	if _, err := s2.Connect(ctx, addr); err != nil {
		t.Fatal(err)
	}

	expectPeers(t, s1, overlay2)

	// error is not checked as opening a new stream should cause disconnect from s1 which is async and can make errors in newStream function
	// it is important to validate that disconnect will happen after NewStream()
	_, _ = s2.NewStream(ctx, overlay1, nil, testProtocolName, testProtocolVersion, testStreamName)
	expectPeersEventually(t, s1)
}

func TestConnectDisconnectEvents(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s1, overlay1 := newService(t, 1, libp2pServiceOpts{libp2pOpts: libp2p.Options{
		FullNode: true,
	}})

	s2, _ := newService(t, 1, libp2pServiceOpts{})
	testProtocol := newTestProtocol(func(_ context.Context, _ p2p.Peer, _ p2p.Stream) error {
		return nil
	})

	cinCount, coutCount, dinCount, doutCount := 0, 0, 0, 0
	var countMU sync.Mutex

	testProtocol.ConnectIn = func(c context.Context, p p2p.Peer) error {
		countMU.Lock()
		cinCount++
		countMU.Unlock()
		return nil
	}

	testProtocol.ConnectOut = func(c context.Context, p p2p.Peer) error {
		countMU.Lock()
		coutCount++
		countMU.Unlock()
		return nil
	}

	testProtocol.DisconnectIn = func(p p2p.Peer) error {
		countMU.Lock()
		dinCount++
		countMU.Unlock()
		return nil
	}

	testProtocol.DisconnectOut = func(p p2p.Peer) error {
		countMU.Lock()
		doutCount++
		countMU.Unlock()
		return nil
	}

	if err := s1.AddProtocol(testProtocol); err != nil {
		t.Fatal(err)
	}

	if err := s2.AddProtocol(testProtocol); err != nil {
		t.Fatal(err)
	}

	addr := serviceUnderlayAddress(t, s1)

	if _, err := s2.Connect(ctx, addr); err != nil {
		t.Fatal(err)
	}

	expectCounter(t, &cinCount, 1, &countMU)
	expectCounter(t, &coutCount, 1, &countMU)
	expectCounter(t, &dinCount, 0, &countMU)
	expectCounter(t, &doutCount, 0, &countMU)

	if err := s2.Disconnect(overlay1, "test disconnect"); err != nil {
		t.Fatal(err)
	}

	cinCount = 0
	coutCount = 0

	expectCounter(t, &cinCount, 0, &countMU)
	expectCounter(t, &coutCount, 0, &countMU)
	expectCounter(t, &dinCount, 1, &countMU)
	expectCounter(t, &doutCount, 1, &countMU)

}

func TestPing(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	s1, _ := newService(t, 1, libp2pServiceOpts{
		libp2pOpts: libp2p.WithHostFactory(
			func(...libp2pm.Option) (host.Host, error) {
				return bhost.NewHost(swarmt.GenSwarm(t), &bhost.HostOpts{EnablePing: true})
			},
		),
	})

	s2, _ := newService(t, 1, libp2pServiceOpts{
		libp2pOpts: libp2p.WithHostFactory(
			func(...libp2pm.Option) (host.Host, error) {
				host, err := bhost.NewHost(swarmt.GenSwarm(t), &bhost.HostOpts{EnablePing: true})
				if err != nil {
					t.Fatalf("start host: %v", err)
				}
				host.Start()
				return host, nil
			},
		),
	})

	addr := serviceUnderlayAddress(t, s1)

	if _, err := s2.Ping(ctx, addr); err != nil {
		t.Fatal(err)
	}
}

const (
	testProtocolName     = "testing"
	testProtocolVersion  = "2.3.4"
	testStreamName       = "messages"
	testSecondStreamName = "cookies"
)

func newTestProtocol(h p2p.HandlerFunc) p2p.ProtocolSpec {
	return p2p.ProtocolSpec{
		Name:    testProtocolName,
		Version: testProtocolVersion,
		StreamSpecs: []p2p.StreamSpec{
			{
				Name:    testStreamName,
				Handler: h,
			},
		},
	}
}

func newTestMultiProtocol(h1, h2 p2p.HandlerFunc) p2p.ProtocolSpec {
	return p2p.ProtocolSpec{
		Name:    testProtocolName,
		Version: testProtocolVersion,
		StreamSpecs: []p2p.StreamSpec{
			{
				Name:    testStreamName,
				Handler: h1,
			},
			{
				Name:    testSecondStreamName,
				Handler: h2,
			},
		},
	}
}

func expectErrNotSupported(t *testing.T, err error) {
	t.Helper()
	if e := (*p2p.IncompatibleStreamError)(nil); !errors.As(err, &e) {
		t.Fatalf("got error %v, want %T", err, e)
	}
	var e2 multistream.ErrNotSupported[protocol.ID]
	if !errors.As(err, &e2) {
		t.Fatalf("got error %v, want %v", err, &e2)
	}
}

func expectCounter(t *testing.T, c *int, expected int, mtx *sync.Mutex) {
	t.Helper()

	err := spinlock.Wait(time.Second, func() bool {
		mtx.Lock()
		defer mtx.Unlock()
		return *c == expected
	})
	if err != nil {
		t.Fatal("timed out waiting for counter to be set")
	}
}
