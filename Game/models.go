package game

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type StringArray []string

// Value implements the driver.Valuer interface (for storing into DB)
func (a StringArray) Value() (driver.Value, error) {
	return json.Marshal(a) // Convert StringArray to JSON
}

// Scan implements the sql.Scanner interface (for retrieving from DB)
func (a *StringArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert value to byte array")
	}

	return json.Unmarshal(bytes, a) // Convert JSON to StringArray
}

type Game struct {
	ID        string   `json:"id" gorm:"primary_key"`

	Player1ID string   `json:"player1_id"`
	Player2ID string   `json:"player2_id"`

	Moves     StringArray `json:"moves" gorm:"type:json"`

	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	
    Status    string    `json:"status"`
	
    GameTime  int       `json:"game_time"`
	GameType  string    `json:"game_type"`
}