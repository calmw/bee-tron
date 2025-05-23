// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package postage_test

import (
	"bytes"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/calmw/bee-tron/pkg/crypto"
	"github.com/calmw/bee-tron/pkg/postage"
	"github.com/calmw/bee-tron/pkg/postage/batchstore/mock"
	postagetesting "github.com/calmw/bee-tron/pkg/postage/testing"
	"github.com/calmw/bee-tron/pkg/storage"
	"github.com/calmw/bee-tron/pkg/storage/inmemstore"
	chunktesting "github.com/calmw/bee-tron/pkg/storage/testing"
)

// TestStampMarshalling tests the idempotence  of binary marshal/unmarshals for Stamps.
func TestStampMarshalling(t *testing.T) {
	t.Parallel()

	sExp := postagetesting.MustNewStamp()
	buf, _ := sExp.MarshalBinary()
	if len(buf) != postage.StampSize {
		t.Fatalf("invalid length for serialised stamp. expected %d, got  %d", postage.StampSize, len(buf))
	}
	s := postage.NewStamp(nil, nil, nil, nil)
	if err := s.UnmarshalBinary(buf); err != nil {
		t.Fatalf("unexpected error unmarshalling stamp: %v", err)
	}
	compareStamps(t, sExp, s)
}

// TestStampMarshalling tests the idempotence  of binary marshal/unmarshals for Stamps.
func TestStampJsonMarshalling(t *testing.T) {
	t.Parallel()

	sExp := postagetesting.MustNewStamp()

	b, err := json.Marshal(sExp)
	if err != nil {
		t.Fatal(err)
	}

	s := postage.NewStamp(nil, nil, nil, nil)
	err = json.Unmarshal(b, s)
	if err != nil {
		t.Fatal(err)
	}

	compareStamps(t, sExp, s)
}

func compareStamps(t *testing.T, s1, s2 *postage.Stamp) {
	t.Helper()

	if !bytes.Equal(s1.BatchID(), s2.BatchID()) {
		t.Fatalf("id mismatch, expected %x, got %x", s1.BatchID(), s2.BatchID())
	}
	if !bytes.Equal(s1.Index(), s2.Index()) {
		t.Fatalf("index mismatch, expected %x, got %x", s1.Index(), s2.Index())
	}
	if !bytes.Equal(s1.Timestamp(), s2.Timestamp()) {
		t.Fatalf("timestamp mismatch, expected %x, got %x", s1.Index(), s2.Index())
	}
	if !bytes.Equal(s1.Sig(), s2.Sig()) {
		t.Fatalf("sig mismatch, expected %x, got %x", s1.Sig(), s2.Sig())
	}
}

// TestStampIndexMarshalling tests the idempotence of stamp index serialisation.
func TestStampIndexMarshalling(t *testing.T) {
	t.Parallel()

	var (
		expBucket uint32 = 11789
		expIndex  uint32 = 199999
	)
	index := postage.IndexToBytes(expBucket, expIndex)
	bucket, idx := postage.BucketIndexFromBytes(index)
	if bucket != expBucket {
		t.Fatalf("bucket mismatch. want %d, got %d", expBucket, bucket)
	}
	if idx != expIndex {
		t.Fatalf("index mismatch. want %d, got %d", expIndex, idx)
	}
}

func TestValidStamp(t *testing.T) {
	t.Parallel()

	privKey, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		t.Fatal(err)
	}

	owner, err := crypto.NewEthereumAddress(privKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	b := postagetesting.MustNewBatch(postagetesting.WithOwner(owner))
	bs := mock.New(mock.WithBatch(b))
	signer := crypto.NewDefaultSigner(privKey)
	issuer := postage.NewStampIssuer("label", "keyID", b.ID, big.NewInt(3), b.Depth, b.BucketDepth, 1000, true)
	stamper := postage.NewStamper(inmemstore.New(), issuer, signer)

	// this creates a chunk with a mocked stamp. ValidStamp will override this
	// stamp on execution
	ch := chunktesting.GenerateTestRandomChunk()

	idAddress, err := storage.IdentityAddress(ch)
	if err != nil {
		t.Fatal(err)
	}

	st, err := stamper.Stamp(ch.Address(), idAddress)
	if err != nil {
		t.Fatal(err)
	}

	ch.WithStamp(st)

	// ensure the chunk doesn't have the batch details filled before we validate stamp
	if ch.Depth() == b.Depth || ch.BucketDepth() == b.BucketDepth {
		t.Fatal("expected chunk to not have correct depth and bucket depth at start")
	}

	ch, err = postage.ValidStamp(bs)(ch)
	if err != nil {
		t.Fatal(err)
	}

	compareStamps(t, st, ch.Stamp().(*postage.Stamp))

	if ch.Depth() != b.Depth {
		t.Fatalf("invalid batch depth added on chunk exp %d got %d", b.Depth, ch.Depth())
	}
	if ch.BucketDepth() != b.BucketDepth {
		t.Fatalf("invalid bucket depth added on chunk exp %d got %d", b.BucketDepth, ch.BucketDepth())
	}
	if ch.Immutable() != b.Immutable {
		t.Fatalf("invalid batch immutablility added on chunk exp %t got %t", b.Immutable, ch.Immutable())
	}
}
