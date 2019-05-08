package function

import (
	"net/http"
)

// Handle a serverless request
func Handle(req http.Request) ([]byte, error) {
	return []byte("hi"), nil
}
