package utils

import "testing"

func TestMapPtr(t *testing.T) {
	type A struct{ N int }
	type B struct{ Twice int }

	items := []*A{{N: 1}, nil, {N: 3}}
	got := MapPtr(items, func(a *A) *B {
		return &B{Twice: a.N * 2}
	})
	if len(got) != 3 {
		t.Fatalf("len %d", len(got))
	}
	if got[0].Twice != 2 || got[1] != nil || got[2].Twice != 6 {
		t.Fatalf("got %#v", got)
	}
}

func TestMapPtr_empty(t *testing.T) {
	type A struct{}
	type B struct{}
	if got := MapPtr([]*A(nil), func(*A) *B { return nil }); got != nil {
		t.Fatalf("got %#v", got)
	}
}
