package game

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init(database *gorm.DB) {
	db = database
}

// POST
// Func to create Game
func CreateGame(c *gin.Context) {
    var game Game

    if err := c.ShouldBindJSON(&game); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

	game.GameID = uuid.New().String()

    game.StartTime = time.Now()

    game.GameStatus = "ongoing" 
    game.Moves = " "

    if err := db.Create(&game).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, game)
}

// POST
// Func to Add Move to the game 
// req = id(from the url)
func MakeMove(c *gin.Context) {
	GameID := c.Param("id")

	type Info struct {
		Move 		string 		`json:"move"`
	}

	var input Info
	if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

	var game Game
	if err := db.Where("game_id= ?", GameID).First(&game).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Game isn't found"})
		return
	}

	game.MovesNum += 1
	game.Moves += input.Move

	c.JSON(http.StatusCreated, gin.H{"message": "Move has been added"})
}

// PUT
// Func to Edit game status
func EndGame(c *gin.Context) {
	GameID := c.Param("id")

	type Info struct {
		Status 		string 		`json:"status"`
	}

	var input Info
	if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

	var game Game
	if err := db.Where("game_id= ?", GameID).First(&game).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Game isn't found"})
		return
	}

	game.GameStatus = input.Status

	if err := db.Model(&game).Update("GameStatus", input.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save reset code"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Game's Status has been edited"})
}

// GET
// Get All Games of you using the token
func GetMyGames(c *gin.Context) {
    accountID, ID_exists := c.Get("accountID")
    isAdmin, Admin_exists := c.Get("isAdmin")
	
	if !ID_exists || !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if isAdmin.(bool) {
		GetGames(c)
		return
	}

    var games []Game
    if err := db.Where("player1_id = ? OR player2_id = ?", accountID, accountID).Find(&games).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve games"})
        return
    }

    c.JSON(http.StatusOK, games)
}

// Get All Games If you were Admin
func GetGames(c *gin.Context) {
	var games []Game
	if err := db.Find(&games).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve Games"})
		return
	}

	c.JSON(http.StatusOK, games)
}

// DELETE
// Delete any Game if you were Admin
func DeleteGame(c *gin.Context) {
	GameID := c.Param("id")

	isAdmin, Admin_exists := c.Get("isAdmin")

	if !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if !isAdmin.(bool) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This url is for ADMIN ONLY."})
		return
	}

	var game Game
	if err := db.First(&game, "game_id = ?", GameID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	if err := db.Delete(&game).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete game"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game deleted successfully"})
}