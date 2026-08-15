// SPDX-License-Identifier: MIT
package nomadnet

import (
	"math"
	"reflect"
	"strings"
	"time"

	"quad4/reticulum-go/pkg/common"
	rlink "quad4/reticulum-go/pkg/link"
	"quad4/reticulum-go/pkg/transport"
)

const (
	minWindowBitrate    int64 = 50
	pathExchangeBytes         = 240
	pathRequestTimeout        = 15 * time.Second
	linkEstablishMargin       = 6 * time.Second
	fetchBudgetSlack          = 10 * time.Second
	fileFetchBudgetMin        = 300 * time.Second
)

func bitrateWindow(bitrate int64) time.Duration {
	if bitrate <= 0 {
		return 0
	}
	if bitrate < minWindowBitrate {
		bitrate = minWindowBitrate
	}
	sec := 2*(float64(pathExchangeBytes)*8/float64(bitrate)) + 10
	return time.Duration(sec * float64(time.Second))
}

func interfaceBitrate(iface common.NetworkInterface) int64 {
	if iface == nil {
		return 0
	}
	switch br := iface.(type) {
	case interface{ GetBitrate() int64 }:
		if v := br.GetBitrate(); v > 0 {
			return v
		}
	case interface{ GetBitrate() int }:
		if v := br.GetBitrate(); v > 0 {
			return int64(v)
		}
	case interface{ GetBitrate() uint64 }:
		if v := br.GetBitrate(); v > 0 {
			if v > math.MaxInt64 {
				return math.MaxInt64
			}
			return int64(v)
		}
	}
	rv := reflect.ValueOf(iface)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return 0
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() || rv.Kind() != reflect.Struct {
		return 0
	}
	f := rv.FieldByName("Bitrate")
	if f.IsValid() && f.CanInt() {
		return f.Int()
	}
	return 0
}

func slowestOnlineBitrate(tr *transport.Transport) int64 {
	if tr == nil {
		return 0
	}
	var slowest int64
	for _, iface := range tr.GetInterfaces() {
		if iface == nil || !iface.IsOnline() {
			continue
		}
		br := interfaceBitrate(iface)
		if br <= 0 {
			continue
		}
		if slowest == 0 || br < slowest {
			slowest = br
		}
	}
	return slowest
}

func pathResponseWindow(tr *transport.Transport, destHash []byte) time.Duration {
	var window time.Duration
	if tr != nil {
		window = time.Duration(tr.GetFirstHopTimeoutRPC(destHash) * float64(time.Second))
	}
	if w := bitrateWindow(slowestOnlineBitrate(tr)); w > window {
		window = w
	}
	if window < pathRequestTimeout {
		return pathRequestTimeout
	}
	return window
}

func linkEstablishWindow(tr *transport.Transport, destHash []byte) time.Duration {
	reported := time.Duration(rlink.EstablishmentTimeoutPerHop) * time.Second
	if tr != nil {
		firstHop := time.Duration(tr.GetFirstHopTimeoutRPC(destHash) * float64(time.Second))
		extra := 1
		hops := tr.HopsTo(destHash)
		if hops > 0 && hops < transport.PathfinderM {
			extra = max(int(hops)-1, 1)
		}
		reported = firstHop + time.Duration(extra)*time.Duration(rlink.EstablishmentTimeoutPerHop)*time.Second
	}
	window := reported
	if w := bitrateWindow(slowestOnlineBitrate(tr)); w > window {
		window = w
	}
	return window + linkEstablishMargin
}

// FetchBudget is the outer fetch deadline covering path window, link handshake,
// and receipt wait, sized from online interface bitrate so slow links can finish.
func FetchBudget(tr *transport.Transport, nodeHash, path string) time.Duration {
	destHash, _ := decodeNodeHash(nodeHash)
	_, receipt := requestTimeouts(path)
	budget := pathResponseWindow(tr, destHash) + linkEstablishWindow(tr, destHash) + receipt + fetchBudgetSlack
	if strings.HasPrefix(path, "/file/") && budget < fileFetchBudgetMin {
		return fileFetchBudgetMin
	}
	return budget
}
