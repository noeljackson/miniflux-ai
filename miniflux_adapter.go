package main

import (
	miniflux "miniflux.app/v2/client"
)

type MinifluxUpdater struct {
	client *miniflux.Client
}

func NewMinifluxUpdater(url, apiKey string) *MinifluxUpdater {
	return &MinifluxUpdater{
		client: miniflux.NewClient(url, apiKey),
	}
}

func (m *MinifluxUpdater) UpdateEntryContent(entryID int64, content string) error {
	_, err := m.client.UpdateEntry(entryID, &miniflux.EntryModificationRequest{
		Content: miniflux.SetOptionalField(content),
	})
	return err
}

func (m *MinifluxUpdater) StarEntry(entryID int64) error {
	return m.client.ToggleStarred(entryID)
}
