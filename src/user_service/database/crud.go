package database

import (
	"context"

	"github.com/google/uuid"
)

// Get user by email from DB
func GetUserByEmailDB(c *Config, ctx context.Context, email string) (*User, error) {
	user, err := c.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Get user by ID from DB
func GetUserByIDFromDB(c *Config, ctx context.Context, userID uuid.UUID) (*User, error) {
	user, err := c.Queries.GetUserById(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update user details in DB
func UpdateUserDetailDB(c *Config, ctx context.Context, params UpdateUserDetailsParams) (*User, error) {
	// Begin DB transaction
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := c.Queries.WithTx(tx)

	// Update user details
	user, err := qtx.UpdateUserDetails(ctx, params)
	if err != nil {
		return nil, err
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &user, nil
}

// Update user password in DB
func UpdateUserPasswordDB(c *Config, ctx context.Context, params UpdateUserPasswordParams) error {
	// Begin DB transaction
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := c.Queries.WithTx(tx)

	// Update user password
	if err := qtx.UpdateUserPassword(ctx, params); err != nil {
		return err
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

// Delete user from DB
func DeleteUserDB(apiCfg *Config, ctx context.Context, userID uuid.UUID) error {
	// Begin DB transaction
	tx, err := apiCfg.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := apiCfg.Queries.WithTx(tx)

	if err := qtx.DeleteUser(ctx, userID); err != nil {
		return err
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}
