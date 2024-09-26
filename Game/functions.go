package game

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/notnil/chess"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init(database *gorm.DB) {
	db = database
}

// POST
// Create a game 
func CreateGame(c *gin.Context) {
    var input struct {
        Player1ID string `json:"player1_id"`
        Player2ID string `json:"player2_id"`
        GameType  string `json:"game_type"`
        GameTime  int `json:"game_time"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

     
    if input.Player1ID == input.Player2ID { 
        c.JSON(http.StatusBadRequest, gin.H{"error": "Player1 and Player2 cannot be the same"})
        return
    }

    if input.GameType != "blitz" && input.GameType != "bullet" && input.GameType != "classic" { 
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Game Type"})
        return
    }

    newGame := Game{
		ID: uuid.New().String(),
        Player1ID: input.Player1ID,
        Player2ID: input.Player2ID,
        StartTime: time.Now(),
        GameType:  input.GameType,
        GameTime:  input.GameTime,
        Status:    "ongoing",
    }

    if err := db.Create(&newGame).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
	
    c.JSON(http.StatusOK, gin.H{"game": newGame})
}

// PUT
// Func to Add Move to the game 
// req = id(from the url)
func EndGame(c *gin.Context) {
	gameID := c.Param("id")

    var game Game
    if err := db.First(&game, "id = ?", gameID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
        return
    }

    if game.Status == "ongoing" {
        endTime := time.Now()
        game.Status = "completed" // Or determine if it was a draw, win, etc.
        game.EndTime = endTime

		if err := db.Save(&game).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Game ended"})
    } else {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Game already ended"})
    }
}

// POST
// Func to Add Move to the game 
func MakeMove(c *gin.Context) {
	gameID := c.Param("id")

    var game Game
    if err := db.First(&game, "id = ?", gameID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
        return
    }

    if game.Status != "ongoing" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Game already ended"})
    }

    var input struct {
        Move string `json:"move"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Initialize a new chess game
    chessGame := chess.NewGame()

    // Replay all previous moves from the game
    for _, moveStr := range game.Moves {
        move, err := chess.AlgebraicNotation{}.Decode(chessGame.Position(), moveStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid previous move: " + moveStr})
            return
        }
        chessGame.Move(move)
    }

    // Apply the new move
    newMove, err := chess.AlgebraicNotation{}.Decode(chessGame.Position(), input.Move)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid move format"})
        return
    }

    if err := chessGame.Move(newMove); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Illegal move"})
        return
    }

    game.Moves = append(game.Moves, input.Move)
	
    if err := db.Save(&game).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

    c.JSON(http.StatusOK, gin.H{"game": game})
}

// GET
// Get all moves of one game 
func GetMoves(c *gin.Context) {
    var game Game
    if err := db.First(&game, "id = ?", c.Param("id")).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"moves": game.Moves})
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
        c.JSON(http.StatusNotFound, gin.H{"error": "No games found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"games": games})
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
	if err := db.First(&game, "id = ?", GameID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	if err := db.Delete(&game).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete game"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game deleted successfully"})
}