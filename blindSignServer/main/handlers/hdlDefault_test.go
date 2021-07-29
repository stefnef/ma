package handlers

import (
	"blindSignAccount/main/model"
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var r *gin.Engine

func callURL(method, url string, expStatus int, body io.Reader, t *testing.T) *bytes.Buffer {
	// request for all flights
	req, _ := http.NewRequest(method, url, body)
	req.Header.Add("Accept", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != expStatus {
		t.Error("Bad status: " + strconv.Itoa(w.Code))
		t.Fail() // build the list of all flights
	}
	return w.Body
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	r = gin.Default()
	InitRoutes(r)
	os.Exit(m.Run())
}

func setup(t *testing.T) {
	Server = model.NewServer()
	if Server == nil {
		t.Error("Server is nil")
		t.Fail()
	}
}
