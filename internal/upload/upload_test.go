package upload

import (
	"net/http"
	"net/http/httptest"
	"os"

	"testing"

	"github.com/stretchr/testify/assert"
)

// Test UploadDir, create a temp dir, zip it and upload it to a pseudo server with http recorder

func TestUploadDir(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

	}))
	defer svr.Close()

	// create a temp dir
	dir, err := os.MkdirTemp("", "tf-registry-test")
	if err != nil {
		t.Fatalf("failed to create temp dir, %v", err)
	}
	defer os.RemoveAll(dir)
	// call upload dir
	err = UploadDir(dir, svr.URL, "test", "test", "test", "test")

	assert.NoError(t, err)

}
