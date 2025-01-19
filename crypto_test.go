package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCopyEncryptDecrypt(t *testing.T) {
	payload := []byte("Foo not Bar")
	src := bytes.NewReader(payload)
	dst := new(bytes.Buffer)
	key := newEncryptionKey()
	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(len(payload))
	fmt.Println(len(dst.Bytes()))

	out := new(bytes.Buffer)
	nw, err := copyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}

	if nw != 16+len(payload) {
		t.Fail()
	}

	if out.String() != string(payload) {
		t.Errorf("expected %s, got %s", string(payload), out.String())
	}
}
