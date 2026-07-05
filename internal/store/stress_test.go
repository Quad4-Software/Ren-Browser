//go:build stress

// SPDX-License-Identifier: MIT

package store_test

import (
	"sync"
	"testing"

	"renbrowser/internal/store"
)

func TestStoreConcurrentFavorites(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/stress.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	const workers = 16
	const rounds = 64
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < rounds; i++ {
				url := "deadbeef:/page/index.mu"
				_ = s.AddFavorite(url)
				_ = s.Favorites()
				_ = s.RemoveFavorite(url)
			}
		}(w)
	}
	wg.Wait()
}

func TestStoreConcurrentSettings(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/settings-stress.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	var wg sync.WaitGroup
	wg.Add(8)
	for w := 0; w < 8; w++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 32; i++ {
				key := "k"
				_ = s.SetSetting(key, "v")
				_, _ = s.GetSetting(key)
			}
		}(w)
	}
	wg.Wait()
}
