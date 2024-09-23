package main

import (
	"log"
	account "project/Account"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	router := gin.Default()

	ERR := godotenv.Load()
    if ERR != nil {
        log.Fatalf("Error loading .env file")
    }
	
	var err error
	db, err = gorm.Open(sqlite.Open("chess-api.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate 
	if err := db.AutoMigrate(&account.Account{}); err != nil {
		panic("failed to migrate database")
	}

	// Account Part ===================================================
	account.Init(db)

	router.POST("/login", account.Login)
	router.POST("/accounts", account.CreateAccount)

		// Account Activation part
	router.GET("/activate", account.ActivateAccount)

		// Pass Reset part
	router.POST("/passreset", account.ForgetPass)
	router.PUT("/update-password", account.UpdatingPassword)

	protected := router.Group("/")
	protected.Use(account.AuthMiddleware())

	protected.GET("/accounts", account.GetMyAccount)
	protected.PUT("/accounts/:id", account.UpdateAccountByID)
	protected.DELETE("/accounts/:id", account.DeleteAccountbyid)


	// Game Part =======================================================

	

	router.Run(":8081")
}