// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"renbrowser/internal/nomadnet"
)

// DownloadStatus is the lifecycle state of a tracked download shown in the
// downloads UI.
type DownloadStatus string

const (
	DownloadStatusPending     DownloadStatus = "pending"
	DownloadStatusDownloading DownloadStatus = "downloading"
	DownloadStatusCompleted   DownloadStatus = "completed"
	DownloadStatusFailed      DownloadStatus = "failed"
	DownloadStatusCanceled    DownloadStatus = "canceled"
)

// completedRetention is how long a finished download stays in the active
// list before it drops off (it remains visible in ListDownloads' on-disk
// history afterwards). Failed and canceled entries stay until dismissed so
// the error is not missed.
const completedRetention = 5 * time.Second

// ActiveDownload is the live state of a tracked mesh file download, used to
// drive the downloads panel's progress list, speed/ETA readouts, and the
// badge counter on the download icon.
type ActiveDownload struct {
	ID        string         `json:"id"`
	URL       string         `json:"url"`
	Name      string         `json:"name"`
	Path      string         `json:"path,omitempty"`
	Received  int64          `json:"received"`
	Total     int64          `json:"total"`
	Status    DownloadStatus `json:"status"`
	Error     string         `json:"error,omitempty"`
	StartedAt int64          `json:"startedAt"`
	UpdatedAt int64          `json:"updatedAt"`

	cancel context.CancelFunc
}

// downloadManager tracks in-flight/recent downloads in memory and notifies a
// subscriber (the app's event bus) whenever the set changes.
type downloadManager struct {
	mu         sync.Mutex
	items      map[string]*ActiveDownload
	order      []string
	nextID     int64
	lastNotify time.Time
	onChange   func([]ActiveDownload)
}

func newDownloadManager() *downloadManager {
	return &downloadManager{items: make(map[string]*ActiveDownload)}
}

func (m *downloadManager) start(url, name string) string {
	m.mu.Lock()
	m.nextID++
	id := fmt.Sprintf("dl-%d", m.nextID)
	now := time.Now().UnixMilli()
	m.items[id] = &ActiveDownload{
		ID:        id,
		URL:       url,
		Name:      name,
		Status:    DownloadStatusPending,
		StartedAt: now,
		UpdatedAt: now,
	}
	m.order = append([]string{id}, m.order...)
	m.mu.Unlock()
	m.notify()
	return id
}

func (m *downloadManager) bindCancel(id string, cancel context.CancelFunc) {
	m.mu.Lock()
	if d, ok := m.items[id]; ok {
		d.cancel = cancel
	}
	m.mu.Unlock()
}

// reportProgress records byte progress for id. Notifications are throttled
// app-wide so a fast local transfer can't flood the frontend with events.
func (m *downloadManager) reportProgress(id string, received, total int64) {
	m.mu.Lock()
	d, ok := m.items[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	firstByte := d.Received == 0 && received > 0
	d.Received = received
	if total > 0 {
		d.Total = total
	}
	d.Status = DownloadStatusDownloading
	d.UpdatedAt = time.Now().UnixMilli()
	force := firstByte || time.Since(m.lastNotify) >= 200*time.Millisecond
	if force {
		m.lastNotify = time.Now()
	}
	m.mu.Unlock()
	if force {
		m.notify()
	}
}

func (m *downloadManager) complete(id, path string, size int64) {
	m.mu.Lock()
	d, ok := m.items[id]
	if ok {
		d.Status = DownloadStatusCompleted
		d.Path = path
		if size > 0 {
			d.Received = size
			d.Total = size
		} else {
			d.Received = d.Total
		}
		d.UpdatedAt = time.Now().UnixMilli()
	}
	m.mu.Unlock()
	if !ok {
		return
	}
	m.notify()
	time.AfterFunc(completedRetention, func() { m.remove(id) })
}

func (m *downloadManager) fail(id, errMsg string) {
	m.mu.Lock()
	d, ok := m.items[id]
	if ok && d.Status == DownloadStatusCanceled {
		m.mu.Unlock()
		return
	}
	if ok {
		d.Status = DownloadStatusFailed
		d.Error = errMsg
		d.UpdatedAt = time.Now().UnixMilli()
	}
	m.mu.Unlock()
	if !ok {
		return
	}
	m.notify()
}

// cancel requests cancellation of an in-flight download. Returns false if
// the download is unknown or already finished.
func (m *downloadManager) cancel(id string) bool {
	m.mu.Lock()
	d, ok := m.items[id]
	if !ok || d.Status == DownloadStatusCompleted || d.Status == DownloadStatusFailed ||
		d.Status == DownloadStatusCanceled {
		m.mu.Unlock()
		return false
	}
	cancelFn := d.cancel
	d.Status = DownloadStatusCanceled
	d.Error = ""
	d.UpdatedAt = time.Now().UnixMilli()
	m.mu.Unlock()
	if cancelFn != nil {
		cancelFn()
	}
	m.notify()
	time.AfterFunc(completedRetention, func() { m.remove(id) })
	return true
}

func (m *downloadManager) dismiss(id string) {
	m.remove(id)
}

func (m *downloadManager) remove(id string) {
	m.mu.Lock()
	if _, ok := m.items[id]; !ok {
		m.mu.Unlock()
		return
	}
	delete(m.items, id)
	filtered := make([]string, 0, len(m.order))
	for _, oid := range m.order {
		if oid != id {
			filtered = append(filtered, oid)
		}
	}
	m.order = filtered
	m.mu.Unlock()
	m.notify()
}

func (m *downloadManager) list() []ActiveDownload {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]ActiveDownload, 0, len(m.order))
	for _, id := range m.order {
		if d, ok := m.items[id]; ok {
			cp := *d
			cp.cancel = nil
			out = append(out, cp)
		}
	}
	return out
}

func (m *downloadManager) notify() {
	if m.onChange == nil {
		return
	}
	m.onChange(m.list())
}

// downloadTracker binds a fetch in progress to a single ActiveDownload
// entry so its byte progress and cancellation can be observed by the
// downloads UI.
type downloadTracker struct {
	mgr *downloadManager
	id  string
}

func (t *downloadTracker) fetchHooks() *nomadnet.FetchHooks {
	if t == nil {
		return nil
	}
	return &nomadnet.FetchHooks{
		OnProgress: func(p nomadnet.FetchProgress) {
			t.mgr.reportProgress(t.id, p.Received, p.Total)
		},
	}
}

// mergeFetchHooks combines any number of fetch hook sets so a single fetch
// can report to both the dev-log console and the download manager.
func mergeFetchHooks(hooks ...*nomadnet.FetchHooks) *nomadnet.FetchHooks {
	return &nomadnet.FetchHooks{
		OnStage: func(stage, detail string) {
			for _, h := range hooks {
				if h != nil && h.OnStage != nil {
					h.OnStage(stage, detail)
				}
			}
		},
		OnProgress: func(p nomadnet.FetchProgress) {
			for _, h := range hooks {
				if h != nil && h.OnProgress != nil {
					h.OnProgress(p)
				}
			}
		},
	}
}
