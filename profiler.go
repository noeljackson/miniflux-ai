package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	miniflux "miniflux.app/v2/client"
)

type Profiler struct {
	client     *miniflux.Client
	mu         sync.RWMutex
	profile    string
	lastUpdate time.Time
	maxAge     time.Duration
}

func NewProfiler(client *miniflux.Client) *Profiler {
	return &Profiler{
		client: client,
		maxAge: 30 * time.Minute,
	}
}

// Profile returns the cached user preference profile, refreshing if stale.
func (p *Profiler) Profile() string {
	p.mu.RLock()
	if time.Since(p.lastUpdate) < p.maxAge && p.profile != "" {
		defer p.mu.RUnlock()
		return p.profile
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock.
	if time.Since(p.lastUpdate) < p.maxAge && p.profile != "" {
		return p.profile
	}

	profile, err := p.buildProfile()
	if err != nil {
		log.Printf("profiler: failed to refresh: %v", err)
		return p.profile // return stale profile on error
	}

	p.profile = profile
	p.lastUpdate = time.Now()
	log.Printf("profiler: refreshed (%d chars)", len(profile))
	return p.profile
}

func (p *Profiler) buildProfile() (string, error) {
	result, err := p.client.Entries(&miniflux.Filter{
		Starred:   miniflux.FilterOnlyStarred,
		Limit:     50,
		Order:     "published_at",
		Direction: "desc",
	})
	if err != nil {
		return "", fmt.Errorf("fetch starred entries: %w", err)
	}

	if len(result.Entries) == 0 {
		return "", nil
	}

	var lines []string
	for _, entry := range result.Entries {
		line := fmt.Sprintf("- %s", entry.Title)
		if entry.Feed != nil {
			line += fmt.Sprintf(" [%s]", entry.Feed.Title)
		}
		lines = append(lines, line)
	}

	return fmt.Sprintf("User's %d most recent starred (favorite) articles:\n%s",
		len(result.Entries), strings.Join(lines, "\n")), nil
}
