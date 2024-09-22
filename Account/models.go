package account

import (
	"time"
)

type Account struct {
	ID		  string 		`gorm:"primaryKey"`
	Username  string 		`gorm:"unique;not null"`
	Email     string 		`gorm:"unique;not null"`
	Password  string

	code 	  string		`gorm:"default:' '"`

	StartDay time.Time		
	BulletElo int			`gorm:"default:200"`
	BlitzElo int			`gorm:"default:200"`
	RapidElo int			`gorm:"default:200"`

	ActivationToken string    `json:"activation_token"`
	TokenExpiresAt  time.Time `json:"token_expires_at"`

	IsActive  bool 			`gorm:"default:false"`

	IsAdmin   bool 			`gorm:"default:false"`
}