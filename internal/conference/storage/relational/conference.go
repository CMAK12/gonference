package relational

import (
	"context"
	"errors"
	"fmt"

	"github.com/CMAK12/gonference/internal/conference/storage"
	"github.com/jackc/pgx/v5"

	"github.com/CMAK12/gonference/internal/conference/entity"
)

type Conference struct {
	db Querier
}

func NewConferenceStorage(db Querier) *Conference {
	return &Conference{
		db: db,
	}
}

func (c *Conference) CreateConference(ctx context.Context, conf *entity.Conference) error {
	const query = `
		INSERT INTO conferences
			(id, name, creator_id, invited_members, token, start_time, end_time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := c.db.Exec(ctx, query,
		conf.ID, conf.Name, conf.CreatorID, conf.InvitedMembers,
		conf.Token, conf.StartTime, conf.EndTime, conf.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create conference %s: %w", conf.ID, err)
	}

	return nil
}

func (c *Conference) JoinConference(ctx context.Context, conf *entity.Conference) error {
	const query = `
		SELECT id, name, creator_id, invited_members, token, start_time, end_time, created_at
		FROM conferences
		WHERE id = $1`

	err := c.db.QueryRow(ctx, query, conf.ID).Scan(
		&conf.ID, &conf.Name, &conf.CreatorID, &conf.InvitedMembers,
		&conf.Token, &conf.StartTime, &conf.EndTime, &conf.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return storage.ErrConferenceNotFound
		}
		return fmt.Errorf("join conference %s: %w", conf.ID, err)
	}

	return nil
}
