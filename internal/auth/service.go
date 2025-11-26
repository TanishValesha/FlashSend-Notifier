package auth

import (
	"errors"
	"time"

	"github.com/TanishValesha/FlashSend-Notifier/internal/config"
	"github.com/TanishValesha/FlashSend-Notifier/internal/db"
	"github.com/TanishValesha/FlashSend-Notifier/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(config.Cfg.JwtSecret)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(hash), err
}

func CheckPassword(hashedPassword string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func GenerateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func RegisterUser(email, password string) (*models.User, error) {
	var existing models.User
	err := db.DB.Where("email = ?", email).First(&existing).Error
	if err == nil {
		return nil, errors.New("Email Already Registered")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Email:    email,
		Password: hash,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func LoginUser(email, password string) (*models.User, error) {
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("User Not Found")
	}

	if !CheckPassword(user.Password, password) {
		return nil, errors.New("Invalid Password")
	}

	return &user, nil
}
