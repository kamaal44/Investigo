package main_test

import (
	"plugin"
	"testing"
)

func TestSearch(t *testing.T) {
	p, err := plugin.Open("twitter.so")
	if err != nil {
		panic(err)
	}

	f, err := p.Lookup("Search")
	if err != nil {
		panic(err)
	}

	search := f.(func(string, string) []byte)
	results := search("Unfriended", "qwerty")
	t.Log(string(results))
}
