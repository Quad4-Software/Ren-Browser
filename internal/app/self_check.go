// SPDX-License-Identifier: MIT
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"quad4/reticulum-go/pkg/common"
	"quad4/reticulum-go/pkg/reticulumconfig"

	"renbrowser/internal/brand"
	"renbrowser/internal/nomadnet"
	"renbrowser/internal/rns"
)

type CheckStatus struct {
	Passed bool   `json:"passed"`
	Reason string `json:"reason,omitempty"`
}

type SelfCheckResult struct {
	StackUp       CheckStatus `json:"stackUp"`
	ConfigGood    CheckStatus `json:"configGood"`
	DBGood        CheckStatus `json:"dbGood"`
	ReadWriteGood CheckStatus `json:"readWriteGood"`
	DownloadsGood CheckStatus `json:"downloadsGood"`
	Interfaces    CheckStatus `json:"interfaces"`
	Discovery     CheckStatus `json:"discovery"`
	PageFetch     CheckStatus `json:"pageFetch"`
	AllPassed     bool        `json:"allPassed"`
	MeshEnabled   bool        `json:"meshEnabled"`
}

const (
	selfCheckMeshEnvKey       = brand.EnvPrefix + "_SELF_CHECK_MESH"
	selfCheckMeshCountEnvKey  = brand.EnvPrefix + "_SELF_CHECK_MESH_COUNT"
	selfCheckMeshWaitEnvKey   = brand.EnvPrefix + "_SELF_CHECK_MESH_WAIT_SEC"
	defaultSelfCheckMeshCount = 4
	maxSelfCheckMeshCount     = 5
	defaultSelfCheckMeshWait  = 45 * time.Second
)

// RunSelfCheck performs a comprehensive internal health check of the application components.
// Set REN_BROWSER_SELF_CHECK_MESH=1 to also seed community TCP interfaces, wait for
// announces, and attempt a page fetch against a discovered node.
func (s *BrowserService) RunSelfCheck() SelfCheckResult {
	res := SelfCheckResult{
		AllPassed:   true,
		MeshEnabled: selfCheckMeshEnabled(),
	}

	s.checkStack(&res)
	s.checkConfig(&res)
	s.checkDB(&res)
	s.checkReadWrite(&res)
	s.checkDownloads(&res)
	s.checkInterfaces(&res)

	if res.MeshEnabled {
		s.checkMeshDiscoveryAndPage(&res)
	} else {
		res.Discovery = CheckStatus{Passed: true, Reason: "skipped (set " + selfCheckMeshEnvKey + "=1 to enable)"}
		res.PageFetch = CheckStatus{Passed: true, Reason: "skipped (set " + selfCheckMeshEnvKey + "=1 to enable)"}
	}

	return res
}

func selfCheckMeshEnabled() bool {
	v := strings.TrimSpace(os.Getenv(selfCheckMeshEnvKey))
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func selfCheckMeshCount() int {
	v := strings.TrimSpace(os.Getenv(selfCheckMeshCountEnvKey))
	if v == "" {
		return defaultSelfCheckMeshCount
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return defaultSelfCheckMeshCount
	}
	if n > maxSelfCheckMeshCount {
		return maxSelfCheckMeshCount
	}
	return n
}

func selfCheckMeshWait() time.Duration {
	v := strings.TrimSpace(os.Getenv(selfCheckMeshWaitEnvKey))
	if v == "" {
		return defaultSelfCheckMeshWait
	}
	sec, err := strconv.Atoi(v)
	if err != nil || sec < 5 {
		return defaultSelfCheckMeshWait
	}
	if sec > 180 {
		sec = 180
	}
	return time.Duration(sec) * time.Second
}

func (s *BrowserService) checkStack(res *SelfCheckResult) {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()

	if stack == nil {
		res.StackUp = CheckStatus{Passed: false, Reason: "Reticulum stack is not initialized"}
		res.AllPassed = false
		return
	}
	if !stack.IsStarted() {
		res.StackUp = CheckStatus{Passed: false, Reason: "Reticulum stack is initialized but not started"}
		res.AllPassed = false
		return
	}
	res.StackUp = CheckStatus{Passed: true}
}

func (s *BrowserService) checkConfig(res *SelfCheckResult) {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()

	configPath := ""
	if stack != nil {
		configPath = stack.ConfigPath()
	}
	if configPath == "" {
		configPath = filepath.Join(rns.DefaultConfigDir(), "config")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		res.ConfigGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Reticulum config file not found at: %s", configPath)}
		res.AllPassed = false
		return
	}
	if _, err := reticulumconfig.LoadConfig(configPath); err != nil {
		res.ConfigGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Failed to parse Reticulum config: %v", err)}
		res.AllPassed = false
		return
	}
	res.ConfigGood = CheckStatus{Passed: true}
}

func (s *BrowserService) checkDB(res *SelfCheckResult) {
	health := s.GetStoreHealth()
	if !health.OK {
		res.DBGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Database health check failed: %s (%s)", health.Detail, health.Kind)}
		res.AllPassed = false
		return
	}
	res.DBGood = CheckStatus{Passed: true}
}

func (s *BrowserService) checkReadWrite(res *SelfCheckResult) {
	if s.store == nil {
		res.ReadWriteGood = CheckStatus{Passed: false, Reason: "Database store is unavailable"}
		res.AllPassed = false
		return
	}
	testKey := "self_check_last_run"
	testVal := time.Now().UTC().Format(time.RFC3339)
	if err := s.store.SetSetting(testKey, testVal); err != nil {
		res.ReadWriteGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Database write failed: %v", err)}
		res.AllPassed = false
		return
	}
	val, err := s.store.GetSetting(testKey)
	if err != nil {
		res.ReadWriteGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Database read failed: %v", err)}
		res.AllPassed = false
		return
	}
	if val != testVal {
		res.ReadWriteGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Database read value mismatch: expected %q, got %q", testVal, val)}
		res.AllPassed = false
		return
	}
	res.ReadWriteGood = CheckStatus{Passed: true}
}

func (s *BrowserService) checkDownloads(res *SelfCheckResult) {
	downloadDir := s.GetDownloadDir()
	if downloadDir == "" {
		res.DownloadsGood = CheckStatus{Passed: false, Reason: "Downloads directory path is empty"}
		res.AllPassed = false
		return
	}
	// #nosec G301 -- downloads folder permissions are intended to be user-readable/writable (0755)
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		res.DownloadsGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Failed to create downloads directory: %v", err)}
		res.AllPassed = false
		return
	}
	tempFile := filepath.Join(downloadDir, ".self_check_temp")
	testData := []byte("self_check_data")
	// #nosec G306 -- self-check temporary file permissions are fine at 0644
	if err := os.WriteFile(tempFile, testData, 0o644); err != nil {
		res.DownloadsGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Failed to write to downloads directory: %v", err)}
		res.AllPassed = false
		return
	}
	// #nosec G304 -- tempFile path is safely constructed within the downloads directory
	readData, err := os.ReadFile(tempFile)
	_ = os.Remove(tempFile)
	if err != nil {
		res.DownloadsGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Failed to read from downloads directory: %v", err)}
		res.AllPassed = false
		return
	}
	if string(readData) != string(testData) {
		res.DownloadsGood = CheckStatus{Passed: false, Reason: "Downloads read/write data mismatch"}
		res.AllPassed = false
		return
	}
	res.DownloadsGood = CheckStatus{Passed: true}
}

func (s *BrowserService) checkInterfaces(res *SelfCheckResult) {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		res.Interfaces = CheckStatus{Passed: false, Reason: "Reticulum stack is not initialized"}
		res.AllPassed = false
		return
	}

	ifaces := stack.ListInterfaces()
	enabled := make([]rns.InterfaceInfo, 0, len(ifaces))
	for _, iface := range ifaces {
		if iface.Enabled {
			enabled = append(enabled, iface)
		}
	}
	if len(enabled) == 0 {
		res.Interfaces = CheckStatus{Passed: true, Reason: "no enabled interfaces configured"}
		return
	}

	online := 0
	var totalTx, totalRx uint64
	parts := make([]string, 0, len(enabled))
	for _, iface := range enabled {
		totalTx += iface.TxBytes
		totalRx += iface.RxBytes
		state := "offline"
		if iface.Online {
			online++
			state = "online"
		}
		parts = append(parts, fmt.Sprintf("%s=%s tx=%d rx=%d", iface.Name, state, iface.TxBytes, iface.RxBytes))
	}

	summary := fmt.Sprintf("%d/%d online; tx=%d rx=%d; %s", online, len(enabled), totalTx, totalRx, strings.Join(parts, "; "))
	if online == 0 {
		res.Interfaces = CheckStatus{Passed: false, Reason: summary}
		res.AllPassed = false
		return
	}
	res.Interfaces = CheckStatus{Passed: true, Reason: summary}
}

func (s *BrowserService) checkMeshDiscoveryAndPage(res *SelfCheckResult) {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil || !stack.IsStarted() {
		res.Discovery = CheckStatus{Passed: false, Reason: "stack not started; cannot run mesh discovery"}
		res.PageFetch = CheckStatus{Passed: false, Reason: "skipped (stack not started)"}
		res.AllPassed = false
		return
	}

	added, err := s.ensureSelfCheckCommunityInterfaces(stack)
	if err != nil {
		res.Discovery = CheckStatus{Passed: false, Reason: fmt.Sprintf("community interfaces: %v", err)}
		res.PageFetch = CheckStatus{Passed: false, Reason: "skipped (community interfaces failed)"}
		res.AllPassed = false
		return
	}

	wait := selfCheckMeshWait()
	deadline := time.Now().Add(wait)
	var nodeCount int
	var onlineIfaces int
	for {
		st := stack.Status()
		nodeCount = st.NodeCount
		onlineIfaces = st.InterfacesOnline
		if nodeCount > 0 && onlineIfaces > 0 {
			break
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	addedNote := "using existing interfaces"
	if len(added) > 0 {
		addedNote = fmt.Sprintf("seeded %d community TCP: %s", len(added), strings.Join(added, ", "))
	}
	discoveryReason := fmt.Sprintf("%s; online_ifaces=%d; discovered_nodes=%d (waited up to %s)", addedNote, onlineIfaces, nodeCount, wait.Round(time.Second))
	if nodeCount == 0 {
		res.Discovery = CheckStatus{Passed: false, Reason: discoveryReason}
		res.PageFetch = CheckStatus{Passed: false, Reason: "skipped (no discovered nodes)"}
		res.AllPassed = false
		return
	}
	res.Discovery = CheckStatus{Passed: true, Reason: discoveryReason}

	pageURL, pageErr := s.selfCheckPageURL(stack)
	if pageErr != nil {
		res.PageFetch = CheckStatus{Passed: false, Reason: pageErr.Error()}
		res.AllPassed = false
		return
	}
	page := s.OpenFreshURL(pageURL)
	if page.Error != "" {
		res.PageFetch = CheckStatus{Passed: false, Reason: fmt.Sprintf("%s: %s", pageURL, page.Error)}
		res.AllPassed = false
		return
	}
	bytes := len(page.Raw)
	if bytes == 0 {
		bytes = len(page.HTML)
	}
	res.PageFetch = CheckStatus{
		Passed: true,
		Reason: fmt.Sprintf("%s ok (%d bytes, %dms, hops=%d)", pageURL, bytes, page.DurationMs, page.Hops),
	}
}

func (s *BrowserService) ensureSelfCheckCommunityInterfaces(stack *rns.Stack) ([]string, error) {
	cfg := stack.Config()
	if cfg == nil {
		return nil, fmt.Errorf("config not loaded")
	}
	if rns.ConfigHasOutboundCommunityInterfaces(cfg) {
		// Still ensure at least one enabled interface is online-capable; do not reseed.
		return nil, nil
	}

	result, err := rns.FetchCommunityInterfaces(nil)
	if err != nil {
		return nil, err
	}
	count := selfCheckMeshCount()
	picked := rns.PickSeedableCommunityInterfaces(result.Items, count)
	if len(picked) == 0 {
		return nil, fmt.Errorf("no seedable community TCP interfaces available")
	}

	cloned := cloneReticulumConfig(cfg)
	added := rns.ApplyCommunityInterfacesToConfig(cloned, picked)
	if len(added) == 0 {
		return nil, fmt.Errorf("failed to apply community interfaces to config")
	}
	if err := stack.ApplyConfig(cloned); err != nil {
		return nil, err
	}
	return added, nil
}

func cloneReticulumConfig(cfg *common.ReticulumConfig) *common.ReticulumConfig {
	if cfg == nil {
		return nil
	}
	out := *cfg
	out.Interfaces = make(map[string]*common.InterfaceConfig, len(cfg.Interfaces))
	for name, iface := range cfg.Interfaces {
		if iface == nil {
			out.Interfaces[name] = nil
			continue
		}
		cp := *iface
		if iface.I2PPeers != nil {
			cp.I2PPeers = append([]string(nil), iface.I2PPeers...)
		}
		if iface.Devices != nil {
			cp.Devices = append([]string(nil), iface.Devices...)
		}
		if iface.IgnoredDevices != nil {
			cp.IgnoredDevices = append([]string(nil), iface.IgnoredDevices...)
		}
		out.Interfaces[name] = &cp
	}
	if cfg.RPCKey != nil {
		out.RPCKey = append([]byte(nil), cfg.RPCKey...)
	}
	return &out
}

func (s *BrowserService) selfCheckPageURL(stack *rns.Stack) (string, error) {
	if stack == nil || stack.Handler() == nil {
		return "", fmt.Errorf("announce handler unavailable")
	}
	nodes := stack.Handler().List()
	var pick *nomadnet.Node
	for i := range nodes {
		n := &nodes[i]
		if !n.Enabled {
			continue
		}
		if pick == nil || n.LastSeen > pick.LastSeen {
			pick = n
		}
	}
	if pick == nil && len(nodes) > 0 {
		pick = &nodes[0]
	}
	if pick == nil || strings.TrimSpace(pick.Hash) == "" {
		return "", fmt.Errorf("no node hash available for page fetch")
	}
	return nomadnet.FormatURL(pick.Hash, "/page/index.mu"), nil
}
