// SPDX-License-Identifier: MIT
package nomadnet

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"quad4/reticulum-go/pkg/interfaces"
	rlink "quad4/reticulum-go/pkg/link"
)

func TestBitrateWindowMatchesPathExchangeFormula(t *testing.T) {
	cases := []struct {
		bps  int64
		want float64
	}{
		{125, 2*(240.0*8/125) + 10},
		{50, 2*(240.0*8/50) + 10},
		{5, 2*(240.0*8/50) + 10},
		{250, 2*(240.0*8/250) + 10},
		{10_000_000, 2*(240.0*8/10_000_000) + 10},
	}
	for _, tc := range cases {
		got := bitrateWindow(tc.bps)
		want := time.Duration(tc.want * float64(time.Second))
		if got != want {
			t.Fatalf("bitrateWindow(%d)=%v want %v", tc.bps, got, want)
		}
	}
	if bitrateWindow(0) != 0 {
		t.Fatal("zero bitrate should yield no window")
	}
}

func TestPathResponseWindowFloorsAtPathRequestTimeout(t *testing.T) {
	tr := testPathTransport(t)
	dest := bytes.Repeat([]byte{0x11}, 16)
	got := pathResponseWindow(tr, dest)
	if got < pathRequestTimeout {
		t.Fatalf("got %v, want at least %v", got, pathRequestTimeout)
	}
}

func TestPathResponseWindowFollowsSlowestOnlineBitrate(t *testing.T) {
	tr := testPathTransport(t)
	dest := bytes.Repeat([]byte{0x12}, 16)
	slow, err := interfaces.NewUDPInterface("slow-pr", "127.0.0.1:0", "", true)
	if err != nil {
		t.Fatal(err)
	}
	slow.Bitrate = 125
	slow.Enable()
	if err := tr.RegisterInterface("slow-pr", slow); err != nil {
		t.Fatal(err)
	}

	got := pathResponseWindow(tr, dest)
	want := bitrateWindow(125)
	if first := time.Duration(tr.GetFirstHopTimeoutRPC(dest) * float64(time.Second)); first > want {
		want = first
	}
	if want < pathRequestTimeout {
		want = pathRequestTimeout
	}
	if got != want {
		t.Fatalf("pathResponseWindow=%v want %v", got, want)
	}
}

func TestPathResponseWindowClampsBitrateToFifty(t *testing.T) {
	tr := testPathTransport(t)
	dest := bytes.Repeat([]byte{0x13}, 16)
	slow, err := interfaces.NewUDPInterface("clamp-pr", "127.0.0.1:0", "", true)
	if err != nil {
		t.Fatal(err)
	}
	slow.Bitrate = 5
	slow.Enable()
	if err := tr.RegisterInterface("clamp-pr", slow); err != nil {
		t.Fatal(err)
	}
	got := pathResponseWindow(tr, dest)
	want := bitrateWindow(minWindowBitrate)
	if got != want {
		t.Fatalf("pathResponseWindow=%v want clamped %v", got, want)
	}
}

func TestLinkEstablishWindowAddsMargin(t *testing.T) {
	tr := testPathTransport(t)
	dest := bytes.Repeat([]byte{0x14}, 16)
	got := linkEstablishWindow(tr, dest)
	if got <= time.Duration(rlink.EstablishmentTimeoutPerHop)*time.Second {
		t.Fatalf("link window %v should exceed per-hop timeout", got)
	}
	if got < pathRequestTimeout+linkEstablishMargin {
		t.Fatalf("link window %v too small", got)
	}
}

func TestLinkEstablishWindowUsesBitrateWhenSlower(t *testing.T) {
	tr := testPathTransport(t)
	dest := bytes.Repeat([]byte{0x15}, 16)
	slow, err := interfaces.NewUDPInterface("slow-link", "127.0.0.1:0", "", true)
	if err != nil {
		t.Fatal(err)
	}
	slow.Bitrate = 50
	slow.Enable()
	if err := tr.RegisterInterface("slow-link", slow); err != nil {
		t.Fatal(err)
	}
	got := linkEstablishWindow(tr, dest)
	want := bitrateWindow(50) + linkEstablishMargin
	if got != want {
		t.Fatalf("linkEstablishWindow=%v want %v", got, want)
	}
}

func TestFetchBudgetCoversPathLinkAndReceipt(t *testing.T) {
	tr := testPathTransport(t)
	hash := "16161616161616161616161616161616"
	dest := bytes.Repeat([]byte{0x16}, 16)
	got := FetchBudget(tr, hash, "/page/index.mu")
	min := pathResponseWindow(tr, dest) + linkEstablishWindow(tr, dest) + defaultReceiptTimeout + fetchBudgetSlack
	if got != min {
		t.Fatalf("FetchBudget=%v want %v", got, min)
	}
}

func TestFetchBudgetFilesHaveMinimum(t *testing.T) {
	tr := testPathTransport(t)
	got := FetchBudget(tr, "17171717171717171717171717171717", "/file/song.mp3")
	if got < fileFetchBudgetMin {
		t.Fatalf("file FetchBudget=%v want at least %v", got, fileFetchBudgetMin)
	}
}

func TestWaitLinkEstablishedRespectsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := waitLinkEstablished(ctx, nil, make(chan struct{}), time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestWaitLinkEstablishedTimesOut(t *testing.T) {
	start := time.Now()
	err := waitLinkEstablished(context.Background(), nil, make(chan struct{}), 80*time.Millisecond)
	if !errors.Is(err, errLinkEstablishTimeout) {
		t.Fatalf("got %v", err)
	}
	if elapsed := time.Since(start); elapsed < 50*time.Millisecond || elapsed > time.Second {
		t.Fatalf("unexpected duration %v", elapsed)
	}
}
