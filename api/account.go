package api

import (
	"database/sql"
	"errors"
	"fmt"

	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/itsadijmbt/simple_bank/db/sqlc"
	"github.com/itsadijmbt/simple_bank/token"
	"github.com/lib/pq"
)

//& why we have to create a seperate user for each test

//! function returns a handler to the POST routte

//	type createAccountRequest struct {
//		Owner string `json:"owner" binding:"required" `
//		 Balance  int64  `json:"balance"`
//		^ binding has conditons inside "  con1 , con2 , con3 " i.e comma seperated conditions
//		Currency string `json:"currency" binding:"required, oneof = USD EUR"  `
//	}
type createAccountRequest struct {
	Owner string `json:"owner"    binding:"required"`
	//!we use "currency" as it as under gin default validator playground Engine
	Currency string `json:"currency" binding:"required,currency"`
}

func (server *Server) createAccount(ctx *gin.Context) {

	var req createAccountRequest

	//* it checks whether json is correct and if correct the actual result obj
	if err := ctx.ShouldBindJSON(&req); err != nil {
		//* JSON OBJECT WITH INTERFACE + REQUEST IS SENT BACK
		log.Println("Validation error:", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	//! ******************************AUTHORISATION PART*****************************

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := server.store.CreateAccount(ctx, arg)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
			}
			log.Println(pqErr.Code.Name())
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	//^ account created passed and correct obj is passed
	ctx.JSON(http.StatusOK, account)

}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccount(ctx *gin.Context) {

	var req getAccountRequest

	//!check if the binded json[header for id] is coorect or not
	//
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//! if we remove this block the api test will fail as we GETACCOUNT funciton in API_TESTING!
	accounts, err := server.store.GetAccount(ctx, req.ID)

	if err != nil {

		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))

		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))

	}
	//^ check if the user has the authorization to recive it
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	if authPayload.Username != accounts.Owner {
		err := errors.New("account does not belong to authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)

}

type listAccountRequest struct {
	PageID   int32 `form:"page_id"  binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=1,max=10"`
}

func (server *Server) listAccount(ctx *gin.Context) {

	var req listAccountRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//! this is api auth
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	//! an owner feild was added to safeguard authorization
	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	accounts, err := server.store.ListAccounts(ctx, arg)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
	fmt.Println(accounts)

	//& we might get null because of any empty var items []Account so to solve this
	//& in latest sqlc.yaml we use emit_empty_slice=true;
	//* we dont need this 	accounts = []db.Account{}

}
