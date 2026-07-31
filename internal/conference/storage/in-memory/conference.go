package in_memory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/CMAK12/gonference/internal/conference/entity"
)

// conferenceKey builds the key a conference is stored under.
func conferenceKey(id string) string {
	return "conference:" + id
}

func (c *Storage) CreateConference(ctx context.Context, conf *entity.Conference) error {
	payload, err := json.Marshal(conf)
	if err != nil {
		return fmt.Errorf("marshal conference %s: %w", conf.ID, err)
	}

	if err := c.client.Set(ctx, conferenceKey(conf.ID), string(payload), 0); err != nil {
		return fmt.Errorf("create conference %s: %w", conf.ID, err)
	}

	return nil
}

func (c *Storage) JoinConference(ctx context.Context, conf *entity.Conference) error {
	payload, err := c.client.Get(ctx, conferenceKey(conf.ID))
	if err != nil {
		return fmt.Errorf("join conference %s: %w", conf.ID, err)
	}

	if err := json.Unmarshal([]byte(payload), conf); err != nil {
		return fmt.Errorf("unmarshal conference %s: %w", conf.ID, err)
	}

	return nil
}
