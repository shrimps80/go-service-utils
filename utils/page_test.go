package utils

import (
	"encoding/json"
	"testing"
)

func TestNewPageData_JSON(t *testing.T) {
	p := NewPageData([]string{"a"}, 2, 10, int64(25))
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"list":["a"],"pageNum":2,"pageSize":10,"total":25}`
	if string(b) != want {
		t.Fatalf("got %s want %s", b, want)
	}
}
