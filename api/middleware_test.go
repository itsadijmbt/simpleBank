package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itsadijmbt/simple_bank/token"
	"github.com/stretchr/testify/require"
)

func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	token, err := tokenMaker.CreateToken(username, duration)

	require.NoError(t, err)

	authHeader := fmt.Sprintf("%s %s", authorizationType, token)

	request.Header.Set(authorizationHeaderKey, authHeader)
}

func TestAuthMiddleware(t *testing.T) {

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Ok",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			//^ CLIENT 	does not provide auth
			//! so we remove the auth block
			name: "NOauth",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "unsupported Auth",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "unsupported", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},

		{
			name: "invalid  Auth Format",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}
	// Run all test cases in table-driven format
	for i := range testCases {
		// Important: create a new variable to avoid closure issue
		tc := testCases[i]

		// Run sub-test for each case using the given name
		t.Run(tc.name, func(t *testing.T) {

			// Create a test server instance
			// This likely includes a gin.Engine, routes, tokenMaker etc.
			server := NewTestServer(t, nil)

			// Define a simple GET endpoint path that will be protected by the middleware
			authPath := "/auth"

			// Register a GET route with the authentication middleware
			// Middleware checks token, and if valid, runs the final handler
			server.router.GET(
				authPath,
				authMiddleware(server.tokenMaker),
				func(ctx *gin.Context) {
					//  Handler executed only if middleware passes
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)

			//  Prepare a new HTTP GET request to the `/auth` route
			recorder := httptest.NewRecorder()                             // Used to capture the response for verification
			request, err := http.NewRequest(http.MethodGet, authPath, nil) // No body needed
			require.NoError(t, err)                                        // Fail test if request creation failed


			// This could attach a Bearer token, unsupported scheme, or no header
			tc.setupAuth(t, request, server.tokenMaker)

			// → Request hits router
			// → Goes through middleware
			// → Middleware decides: forward or block
			server.router.ServeHTTP(recorder, request)

			//  Verify the result of middleware + handler execution
			// Test if middleware behaves as expected: status 200, 401 etc.
			tc.checkResponse(t, recorder)

		})
	}

}
