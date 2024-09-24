package team

import (
	"net/http"
	account "project/Account"
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
// Create a Team
func CreateTeam(c *gin.Context) {
	accountID, ID_exists := c.Get("accountID")
	_, Admin_exists := c.Get("isAdmin")
	
	if !ID_exists || !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var account account.Account
	if err := db.First(&account, "id = ?", accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

    var input struct {
        Name      string `json:"name" binding:"required"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Team name is required"}) 
		return
	}

    newTeam := Team{
        ID:       
		 uuid.New().String(),
        Name:      input.Name,
        LeaderID:  account.ID,
        LeaderName: account.Username,
        StartDate: time.Now(),
    }

    if err := db.Create(&newTeam).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"team": newTeam})
}

// POST
// Adding Member to the team
func AddMember(c *gin.Context) {
	accountID, ID_exists := c.Get("accountID")

	if !ID_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var account account.Account
	if err := db.First(&account, "id = ?", accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	var input struct {
		TeamIdentifier string `json:"team"` 
	}

	if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

	var team Team
    if err := db.Preload("Members").First(&team, "id = ? or name = ?", input.TeamIdentifier, input.TeamIdentifier).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
        return
    }

	newMember := Member{
        ID:       account.ID,
        Username: account.Username,
    }

	team.Members = append(team.Members, newMember)

    if err := db.Save(&team).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"team": team})
}

// DELETE
// Remove Member from the team
func RemoveMember(c *gin.Context) {
	accountID, ID_exists := c.Get("accountID")

	if !ID_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	teamID := c.Param("teamid")

	var team Team
    if err := db.Preload("Members").First(&team, "id = ?", teamID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
        return
    }

	if team.LeaderID != accountID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You aren't Team's Leader"})
		return
	}

	for i := 0; i < len(team.Members); i++ {
		if team.Members[i].ID == accountID {
			team.Members = append(team.Members[:i], team.Members[i+1:]...) // Edited: Remove member by index
			break
		}
	}

    if err := db.Save(&team).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"team": team})
}

// GET
// Get All members in your team
func GetMembers(c *gin.Context) {
    teamID := c.Param("id")

    var team Team
    if err := db.Preload("Members").First(&team, "id = ?", teamID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
		"TeamID": team.ID,
		"TeamName": team.Name,
		"members": team.Members,
	})
}

// GET
// Get All teams 
// Func for ADMINS ONLY
func GetTeams(c *gin.Context) {
	isAdmin , Admin_exists := c.Get("isAdmin")
	
	if !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if !isAdmin.(bool) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This url is for ADMIN ONLY."})
		return
	}
	
    var teams []Team
    if err := db.Preload("Members").Find(&teams).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"teams": teams})
}

// DELETE
// Delete Team
// Func for ADMINS ONLY
func DeleteTeam(c *gin.Context) {
	accountID, ID_exists := c.Get("accountID")
	isAdmin, Admin_exists := c.Get("isAdmin")
	
	if !ID_exists || !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	teamID := c.Param("id")

	var team Team
    if err := db.Preload("Members").First(&team, "id = ?", teamID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
        return
    }

	if !isAdmin.(bool) {
		if accountID != team.LeaderID {
			c.JSON(http.StatusNotFound, gin.H{"error": "You aren't ADMIN of the page nor Leader of the Team"})
			return
		}
	}

	if err := db.Delete(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete team"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team deleted successfully"})
}