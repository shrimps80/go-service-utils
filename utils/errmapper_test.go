package utils

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

var errMapFoo = errors.New("foo")
var errMapBar = errors.New("bar")

func TestMapper_OnIs_firstMatch(t *testing.T) {
	var m Mapper
	codeFoo := &ErrorCode{Code: 9001, Message: "foo", Type: ErrorTypeBusiness}
	codeBar := &ErrorCode{Code: 9002, Message: "bar", Type: ErrorTypeBusiness}
	m.OnIs(errMapFoo, codeFoo).OnIs(errMapBar, codeBar)

	if got := m.Lookup(errors.Join(errMapFoo, errMapBar)); got != codeFoo {
		t.Fatalf("got %#v", got)
	}
}

func TestMapper_Default(t *testing.T) {
	m := NewMapper().Default(ErrBadRequest)
	if got := m.Lookup(errors.New("unknown")); got != ErrBadRequest {
		t.Fatalf("got %#v", got)
	}
}

func TestMapper_Write(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	m := NewMapper().OnIs(errMapFoo, ErrNotFound).Default(ErrInternalServer)
	m.Write(c, errMapFoo)
	if w.Code != http.StatusNotFound {
		t.Fatalf("http %d", w.Code)
	}
}

func TestMapper_Lookup_nilErr(t *testing.T) {
	var m Mapper
	if m.Lookup(nil) != nil {
		t.Fatal("expected nil")
	}
}
