package game

import (
	"time"
)

type Game struct {
	GameID 		string 		`gorm:"primaryKey"`

	Player1ID	string
	Player2ID	string

	Moves     	string
	MovesNum	int			`gorm:"default:0"`

	StartTime   time.Time 
    EndTime     time.Time

	GameStatus  string

	GameTime    time.Duration 
    GameType    string
}

