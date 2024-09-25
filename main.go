package main

import (
	"log"

	account "project/Account"
	game "project/Game"
	team "project/Team"

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
	if err := db.AutoMigrate(&account.Account{}, &game.Game{}, &team.Team{}); err != nil {
		panic("failed to migrate database")
	}

	// Initialize account and game packages
	account.Init(db)
	game.Init(db)
	team.Init(db)

	// Account Part ===================================================
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
	// error in finding Games 

	protected.POST("/games", game.CreateGame) // Good
	protected.PUT("/games/:id/end", game.EndGame)

	protected.POST("/games/:id/move", game.MakeMove) // good

	protected.GET("/games/:id/moves", game.GetMoves) // has errors

	protected.GET("/games/my", game.GetMyGames) // Has errors

	protected.DELETE("/games/:id", game.DeleteGame)


	// Team Part  =======================================================
	// models should have more info like name of the team := edited successfully

	protected.POST("/teams", team.CreateTeam) // good
	protected.DELETE("/teams/:id", team.DeleteTeam) // edited successfully 
	protected.GET("/teams", team.GetTeams) // good

	protected.POST("/teams/members", team.AddMember) // good
	protected.DELETE("/teams/:id/members", team.RemoveMember)  // edited successfully
	protected.GET("/teams/:id/members", team.GetMembers) // good

	// more function like get the team by the token := edited successfully
	protected.GET("/teams/my", team.GetTeamsByAccountID)


	router.Run(":8081")
}