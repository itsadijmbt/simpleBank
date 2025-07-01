package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// test
func TestMain(m *testing.M) {
	//* this is done as gin would provide a lot of debug statement that are not needed
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
