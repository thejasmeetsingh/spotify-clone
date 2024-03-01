package database

import (
	"context"

	"github.com/google/uuid"
)

// Add content into DB
func AddContentDB(c *Config, ctx context.Context, params AddContentParams) (*Content, error) {
	// Begin DB transaction
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := c.Queries.WithTx(tx)

	// add content into DB
	content, err := qtx.AddContent(ctx, params)
	if err != nil {
		return nil, err
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &content, nil
}

// Update content details
func UpdateContentDetailDB(c *Config, ctx context.Context, params UpdateContentDetailsParams) (*Content, error) {
	// Begin DB transaction
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := c.Queries.WithTx(tx)

	// update content details
	content, err := qtx.UpdateContentDetails(ctx, params)
	if err != nil {
		return nil, err
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &content, nil
}

// Update content s3 key
func UpdateContentS3KeyDB(c *Config, ctx context.Context, params UpdateS3KeyParams) error {
	// Begin DB transaction
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := c.Queries.WithTx(tx)

	// update content s3 key
	if err := qtx.UpdateS3Key(ctx, params); err != nil {
		return err
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

// Delete content from DB
func DeleteContentDB(c *Config, ctx context.Context, params DeleteContentParams) error {
	// Begin DB transaction
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := c.Queries.WithTx(tx)

	// delete content from DB
	if err := qtx.DeleteContent(ctx, params); err != nil {
		return err
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

// Get content detail by ID from DB
func GetContentDetailDB(c *Config, ctx context.Context, contentID uuid.UUID) (*Content, error) {
	content, err := c.Queries.GetContentById(ctx, contentID)
	if err != nil {
		return nil, err
	}
	return &content, nil
}

// Get content posted by a user
func GetUserContentDB(c *Config, ctx context.Context, params GetUserContentParams) ([]interface{}, error) {
	contents, err := c.Queries.GetUserContent(ctx, params)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

// Get contents from DB
func GetContentListDB(c *Config, ctx context.Context, params GetContentListParams) ([]interface{}, error) {
	contents, err := c.Queries.GetContentList(ctx, params)
	if err != nil {
		return nil, err
	}
	return contents, nil
}
