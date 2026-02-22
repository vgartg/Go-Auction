package models

import (
	"time"
)

type LotStatus string

const (
	LotStatusActive   LotStatus = "active"
	LotStatusClosed   LotStatus = "closed"
	LotStatusCanceled LotStatus = "canceled"
)

type Lot struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	StartPrice    float64   `json:"start_price"`
	MinStep       float64   `json:"min_step"`
	CurrentPrice  float64   `json:"current_price"`
	Status        LotStatus `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	ClosingAt     time.Time `json:"closing_at"`
	Version       int       `json:"version"`
	WinnerID      *string   `json:"winner_id,omitempty"`
	ExtendedCount int       `json:"extended_count"`
}

type Bid struct {
	ID        string    `json:"id"`
	LotID     string    `json:"lot_id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserStats struct {
	UserID     string  `json:"user_id"`
	Username   string  `json:"username"`
	BidsCount  int     `json:"bids_count"`
	WinsCount  int     `json:"wins_count"`
	TotalSpent float64 `json:"total_spent"`
}
