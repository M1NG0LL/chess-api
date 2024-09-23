package account

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var db *gorm.DB
var jwtKey = []byte("secret_key")

// JWT Claims structure
type Claims struct {
	ID        string `gorm:"primaryKey"`
	IsActive  bool `gorm:"default:false"`
	IsAdmin   bool `gorm:"default:false"`

	jwt.StandardClaims
}

func Login(c *gin.Context) {
	var loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	
	var account Account
	if err := db.Where("username = ?", loginRequest.Username).First(&account).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(loginRequest.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if !account.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account not active"})
		return
	}

	expirationTime := time.Now().Add(15 * time.Hour)
	claims := &Claims{
		ID:       account.ID,
		IsActive: account.IsActive,
		IsAdmin:  account.IsAdmin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}


// Middleware to verify JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the Authorization header
		tokenString := c.GetHeader("Authorization")
		if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header not provided or malformed"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix from the token string
		tokenString = tokenString[7:]

		// Parse the token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		// Check if the token is valid
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set the account details into the context for further use
		c.Set("accountID", claims.ID)
		c.Set("isActive", claims.IsActive)
		c.Set("isAdmin", claims.IsAdmin)

		// Proceed to the next handler
		c.Next()
	}
}