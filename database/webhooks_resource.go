package database

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
)

type WebhooksResource interface {
	GetWebhooksID(ctx context.Context) (string, error)
	SetWebhooksID(ctx context.Context, webhooksID string) error
}

func (c *connection) GetWebhooksID(ctx context.Context) (string, error) {
	query := "SELECT id FROM webhooks"
	var webhooksID string
	err := c.db.GetContext(ctx, &webhooksID, query)
	if err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", errors.Wrap(err, "failed querying webhooks id from database")
	}
	return webhooksID, nil
}

func (c *connection) SetWebhooksID(ctx context.Context, webhooksID string) error {
	query := "INSERT INTO webhooks (id) VALUES ($1)"
	_, err := c.db.ExecContext(ctx, query, webhooksID)
	if err != nil {
		return errors.Wrap(err, "failed inserting webhooks id into database")
	}
	return nil
}
