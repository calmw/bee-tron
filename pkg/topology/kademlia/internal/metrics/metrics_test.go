// Copyright 2021 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics_test

import (
	"testing"
	"time"

	"github.com/calmw/bee-tron/pkg/p2p"
	"github.com/calmw/bee-tron/pkg/shed"
	"github.com/calmw/bee-tron/pkg/swarm"
	"github.com/calmw/bee-tron/pkg/topology/kademlia/internal/metrics"
	"github.com/calmw/bee-tron/pkg/util/testutil"
	"github.com/google/go-cmp/cmp"
)

func snapshot(t *testing.T, mc *metrics.Collector, sst time.Time, addr swarm.Address) *metrics.Snapshot {
	t.Helper()

	ss := mc.Snapshot(sst, addr)
	if have, want := len(ss), 1; have != want {
		t.Fatalf("Snapshot(%q, ...): length mismatch: have: %d; want: %d", addr, have, want)
	}
	cs, ok := ss[addr.ByteString()]
	if !ok {
		t.Fatalf("Snapshot(%q, ...): missing peer metrics", addr)
	}
	return cs
}

func TestPeerMetricsCollector(t *testing.T) {
	t.Parallel()

	db, err := shed.NewDB("", nil)
	if err != nil {
		t.Fatal(err)
	}
	testutil.CleanupCloser(t, db)

	mc, err := metrics.NewCollector(db)
	if err != nil {
		t.Fatal(err)
	}

	var (
		addr = swarm.MustParseHexAddress("0123456789")

		t1 = time.Now()               // Login time.
		t2 = t1.Add(10 * time.Second) // Snapshot time.
		t3 = t2.Add(55 * time.Second) // Logout time.
		t4 = 10 * time.Millisecond    // Latency duration.
		t5 = 100 * time.Millisecond   // Next latency duration sample.
	)

	// Inc session conn retry.
	mc.Record(addr, metrics.IncSessionConnectionRetry())
	ss := snapshot(t, mc, t2, addr)
	if have, want := ss.SessionConnectionRetry, uint64(1); have != want {
		t.Fatalf("Snapshot(%q, ...): session connection retry counter mismatch: have %d; want %d", addr, have, want)
	}

	// Login.
	mc.Record(addr, metrics.PeerLogIn(t1, metrics.PeerConnectionDirectionInbound))
	ss = snapshot(t, mc, t2, addr)
	if have, want := ss.LastSeenTimestamp, t1.UnixNano(); have != want {
		t.Fatalf("Snapshot(%q, ...): last seen counter mismatch: have %d; want %d", addr, have, want)
	}
	if have, want := ss.SessionConnectionDirection, metrics.PeerConnectionDirectionInbound; have != want {
		t.Fatalf("Snapshot(%q, ...): session connection direction counter mismatch: have %q; want %q", addr, have, want)
	}
	if have, want := ss.SessionConnectionDuration, t2.Sub(t1); have != want {
		t.Fatalf("Snapshot(%q, ...): session connection duration counter mismatch: have %s; want %s", addr, have, want)
	}
	if have, want := ss.ConnectionTotalDuration, t2.Sub(t1); have != want {
		t.Fatalf("Snapshot(%q, ...): connection total duration counter mismatch: have %s; want %s", addr, have, want)
	}

	// Login when already logged in.
	mc.Record(addr, metrics.PeerLogIn(t1.Add(1*time.Second), metrics.PeerConnectionDirectionOutbound))
	ss = snapshot(t, mc, t2, addr)
	if have, want := ss.LastSeenTimestamp, t1.UnixNano(); have != want {
		t.Fatalf("Snapshot(%q, ...): last seen counter mismatch: have %d; want %d", addr, have, want)
	}
	if have, want := ss.SessionConnectionDirection, metrics.PeerConnectionDirectionInbound; have != want {
		t.Fatalf("Snapshot(%q, ...): session connection direction counter mismatch: have %q; want %q", addr, have, want)
	}
	if have, want := ss.SessionConnectionDuration, t2.Sub(t1); have != want {
		t.Fatalf("Snapshot(%q, ...): session connection duration counter mismatch: have %s; want %s", addr, have, want)
	}
	if have, want := ss.ConnectionTotalDuration, t2.Sub(t1); have != want {
		t.Fatalf("Snapshot(%q, ...): connection total duration counter mismatch: have %s; want %s", addr, have, want)
	}

	// Inc session conn retry.
	mc.Record(addr, metrics.IncSessionConnectionRetry())
	ss = snapshot(t, mc, t2, addr)
	if have, want := ss.SessionConnectionRetry, uint64(2); have != want {
		t.Fatalf("Snapshot(%q, ...): session connection retry counter mismatch: have %d; want %d", addr, have, want)
	}

	// Logout.
	mc.Record(addr, metrics.PeerLogOut(t3))
	ss = snapshot(t, mc, t2, addr)
	if have, want := ss.LastSeenTimestamp, t3.UnixNano(); have != want {
		t.Fatalf("Snapshot(%q, ...): last seen counter mismatch: have %d; want %d", addr, have, want)
	}
	if have, want := ss.ConnectionTotalDuration, t3.Sub(t1); have != want {
		t.Fatalf("Snapshot(%q, ...): connection total duration counter mismatch: have %s; want %s", addr, have, want)
	}
	if have, want := ss.SessionConnectionRetry, uint64(2); have != want {
		t.Fatalf("Snapshot(%q, ...): session connection retry counter mismatch: have %d; want %d", addr, have, want)
	}
	if have, want := ss.SessionConnectionDuration, t3.Sub(t1); have != want {
		t.Fatalf("Snapshot(%q, ...): session connection duration counter mismatch: have %q; want %q", addr, have, want)
	}

	// Bootnode.
	mc.Record(addr, metrics.IsBootnode(false))
	ss = snapshot(t, mc, t2, addr)
	if have, want := ss.IsBootnode, false; have != want {
		t.Fatalf("Snapshot(%q, ...): latency mismatch: have %v; want %v", addr, have, want)
	}
	mc.Record(addr, metrics.IsBootnode(true))
	ss = snapshot(t, mc, t2, addr)
	if have, want := ss.IsBootnode, true; have != want {
		t.Fatalf("Snapshot(%q, ...): is bootnode mismatch: have %v; want %v", addr, have, want)
	}

	// Latency.
	mc.Record(addr, metrics.PeerLatency(t4))
	ss = snapshot(t, mc, t2, addr)
	if have, want := ss.LatencyEWMA, t4; have != want {
		t.Fatalf("Snapshot(%q, ...): latency mismatch: have %d; want %d", addr, have, want)
	}
	mc.Record(addr, metrics.PeerLatency(t5))
	ss = snapshot(t, mc, t2, addr)
	if have, want := ss.LatencyEWMA, 19*time.Millisecond; have != want {
		t.Fatalf("Snapshot(%q, ...): latency mismatch: have %d; want %d", addr, have, want)
	}

	// Reachability.
	ss = snapshot(t, mc, t2, addr)
	if have, want := ss.Reachability, p2p.ReachabilityStatusUnknown; have != want {
		t.Fatalf("Snapshot(%q, ...): has reachability status mismatch: have %q; want %q", addr, have, want)
	}
	mc.Record(addr, metrics.PeerReachability(p2p.ReachabilityStatusPublic))
	ss = snapshot(t, mc, t2, addr)
	if have, want := ss.Reachability, p2p.ReachabilityStatusPublic; have != want {
		t.Fatalf("Snapshot(%q, ...): has reachability status mismatch: have %q; want %q", addr, have, want)
	}

	// Health.
	ss = snapshot(t, mc, t2, addr)
	if have, want := ss.Healthy, false; have != want {
		t.Fatalf("Snapshot(%q, ...): has health status mismatch: have %v; want %v", addr, have, want)
	}
	mc.Record(addr, metrics.PeerHealth(true))
	ss = snapshot(t, mc, t2, addr)
	if have, want := ss.Healthy, true; have != want {
		t.Fatalf("Snapshot(%q, ...): has health status mismatch: have %v; want %v", addr, have, want)
	}
	mc.Record(addr, metrics.PeerHealth(false))
	ss = snapshot(t, mc, t2, addr)
	if have, want := ss.Healthy, false; have != want {
		t.Fatalf("Snapshot(%q, ...): has health status mismatch: have %v; want %v", addr, have, want)
	}

	// Inspect.
	have := mc.Inspect(addr)
	want := ss
	if diff := cmp.Diff(have, want); diff != "" {
		t.Fatalf("unexpected snapshot difference:\n%s", diff)
	}

	// Flush.
	if err := mc.Flush(); err != nil {
		t.Fatalf("Flush(): unexpected error: %v", err)
	}

	// Finalize.
	mc.Record(addr, metrics.PeerLogIn(t1, metrics.PeerConnectionDirectionInbound))
	if err := mc.Finalize(t3, true); err != nil {
		t.Fatalf("Finalize(%s): unexpected error: %v", t3, err)
	}
	if have, want := len(mc.Snapshot(t2, addr)), 0; have != want {
		t.Fatalf("Finalize(%s): counters length mismatch: have %d; want %d", t3, have, want)
	}

	// Load the flushed metrics again from the persistent db.
	mc, err = metrics.NewCollector(db)
	if err != nil {
		t.Fatal(err)
	}
	if have, want := len(mc.Snapshot(t2, addr)), 1; have != want {
		t.Fatalf("NewCollector(...): counters length mismatch: have %d; want %d", have, want)
	}
	have = mc.Inspect(addr)
	want = &metrics.Snapshot{
		LastSeenTimestamp:       ss.LastSeenTimestamp,
		ConnectionTotalDuration: 2 * ss.ConnectionTotalDuration, // 2x because we've already logout with t3 and login with t1 again.
		IsBootnode:              true,
	}
	if diff := cmp.Diff(have, want); diff != "" {
		t.Fatalf("unexpected snapshot difference:\n%s", diff)
	}
}

func TestExclude(t *testing.T) {
	t.Parallel()

	db, err := shed.NewDB("", nil)
	if err != nil {
		t.Fatal(err)
	}
	testutil.CleanupCloser(t, db)

	mc, err := metrics.NewCollector(db)
	if err != nil {
		t.Fatal(err)
	}

	var addr = swarm.RandAddress(t)

	// record unhealthy, unreachable, bootnode
	mc.Record(addr, metrics.PeerHealth(false), metrics.IsBootnode(true), metrics.PeerReachability(p2p.ReachabilityStatusPrivate))

	if have, want := mc.Exclude(addr), false; have != want {
		t.Fatal("should not exclude any")
	}

	if have, want := mc.Exclude(addr, metrics.Bootnode()), true; have != want {
		t.Fatal("should exclude bootnodes")
	}

	if have, want := mc.Exclude(addr, metrics.Reachability(false)), true; have != want {
		t.Fatal("should exclude unreachble")
	}

	if have, want := mc.Exclude(addr, metrics.Health(false)), true; have != want {
		t.Fatal("should exclude unhealthy")
	}
}
