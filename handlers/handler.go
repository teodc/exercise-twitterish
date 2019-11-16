package handlers

import (
	"github.com/globalsign/mgo"
)

// Handler struct
type (
	Handler struct {
		DB *mgo.Session
	}
)

const (
	// Key should be imported from .env
	Key = "secret"
)
