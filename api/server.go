package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/itsadijmbt/simple_bank/db/sqlc"
	"github.com/itsadijmbt/simple_bank/db/util"
	"github.com/itsadijmbt/simple_bank/token"
)

// ^ with mock db it is db.store as it is an interface now
// ^ with actual db it is *db.store as it is a struct ptr
type Server struct {
	store db.Store
	//! router help us send api to correct handlder
	router     *gin.Engine
	config     util.Config
	tokenMaker token.Maker
}

// ! NewServer wires together storage, routes, and middleware.

func NewServer(config util.Config, store db.Store) (*Server, error) {

	//* 0 token maker
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("invalid token %w", err)
	}

	//* 1. Allocate the application struct.
	//*    The struct keeps shared dependencies (DB, config, logger, …)
	//*    so handlers can access them through `server.<field>`.
	server := &Server{
		//adding config feild
		config: config,
		store:  store,

		// adding a token maker in this
		tokenMaker: tokenMaker,
	}

	//* 2. Build the Gin engine with sensible defaults:
	//*      • Logger   – writes an access log for every request
	//*      • Recovery – turns panics into 500 JSON errors instead of crashing
	//*    You can swap this for `gin.New()` if you want full manual control.
	router := gin.Default()

	//! to get the validoter engine gin is using type assertion
	//^ I expect that the object returned by .Engine() is of type *validator.Validate
	//^  (i.e., a pointer to validator.Validate struct), so please try to extract it as that."
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	//! ***************************************

	//* 3. Register route handlers.
	//*    gin.Engine <method>(<path>, <handler fn>) stores the mapping
	//*    in a radix tree for O(len(path)) lookup at runtime.
	//*    Path segments that start with `:` become URI parameters.
	//*-------------------------------------------------------------
	//* POST /accounts        → createAccount(ctx *gin.Context)
	router.POST("/accounts", server.createAccount)

	//* GET  /accounts/:id    → getAccount(ctx *gin.Context)
	//*      The `:id` token is read with ctx.Param("id") or via ShouldBindUri.
	router.GET("/accounts/:id", server.getAccount)

	router.GET("/accounts", server.listAccount)

	router.POST("/transfers", server.createTransfer)

	router.POST("/users", server.createUser)

	//* 4. Attach the configured router back to the server struct
	//*    so `main.go` can call `server.router.Run(addr)`.
	server.router = router

	return server, nil
}

// * gin.h is an map[string] interface!

// * Why we need a PUBLIC START function
// && since router feild is pricvate it cant be accessed outiside api package
// ! starts the http server and listen req at given address recievs and address and returns and errror
func (server *Server) Start(address string) error {

	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
