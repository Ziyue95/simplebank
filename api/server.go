package api

import (
	db "db.sqlc.dev/app/db/sqlc"
	"github.com/gin-gonic/gin"
)

// Server serves HTTP requirests for our banking service.
type Server struct {
	store *db.Store // allow us to interact with the database when processing API requests from clients,
	// see store.go for Store struct
	router *gin.Engine // send each API request to the correct handler for processing
}

// NewServer creates a new Server instance, and setup all HTTP API routes for our service on that server.
func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// add routes to router
	router.POST("/accounts", server.createAccount)
	// add a : before id to tell Gin that id is a URI parameter
	router.GET("/accounts/:id", server.getAccount)

	server.router = router
	return server
}

// Start function runs the HTTP server on the input address to start listening for API requests
// It takes an address as input and return an error
func (server *Server) Start(address string) error {
	// server.router field is private, so it cannot be accessed from outside of this api package
	return server.router.Run(address)

	// TODO: add some graceful shutdown logics
}

// errorResponse converts error msg into a key-value object that Gin can serialize to JSON before returning to the client
// gin.H object is a shortcut for map[string]interface{} to store key-value pairs of any types
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
