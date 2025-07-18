package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/itsadijmbt/simple_bank/db/sqlc"
	"github.com/itsadijmbt/simple_bank/db/util"
	"github.com/itsadijmbt/simple_bank/token"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// test
func TestMain(m *testing.M) {
	//* this is done as gin would provide a lot of debug statement that are not needed
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func NewTestServer(t *testing.T, store db.Store) *Server {

	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	require.NoError(t, err)
	server, err := NewServer(config, store)
	require.NoError(t, err)
	server.tokenMaker = tokenMaker
	server.setupRouter()
	return server

}
