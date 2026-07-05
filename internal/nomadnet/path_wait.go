package nomadnet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"quad4/reticulum-go/pkg/transport"
)

var errInvalidPathDestination = errors.New("invalid destination")

const (
	pathPollInterval = 100 * time.Millisecond
	pathWaitDefault  = 35 * time.Second
)

// waitPath resolves a transport route before link establishment.
//
// Discovery is bounded to at most two path-request packets per call:
// PrepareFreshPathRequest once up front (skipped when the cached route is
// still valid), then a single NudgePathRequest after PathRequestMI if the
// route is still missing. The poll loop only checks HasPath.
func waitPath(ctx context.Context, tr *transport.Transport, destHash []byte, total time.Duration) error {
	if tr == nil || len(destHash) != 16 {
		return errInvalidPathDestination
	}
	if total <= 0 {
		total = pathWaitDefault
	}

	switch tr.PrepareFreshPathRequest(destHash) {
	case transport.PrepareFreshInvalidDestination:
		return errInvalidPathDestination
	case transport.PrepareFreshReusedValidPath:
		return nil
	}
	if tr.HasPath(destHash) {
		return nil
	}

	deadline := time.Now().Add(total)
	nudgeAt := time.Now().Add(transport.PathRequestMI)
	nudged := false

	for {
		if tr.HasPath(destHash) {
			return nil
		}

		now := time.Now()
		if !now.Before(deadline) {
			return context.DeadlineExceeded
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !nudged && !now.Before(nudgeAt) {
			_ = tr.NudgePathRequest(destHash)
			nudged = true
		}

		sleep := pathPollInterval
		if remaining := time.Until(deadline); sleep > remaining {
			sleep = remaining
		}
		if !nudged && nudgeAt.After(now) {
			if untilNudge := time.Until(nudgeAt); untilNudge > 0 && untilNudge < sleep {
				sleep = untilNudge
			}
		}

		timer := time.NewTimer(sleep)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func pathWaitError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "path discovery timed out"
	}
	return fmt.Sprintf("%v", err)
}

func transportHops(tr *transport.Transport, destHash []byte, handler *AnnounceHandler, nodeHash string) int {
	if tr != nil {
		if hops := tr.HopsTo(destHash); hops < transport.PathfinderM {
			return int(hops)
		}
	}
	if handler != nil {
		if node, ok := handler.Get(nodeHash); ok {
			return int(node.Hops)
		}
	}
	return -1
}
