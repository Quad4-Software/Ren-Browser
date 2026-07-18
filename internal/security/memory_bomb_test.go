// SPDX-License-Identifier: MIT
package security_test

import (
	"runtime"
	"testing"
	"unsafe"

	"quad4/reticulum-go/pkg/common"
	"quad4/reticulum-go/pkg/resource"

	"renbrowser/internal/limits"
	"renbrowser/internal/nomadnet"
)

// TheoreticalMaxResourceTransferBytes is the upper bound on TransferSize
// reticulum-go will accept for an incoming resource advertisement:
// Parts <= MaxSegments and TransferSize <= Parts * resourceSDU.
// With DefaultMTU the SDU is mtu - 36 (see link.resourceSDU).
func theoreticalMaxResourceTransferBytes() int64 {
	sdu := common.DefaultMTU - 35 - 1
	return int64(resource.MaxSegments) * int64(sdu)
}

func TestResourceAdvertisementAllowsMultiTensOfMiB(t *testing.T) {
	max := theoreticalMaxResourceTransferBytes()
	// RNS wire format allows this. App file policy must not assume MTU-sized replies.
	if max < 80*1024*1024 {
		t.Fatalf("expected MaxSegments*SDU >= 80MiB, got %d", max)
	}
	t.Logf("max accepted TransferSize at DefaultMTU: %d bytes (%.1f MiB)", max, float64(max)/(1024*1024))
}

func TestIncomingResourcePartSlotHeaderCost(t *testing.T) {
	// Mirrors beginIncomingResource pre-alloc of partSlots + mapHashes before
	// any part bytes arrive. A linked peer that advertises MaxSegments forces
	// this cost even if Ren Browser later rejects on size.
	parts := int(resource.MaxSegments)
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	partSlots := make([][]byte, parts)
	mapHashes := make([][]byte, parts)
	runtime.KeepAlive(partSlots)
	runtime.KeepAlive(mapHashes)

	runtime.ReadMemStats(&after)
	delta := int64(after.HeapAlloc) - int64(before.HeapAlloc)
	headerBytes := int64(parts) * int64(2*unsafe.Sizeof([]byte(nil)))
	t.Logf("MaxSegments=%d slice-header estimate=%d heap delta≈%d", parts, headerBytes, delta)
	if headerBytes < 2*1024*1024 {
		t.Fatalf("expected multi-MiB header footprint for MaxSegments, got estimate %d", headerBytes)
	}
	// Sanity: allocation succeeded and is addressable.
	if len(partSlots) != parts || len(mapHashes) != parts {
		t.Fatal("part slot slices not sized to MaxSegments")
	}
}

func TestDefaultFilePolicyIsUnlimited(t *testing.T) {
	t.Setenv("REN_BROWSER_MAX_FILE_BYTES", "")
	if limits.MaxFileBytes() != 0 {
		t.Fatalf("default file cap=%d; want 0 (unlimited)", limits.MaxFileBytes())
	}
	if limits.MaxFetchBytes("/file/payload.bin") != 0 {
		t.Fatal("/file/ paths must use unlimited default")
	}
	if err := nomadnet.CheckResponseSize(make([]byte, 32*1024*1024), 32*1024*1024, 0); err != nil {
		t.Fatalf("unlimited CheckResponseSize rejected 32MiB: %v", err)
	}
}

func TestPagePolicyCapsCheckResponseSize(t *testing.T) {
	limit := limits.DefaultMaxPageBytes
	err := nomadnet.CheckResponseSize([]byte("x"), int64(limit)+1, limit)
	if err == nil {
		t.Fatal("expected advertised oversize page to fail")
	}
	err = nomadnet.CheckResponseSize(make([]byte, limit+1), int64(limit+1), limit)
	if err == nil {
		t.Fatal("expected received oversize page to fail")
	}
}

func TestFileCapEnvEnablesRejection(t *testing.T) {
	t.Setenv("REN_BROWSER_MAX_FILE_BYTES", "1024")
	cap := limits.MaxFetchBytes("/file/big.bin")
	if cap != 1024 {
		t.Fatalf("cap=%d want 1024", cap)
	}
	if err := nomadnet.CheckResponseSize(nil, 2048, cap); err == nil {
		t.Fatal("expected advertised file oversize to fail when cap set")
	}
}

func TestBase64DownloadAmplification(t *testing.T) {
	// DownloadFile returns base64 of the full body to the webview. That alone
	// expands storage ~4/3 on top of the already-buffered raw bytes.
	raw := 48 * 1024 * 1024
	b64 := (raw + 2) / 3 * 4
	if b64 < raw+raw/3 {
		t.Fatalf("base64 size %d unexpectedly small for %d raw", b64, raw)
	}
	t.Logf("48MiB file -> ~%d MiB base64 return path", b64/(1024*1024))
}

func TestCompressedResourceDecompressBound(t *testing.T) {
	// Incoming compressed resources are bounded by AutoCompressMaxSize on the
	// decompressed side. Still a large single allocation from a linked peer.
	if resource.AutoCompressMaxSize < 15*1024*1024 {
		t.Fatalf("AutoCompressMaxSize=%d unexpectedly low", resource.AutoCompressMaxSize)
	}
	t.Logf("bzip2 decompress ceiling: %d bytes", resource.AutoCompressMaxSize)
}
