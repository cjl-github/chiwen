// internal/data/model/tty_token.go
package model

import "time"

type TTYToken struct {
	ID           int64      `db:"id"`
	Token        string     `db:"token"`
	UserID       string     `db:"user_id"`
	AssetID      string     `db:"asset_id"`
	TerminalCols int        `db:"terminal_cols"`
	TerminalRows int        `db:"terminal_rows"`
	Status       string     `db:"status"` // pending, used, expired
	ExpireAt     time.Time  `db:"expire_at"`
	CreatedAt    time.Time  `db:"created_at"`
	UsedAt       *time.Time `db:"used_at"`
}
