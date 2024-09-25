package team

import (
	"time"
)

type Team struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name"`

    LeaderID  string    `json:"leader_id"`
    LeaderName string   `json:"leader_name"`
    
	Members   []Member  `json:"members" gorm:"many2many:team_members"`
    
	StartDate time.Time `json:"start_date"`
}

type Member struct {
    ID       string `json:"id" gorm:"primaryKey"`
    Username string `json:"username"`

    TeamID        string    `json:"team_id"`
    TeamName      string    `json:"team_name"`
}