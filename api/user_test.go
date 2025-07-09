package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	mockdb "github.com/itsadijmbt/simple_bank/db/mock"
	db "github.com/itsadijmbt/simple_bank/db/sqlc"
	"github.com/itsadijmbt/simple_bank/db/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

/*
* Matcher is a interface with |MATCHES(x interface{})bool|  and  |STRING()string| method
* go to matcher implemetnatoin and pick the desired one here we used
type eqMatcher struct {
	x any
}
^as the name suggestes "eqCreateUserParamsMatcher" matches the params of the func
*/

// eqCreateUserParamsMatcher is a custom gomock Matcher that compares
// the CreateUserParams passed into our mock store to an expected value,
// "while" also verifying that the password was hashed correctly.
type eqCreateUserParamsMatcher struct {
	// arg holds the expected CreateUserParams (username, email, etc.)
	// including the expected hashed password once we set it.
	arg db.CreateUserParams

	// password is the plain-text password we originally passed into
	// our CreateUser call. We need this here so we can re-hash
	// and compare inside Matches().
	password string
}

// Matches is invoked by gomock to check whether the actual argument
// x (what the handler passed into store.CreateUser) satisfies our expectations.
func (e eqCreateUserParamsMatcher) Matches(x any) bool {
	// 1. Assert that x is of the correct type
	//    If it's not a CreateUserParams, it definitely doesn’t match.
	//! argeument has already been exracted rom x.CreateUserParams

	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	//^ 2. Verify the password hashing:
	//    util.CheckPassword takes our plain-text password (e.password)
	//    and compares it against the hashed password from the handler (arg.HashedPassword).
	//    If they don’t match—or hashing failed—we bail out.
	if err := util.CheckPassword(e.password, arg.HashedPassword); err != nil {
		return false
	}

	// 3. Now that we know arg.HashedPassword is a valid hash of e.password,
	//    update our expected arg so the two structs only differ in that one field.
	//    This allows us to do a deep-equality on everything else.
	// !e's method arg.hashp is assigned arg.hashp
	// why this is done as everytime has generated is diff

	//^eariler we had only check if the newhash belongs to the naked password
	//^ and now we assign the newhas=oldhas
	e.arg.HashedPassword = arg.HashedPassword

	// 4. Finally, do a deep-equal check of the entire struct:
	//    reflect.DeepEqual will compare all fields (username, email, full name, AND our now-updated HashedPassword).
	return reflect.DeepEqual(e.arg, arg)
}

// String is used by gomock to describe this matcher in error messages.
// It should say what we expected, e.g. “matches CreateUserParams with Username=...”
func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches CreateUserParams {Username:%q, FullName:%q, Email:%q, HashedPassword:<hashed of %q>}",
		e.arg.Username,
		e.arg.FullName,
		e.arg.Email,
		e.password,
	)
}

// * creates an instance for each test
func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {

	return eqCreateUserParamsMatcher{arg, password}
}

// ------------------------------------------------------------------
// Usage in your test setup:
//
//   // Given a plain-text password, and an expected db.CreateUserParams
//   matcher := eqCreateUserParamsMatcher{
//       arg: db.CreateUserParams{
//           Username: user.Username,
//           FullName: user.FullName,
//           Email:    user.Email,
//           // HashedPassword will be filled in by Matches()
//       },
//       password: password,
//   }
//
//   store.EXPECT().
//       CreateUser(gomock.Any(), matcher).
//       Times(1).
//       Return(user, nil)
//
// Now when your handler calls CreateUser(ctx, someParams), gomock will
// invoke matcher.Matches(someParams). Internally it will:
//    1. Confirm someParams is a CreateUserParams.
//    2. Re-hash & compare the password to ensure the handler hashed correctly.
//    3. Update matcher.arg.HashedPassword so that DeepEqual passes on all other fields.
//    4. Return true/false accordingly.
//

func TestCreateUserApi(t *testing.T) {

	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				//! CreateUser(gomock.Any(), gomock.eq(arg)). it fails as when it
				//! passes args and we recieve them the eq matches the hash pass that will
				//! be diff beacause of SALT so even if we pass the same pass it will produce
				//! diff hask
				//^ WE HAVE TO USE  A CUSTOM MATCHER

				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		//TODO:
		// {

		// 	name: "InternalError",
		// 	body: gin.H{
		// 		"username":  user.Username,
		// 		"password":  password,
		// 		"full_name": user.FullName,
		// 		"email":     user.Email,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		arg := db.CreateUserParams{
		// 			Username: user.Username,
		// 			FullName: user.FullName,
		// 			Email:    user.Email,
		// 		}
		// 		store.EXPECT().CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).Times(1).
		// 			Return(db.User{}, sql.ErrConnDone)
		// 	},
		// 	checkResponse: func(recoder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusInternalServerError, recoder.Code)
		// 	},
		// },
		// {
		// 	name: "DuplicateUserName",
		// 	body: gin.H{
		// 		"username":  user.Username,
		// 		"password":  password,
		// 		"full_name": user.FullName,
		// 		"email":     user.Email,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		arg := db.CreateUserParams{
		// 			Username: user.Username,
		// 			FullName: user.FullName,
		// 			Email:    user.Email,
		// 		}
		// 		store.EXPECT().CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).Times(1).
		// 			Return(db.User{}, sql.ErrConnDone)
		// 	},
		// 	checkResponse: func(recoder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusInternalServerError, recoder.Code)
		// 	},
		// },
		// {
		// 	name: "InvalidUsername",
		// 	body: gin.H{
		// 		"username":  "invalid-user#1",
		// 		"password":  password,
		// 		"full_name": user.FullName,
		// 		"email":     user.Email,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			CreateUser(gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusBadRequest, recorder.Code)
		// 	},
		// },
		// {
		// 	name: "InvalidEmail",
		// 	body: gin.H{
		// 		"username":  user.Username,
		// 		"password":  password,
		// 		"full_name": user.FullName,
		// 		"email":     "invalid-email",
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			CreateUser(gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusBadRequest, recorder.Code)
		// 	},
		// },
		// {
		// 	name: "TooShortPassword",
		// 	body: gin.H{
		// 		"username":  user.Username,
		// 		"password":  "123",
		// 		"full_name": user.FullName,
		// 		"email":     user.Email,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			CreateUser(gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusBadRequest, recorder.Code)
		// 	},
		// },
	}

	for i := range testCases {
		// Capture tc to avoid the “range variable reused” gotcha
		tc := testCases[i]

		//* Run subtest named by tc.name
		t.Run(tc.name, func(t *testing.T) {
			//* 1. Create a new gomock Controller for this subtest
			ctrl := gomock.NewController(t)
			//* Ensure all expected calls were made, or fail the test
			defer ctrl.Finish()

			//* 2. Create a mock store and set up its expected behavior
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			//* 3. Initialize your server with the mocked store
			server := NewTestServer(t, store)

			//* 4. Recorder to capture HTTP responses
			recorder := httptest.NewRecorder()

			//* 5. Marshal the test case body into JSON payload
			data, err := json.Marshal(tc.body)
			require.NoError(t, err) // should never fail in test setup

			//^ ->http.NewRequest (and bytes.NewBuffer) to simulate the incoming HTTP POST in your tests,
			//^ ->ctx.JSON to marshal+write the outgoing response inside your actual handler.

			//* 6. Build the HTTP request targeting your user‐creation endpoint
			url := "/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
			require.NoError(t, err)

			//* 7. Dispatch the request through Gin’s router
			server.router.ServeHTTP(recorder, request)

			//! 8. Execute the test case’s custom response checks
			tc.checkResponse(recorder)
		})
	}

}

// func TestLoginUserAPI(t *testing.T) {
// 	user, password := randomUser(t)

// 	testCases := []struct {
// 		name          string
// 		body          gin.H
// 		buildStubs    func(store *mockdb.MockStore)
// 		checkResponse func(recoder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name: "OK",
// 			body: gin.H{
// 				"username": user.Username,
// 				"password": password,
// 			},
// 			buildStubs: func(store *mockdb.MockStore) {
// 				store.EXPECT().
// 					GetUser(gomock.Any(), gomock.Eq(user.Username)).
// 					Times(1).
// 					Return(user, nil)
// 				store.EXPECT().
// 					CreateSession(gomock.Any(), gomock.Any()).
// 					Times(1)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				require.Equal(t, http.StatusOK, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UserNotFound",
// 			body: gin.H{
// 				"username": "NotFound",
// 				"password": password,
// 			},
// 			buildStubs: func(store *mockdb.MockStore) {
// 				store.EXPECT().
// 					GetUser(gomock.Any(), gomock.Any()).
// 					Times(1).
// 					Return(db.User{}, db.ErrRecordNotFound)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				require.Equal(t, http.StatusNotFound, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "IncorrectPassword",
// 			body: gin.H{
// 				"username": user.Username,
// 				"password": "incorrect",
// 			},
// 			buildStubs: func(store *mockdb.MockStore) {
// 				store.EXPECT().
// 					GetUser(gomock.Any(), gomock.Eq(user.Username)).
// 					Times(1).
// 					Return(user, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				require.Equal(t, http.StatusUnauthorized, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "InternalError",
// 			body: gin.H{
// 				"username": user.Username,
// 				"password": password,
// 			},
// 			buildStubs: func(store *mockdb.MockStore) {
// 				store.EXPECT().
// 					GetUser(gomock.Any(), gomock.Any()).
// 					Times(1).
// 					Return(db.User{}, sql.ErrConnDone)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				require.Equal(t, http.StatusInternalServerError, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "InvalidUsername",
// 			body: gin.H{
// 				"username": "invalid-user#1",
// 				"password": password,
// 			},
// 			buildStubs: func(store *mockdb.MockStore) {
// 				store.EXPECT().
// 					GetUser(gomock.Any(), gomock.Any()).
// 					Times(0)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				require.Equal(t, http.StatusBadRequest, recorder.Code)
// 			},
// 		},
// 	}

// 	for i := range testCases {
// 		tc := testCases[i]

// 		t.Run(tc.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			store := mockdb.NewMockStore(ctrl)
// 			tc.buildStubs(store)

// 			server := NewServer(store)
// 			recorder := httptest.NewRecorder()

// 			// Marshal body data to JSON
// 			data, err := json.Marshal(tc.body)
// 			require.NoError(t, err)

// 			url := "/users/login"
// 			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
// 			require.NoError(t, err)

// 			server.router.ServeHTTP(recorder, request)
// 			tc.checkResponse(recorder)
// 		})
// 	}
// }

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(10)
	hashedPassword, err := util.HashedPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	// !t.FailNow() (or t.Fatal()), which immediately stops that test goroutine
	data, err := io.ReadAll(body)

	require.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)
	//! the user response json has no "hashedpassowrd feild";
	// require.NotEmpty(t, gotUser.HashedPassword)
	require.NoError(t, err)

	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.Equal(t, user.Username, gotUser.Username)

}
