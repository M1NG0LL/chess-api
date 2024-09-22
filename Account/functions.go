package account

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Init(database *gorm.DB) {
	db = database
}

// POST
// Creating New Account
func CreateAccount(c *gin.Context) {
	var account Account

	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isValidEmail(account.Email) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
        return
    }

	var existingAccount Account
	if err := db.Where("username = ? OR email = ?", account.Username, account.Email).First(&existingAccount).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username or Email already exists"})
		return
	}

	if err, error := ValidatePassword(account.Password); err {
		c.JSON(http.StatusInternalServerError, gin.H{"error": error})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.MinCost)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	account.Password = string(hashedPassword)

	account.ID = uuid.New().String()

	layout := os.Getenv("LAYOUT")
	startDate, err := time.Parse(layout, account.StartDay.Format(layout))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid StartDate format"})
		return
	}
	
	account.StartDay = startDate

	account.IsActive = false
	account.IsAdmin = false


	// Making Account's Activation token
	activationToken := uuid.New().String()
	tokenExpiresAt := time.Now().Add(24 * time.Hour) // Token valid for 24 hours
	account.ActivationToken = activationToken
	account.TokenExpiresAt = tokenExpiresAt
	
	// Make activation link 
	activationLink := fmt.Sprintf("https://localhost:8081/activate?token=%s", activationToken)
	message := fmt.Sprintf("Welcome to our app!\n\nPlease activate your account by clicking the following link: %s", activationLink)

	if email_err := SendEmail(account.Email, "Account Activation", message); email_err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send activation email"})
		return
	}

	// Save the account
	if err := db.Create(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Account created. Please check your email to activate your account."})
}

// GET
// Activate account
func ActivateAccount(c *gin.Context) {
	// Get the token from the query parameters
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid activation token"})
		return
	}

	var account Account

	if err := db.Where("activation_token = ?", token).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid or expired activation token"})
		return
	}

	if time.Now().After(account.TokenExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Activation token has expired"})
		return
	}

	account.IsActive = true
	account.ActivationToken = ""

	if err := db.Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account successfully activated"})
}

// POST
// Func if you forget the password
func ForgetPass(c *gin.Context) {
	type Info struct {
		Email     string 		`gorm:"unique;not null"`
	}

	var input Info
	var account Account

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Where("email = ?",input.Email).First(&account).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email doesn't exist"})
		return
	}

	Reset_Code := GenerateCode(6)
	account.code = Reset_Code

	passresetLink := fmt.Sprintf("https://localhost:8081/update-password?id=%s&code=%s",account.ID, account.code)
	
	message := fmt.Sprintf("Welcome to our app!\n\nIf you requested password reset click on the following link: %s", passresetLink)

	if err := SendEmail(account.Email, "Password Reset", message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send activation email"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Please check your email to change your password."})
}

// PUT
// Func to update password from url in email
func UpdatingPassword(c *gin.Context) {
	accountID := c.Query("id")
	code := c.Query("code")

	type Pass_reset struct {
		Password  string
	}

	var account Account
	var input Pass_reset

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Where("id = ? and code = ?", accountID, code).First(&account).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "ID or code don't exist"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.MinCost)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	account.Password = string(hashedPassword)

	if err := db.Model(&Account{}).Where("id = ?", accountID).Updates(account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// GET
// get Account info using the Token
func GetMyAccount(c *gin.Context) {
	accountID, ID_exists := c.Get("accountID")
	isAdmin, Admin_exists := c.Get("isAdmin")
	
	if !ID_exists || !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if isAdmin.(bool) {
		GetAccounts(c)
		return
	}

	var account Account
	if err := db.First(&account, "id = ?", accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}
	
	c.JSON(http.StatusOK, account)
}

// get Accounts if you are admin using the same url of getMyAccount()
func GetAccounts(c *gin.Context) {
	var accounts []Account
	if err := db.Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve accounts"})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// PUT
// Update Account info using the Token
func UpdateMyAccount(c *gin.Context) {
	accountID, ID_exists := c.Get("accountID")
	isAdmin, Admin_exists := c.Get("isAdmin")

	if !ID_exists || !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var preaccount Account
	if err := db.First(&preaccount, "id = ?", accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	var account Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isAdmin.(bool) {
		account.ID = preaccount.ID
		account.Password = preaccount.Password

		account.BlitzElo = preaccount.BlitzElo
		account.BulletElo = preaccount.BulletElo
		account.RapidElo = preaccount.RapidElo

		account.IsActive = preaccount.IsActive
		account.IsAdmin = preaccount.IsAdmin
	}

	if err := db.Model(&preaccount).Updates(account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
		return
	}

	c.JSON(http.StatusOK, account)
}

// PUT
// This func is for ADMINS ONLY
// Update any account by putting id in url 
func UpdateAccountByID(c *gin.Context) {
	isAdmin, Admin_exists := c.Get("isAdmin")

	if  !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	paramID := c.Param("id")
	
	var preaccount Account
	if err := db.First(&preaccount, "id = ?", paramID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	if !isAdmin.(bool) {
		c.Set("accountID", paramID) 
		UpdateMyAccount(c)
		return
	} 
	
	var account Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if account.Password != preaccount.Password {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.MinCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		account.Password = string(hashedPassword)
	}

	if err := db.Model(&Account{}).Where("id = ?", paramID).Updates(account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account updated successfully"})
}

// DELETE
// This func is for ADMINS ONLY
// Delete any account by putting id in url 
func DeleteAccountbyid(c *gin.Context)  {
	isAdmin, Admin_exists := c.Get("isAdmin")

	if !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	paramID := c.Param("id")

	if !isAdmin.(bool) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This url is for ADMIN ONLY."})
		return
	}

	var account Account
	if err := db.First(&account, "id = ?", paramID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	if err := db.Delete(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}