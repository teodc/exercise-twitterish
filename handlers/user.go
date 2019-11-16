package handlers

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
	"github.com/teodc/twitterish/models"
)

// Signup a new user
func (handler *Handler) Signup(context echo.Context) (err error) {
	// Bind a new user model to request body data
	user := &models.User{
		ID: bson.NewObjectId(),
	}
	if err = context.Bind(user); err != nil {
		return
	}

	// Validate user data
	// An email address and a password are required
	if user.Email == "" || user.Password == "" {
		return &echo.HTTPError{
			Code:    http.StatusUnprocessableEntity,
			Message: "Invalid email or password",
		}
	}

	// Persist user data
	db := handler.DB.Clone()
	defer db.Close()
	collection := db.DB("twitter").C("users")
	if err = collection.Insert(user); err != nil {
		return
	}

	return context.JSON(http.StatusCreated, user)
}

// Login authenticates a user
func (handler *Handler) Login(context echo.Context) (err error) {
	// Bind a new user model to the request body data
	user := new(models.User)
	if err = context.Bind(user); err != nil {
		return
	}

	// Fetch user from the database
	db := handler.DB.Clone()
	defer db.Close()
	collection := db.DB("twitter").C("users")
	if err = collection.Find(bson.M{"email": user.Email, "password": user.Password}).One(user); err != nil {
		if err == mgo.ErrNotFound {
			return &echo.HTTPError{
				Code:    http.StatusUnauthorized,
				Message: "User not found",
			}
		}
		return
	}

	// Create JWT token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["expires_at"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token and send it as response
	if user.Token, err = token.SignedString([]byte(Key)); err != nil {
		return err
	}

	user.Password = "*****" // Don't send the password

	return context.JSON(http.StatusOK, user)
}

// Follow allows a user to follow another one
func (handler *Handler) Follow(context echo.Context) (err error) {
	userID := userIDFromToken(context)
	id := context.Param("id")

	// Add a follower to the user
	db := handler.DB.Clone()
	defer db.Close()
	collection := db.DB("twitter").C("users")
	if err = collection.UpdateId(bson.ObjectIdHex(id), bson.M{"$addToSet": bson.M{"followers": userID}}); err != nil {
		if err == mgo.ErrNotFound {
			return echo.ErrNotFound
		}
	}

	return
}

func userIDFromToken(context echo.Context) string {
	user := context.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims["id"].(string)
}
