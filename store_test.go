package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "momsbestpicture"
	pathname := CASPathTransformFunc(key)
	fmt.Println("Pathname: ", pathname)
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
