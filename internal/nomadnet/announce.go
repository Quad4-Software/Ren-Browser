// SPDX-License-Identifier: MIT
package nomadnet

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"quad4/msgpack/v5/pkg/msgpack"
	"quad4/reticulum-go/pkg/identity"
)

const (
	AppName   = "nomadnetwork"
	AppAspect = "node"
	AspectKey = AppName + "." + AppAspect
)

type Node struct {
	Hash      string `json:"hash"`
	Name      string `json:"name"`
	Hops      uint8  `json:"hops"`
	Enabled   bool   `json:"enabled"`
	Timestamp int64  `json:"timestamp"`
	MaxSizeKB int16  `json:"maxSizeKb"`
	LastSeen  int64  `json:"lastSeen"`
}

type announcedPeer struct {
	node Node
	id   *identity.Identity
}

type AnnounceHandler struct {
	mu         sync.RWMutex
	nodes      map[string]announcedPeer
	onAnnounce func(Node)
}

func NewAnnounceHandler() *AnnounceHandler {
	return &AnnounceHandler{
		nodes: make(map[string]announcedPeer),
	}
}

func (h *AnnounceHandler) SetOnAnnounce(fn func(Node)) {
	h.mu.Lock()
	h.onAnnounce = fn
	h.mu.Unlock()
}

func (h *AnnounceHandler) AspectFilter() []string {
	return []string{AspectKey}
}

func (h *AnnounceHandler) ReceivePathResponses() bool {
	return true
}

func (h *AnnounceHandler) ReceivedAnnounce(destHash []byte, ident any, appData []byte, hops uint8) error {
	id, ok := ident.(*identity.Identity)
	if !ok || id == nil || len(destHash) != 16 {
		return nil
	}
	if !IsNodeDestination(id, destHash) {
		return nil
	}

	name, enabled, ts, maxKB := parseAnnounceAppData(appData)
	hash := hex.EncodeToString(destHash)
	node := Node{
		Hash:      hash,
		Name:      name,
		Hops:      hops,
		Enabled:   enabled,
		Timestamp: ts,
		MaxSizeKB: maxKB,
		LastSeen:  time.Now().Unix(),
	}

	h.mu.Lock()
	h.nodes[hash] = announcedPeer{node: node, id: id}
	onAnnounce := h.onAnnounce
	h.mu.Unlock()
	if onAnnounce != nil {
		onAnnounce(node)
	}
	return nil
}

func (h *AnnounceHandler) List() []Node {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]Node, 0, len(h.nodes))
	for _, peer := range h.nodes {
		out = append(out, peer.node)
	}
	return out
}

func (h *AnnounceHandler) Get(hash string) (Node, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	peer, ok := h.nodes[strings.ToLower(hash)]
	return peer.node, ok
}

func (h *AnnounceHandler) Identity(hash string) (*identity.Identity, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	peer, ok := h.nodes[strings.ToLower(hash)]
	if !ok || peer.id == nil {
		return nil, false
	}
	return peer.id, true
}

func IsNodeDestination(id *identity.Identity, destHash []byte) bool {
	nameHashFull := sha256.Sum256([]byte(AspectKey))
	nameHash10 := nameHashFull[:10]
	identityHash := identity.TruncatedHash(id.GetPublicKey())
	combined := append(append([]byte(nil), nameHash10...), identityHash...)
	expectedFull := sha256.Sum256(combined)
	expected := expectedFull[:16]
	return string(expected) == string(destHash)
}

func parseAnnounceAppData(appData []byte) (name string, enabled bool, timestamp int64, maxKB int16) {
	if len(appData) == 0 {
		return "", true, 0, 0
	}

	var decoded any
	if err := msgpack.Unmarshal(appData, &decoded); err != nil {
		if utf8.Valid(appData) {
			return strings.TrimSpace(string(appData)), true, 0, 0
		}
		return "", true, 0, 0
	}

	if n, ok := extractNodeName(decoded); ok {
		name = n
	}
	if en, ts, max, ok := parseNodeStatus(decoded); ok {
		enabled = en
		timestamp = ts
		maxKB = max
	}
	if name == "" && utf8.Valid(appData) {
		if raw := strings.TrimSpace(string(appData)); raw != "" {
			name = raw
		}
	}
	return name, enabled, timestamp, maxKB
}

func extractNodeName(decoded any) (string, bool) {
	if decoded == nil {
		return "", false
	}
	if str, ok := msgpackString(decoded); ok {
		str = strings.TrimSpace(str)
		return str, str != ""
	}
	arr, ok := decoded.([]any)
	if !ok {
		return "", false
	}
	for _, v := range arr {
		if str, ok := msgpackString(v); ok {
			str = strings.TrimSpace(str)
			if str != "" {
				return str, true
			}
		}
	}
	return "", false
}

func parseNodeStatus(decoded any) (enabled bool, timestamp int64, maxKB int16, ok bool) {
	arr, ok := decoded.([]any)
	if !ok || len(arr) < 3 {
		return false, 0, 0, false
	}
	enabledVal, ok := arr[0].(bool)
	if !ok {
		return false, 0, 0, false
	}
	ts, ok := msgpackInt64(arr[1])
	if !ok {
		return false, 0, 0, false
	}
	maxVal, ok := msgpackInt64(arr[2])
	if !ok {
		return false, 0, 0, false
	}
	return enabledVal, ts, int16(maxVal), true // #nosec G115 -- max size field fits NomadNet wire format
}

func msgpackString(v any) (string, bool) {
	switch val := v.(type) {
	case string:
		return val, true
	case []byte:
		if utf8.Valid(val) {
			return string(val), true
		}
		return "", false
	default:
		return "", false
	}
}

func msgpackInt64(v any) (int64, bool) {
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int8:
		return int64(val), true
	case int16:
		return int64(val), true
	case int32:
		return int64(val), true
	case int64:
		return val, true
	case uint:
		return int64(val), true // #nosec G115 -- msgpack values bounded by protocol
	case uint8:
		return int64(val), true
	case uint16:
		return int64(val), true
	case uint32:
		return int64(val), true
	case uint64:
		return int64(val), true // #nosec G115 -- msgpack values bounded by protocol
	default:
		return 0, false
	}
}
