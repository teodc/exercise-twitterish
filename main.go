package main

import (
	"github.com/globalsign/mgo"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/teodc/twitterish/handlers"
)

func main() {
	// Configure logging and middleware
	app := echo.New()
	app.Logger.SetLevel(log.ERROR)
	app.Use(middleware.Logger())
	app.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(handlers.Key),
		Skipper: func(context echo.Context) bool {
			// Skip authentication for signup and login
			if context.Path() == "/login" || context.Path() == "/signup" {
				return true
			}
			return false
		},
	}))

	// Database connection
	db, err := mgo.Dial("mongo:27017")
	if err != nil {
		app.Logger.Fatal(err)
	}

	// Create collection indices
	if err = db.Copy().DB("twitter").C("users").EnsureIndex(mgo.Index{
		Key:    []string{"email"},
		Unique: true,
	}); err != nil {
		log.Fatal(err)
	}

	// Initialize request handler
	handler := &handlers.Handler{DB: db}

	// Define routes
	app.POST("/signup", handler.Signup)
	app.POST("/login", handler.Login)
	app.POST("/follow/:id", handler.Follow)
	app.POST("/publish", handler.WritePost)
	app.GET("/feed", handler.ListPosts)

	// Start HTTP server
	app.Logger.Fatal(app.Start(":8080"))
}
