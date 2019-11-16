package handlers

import (
	"net/http"
	"strconv"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
	"github.com/teodc/twitterish/models"
)

// WritePost creates a new post
func (handler *Handler) WritePost(context echo.Context) (err error) {
	// Bind sent post data
	user := &models.User{
		ID: bson.ObjectIdHex(userIDFromToken(context)),
	}
	post := &models.Post{
		ID:   bson.NewObjectId(),
		From: user.ID.Hex(),
	}
	if err = context.Bind(post); err != nil {
		return
	}

	// Validate sent post data
	// A recipient and a message body are required
	if post.To == "" || post.Message == "" {
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Invalid recipient or message body",
		}
	}

	// Fetch user from database
	db := handler.DB.Clone()
	defer db.Close()
	collection := db.DB("twitter").C("users")
	if err = collection.FindId(user.ID).One(user); err != nil {
		if err == mgo.ErrNotFound {
			return echo.ErrNotFound
		}
		return
	}

	// Save post in database
	if err = db.DB("twitter").C("posts").Insert(post); err != nil {
		return
	}

	return context.JSON(http.StatusCreated, post)
}

// ListPosts fetches existing posts
func (handler *Handler) ListPosts(context echo.Context) (err error) {
	userID := userIDFromToken(context)
	page, _ := strconv.Atoi(context.QueryParam("page"))
	limit, _ := strconv.Atoi(context.QueryParam("limit"))

	// Pagination defaults
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 100
	}

	// Retrieve posts from database
	posts := []*models.Post{}
	db := handler.DB.Clone()
	defer db.Close()
	collection := db.DB("twitter").C("posts")
	if err = collection.Find(bson.M{"to": userID}).Skip((page - 1) * limit).Limit(limit).All(&posts); err != nil {
		return
	}

	return context.JSON(http.StatusOK, posts)
}
