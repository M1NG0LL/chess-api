package game

import (
	"time"
)

type Game struct {
    ID         string           `gorm:"primaryKey"`

    Player1ID  string
    Player2ID  string

    Moves      []string       `gorm:"type:text[]"`

    StartTime  time.Time
    EndTime    time.Time

    Status     string         // "win", "draw", "ongoing"
	
    GameType   string         // e.g., "blitz", "rapid"
    GameTime   int            // in seconds
}