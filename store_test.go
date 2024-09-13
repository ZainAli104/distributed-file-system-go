package main

import (
	"bytes"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	// TODO: Implement
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: DefaultPathTransformFunc,
	}
	s := NewStore(opts)

	data := bytes.NewReader([]byte("some jpg bytes"))
	err := s.writeStream("images", data)
	if err != nil {
		t.Error(err)
	}
}
