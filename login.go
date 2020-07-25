package videodir

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}

func (srv *AppServer) Login(c *fiber.Ctx) {
	body := c.Body()
	//log.Printf("body: %s", body)
	var user UserCredentials
	err := json.Unmarshal([]byte(body), &user)
	if err != nil {
		srv.Logger.Errorf("Error unmarshal body: %s %s", body, err.Error())
		c.Status(fiber.StatusBadRequest)
		return
	}

	// validate username and password
	if !srv.validate(&user) {
		srv.Logger.Errorf("Invalid user: %s", user.Username)
		c.Status(fiber.StatusUnauthorized)
		return
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS384)
	claims := make(jwt.MapClaims)
	claims["username"] = user.Username
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(8)).Unix()
	claims["iat"] = time.Now().Unix()
	token.Claims = claims

	tokenString, err := token.SignedString([]byte(srv.Config.JwtSecret))
	if err != nil {
		srv.Logger.Errorf("Error while signing the token: %s", err.Error())
		c.Status(fiber.StatusInternalServerError)
		return
	}

	_ = c.JSON(Token{tokenString})
}

func (srv *AppServer) validate(user *UserCredentials) bool {
	if hash, found := srv.passwords[user.Username]; found {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(user.Password))
		if err != nil {
			return false
		}
		return true
	}
	return false
}
