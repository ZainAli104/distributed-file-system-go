package main

import (
	"bytes"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "momsbestpicture"
	pathKey := CASPathTransformFunc(key)
	expectedFilename := "6804429f74181a63c50c3d81d733a12f14a353ff"
	expectedPathName := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
	if pathKey.PathName != expectedPathName {
		t.Errorf("Expected %s, got %s", expectedPathName, pathKey.PathName)
	}
	if pathKey.Filename != expectedFilename {
		t.Errorf("Expected %s, got %s", expectedFilename, pathKey.Filename)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)

	key := "myspecials"
	data := []byte("some jpg bytes")

	err := s.writeStream(key, bytes.NewReader(data))
	if err != nil {
		t.Error(err)
	}

	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, _ := io.ReadAll(r)

	if !bytes.Equal(b, data) {
		t.Errorf("Expected %s, got %s", data, b)
	}
}

func TestStoreDeleteKey(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)

	key := "myspecialspictures"
	data := []byte("some jpg bytes")

	err := s.writeStream(key, bytes.NewReader(data))
	if err != nil {
		t.Error(err)
	}

	err = s.Delete(key)
	if err != nil {
		t.Error(err)
	}
}
