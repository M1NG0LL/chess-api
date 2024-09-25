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
		TeamID:   team.ID,          
        TeamName: team.Name,        
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

	var memberToRemove Member
	found := false
	for _, member := range team.Members {
		if member.ID == accountID {
			memberToRemove = member
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found in the team"})
		return
	}

	if err := db.Model(&team).Association("Members").Delete(&memberToRemove); err != nil { 
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

    c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully", "team": team})
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
// Get Team info based on your token
func GetTeamsByAccountID(c *gin.Context) {
    accountID, ID_exists := c.Get("accountID")

    if !ID_exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    var teams []Team
    if err := db.Preload("Members").Joins("JOIN team_members ON team_members.team_id = teams.id").
        Where("team_members.member_id = ?", accountID).Find(&teams).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve teams"})
        return
    }

    if len(teams) == 0 {
        c.JSON(http.StatusNotFound, gin.H{"message": "No teams found for the user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"teams": teams})
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

	if err := db.Where("team_id = ?", teamID).Delete(&Member{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove team members"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team deleted successfully"})
}