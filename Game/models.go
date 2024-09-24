package game

import (
	"time"
)

type Game struct {
	ID        string   `json:"id" gorm:"primary_key"`

	Player1ID string   `json:"player1_id"`
	Player2ID string   `json:"player2_id"`

	Moves     []string `json:"moves" gorm:"type:text[]"`

	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	
    Status    string    `json:"status"`
	
    GameTime  int       `json:"game_time"`
	GameType  string    `json:"game_type"`
}