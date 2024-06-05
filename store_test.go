package main

import (
	"bytes"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "momsbestpicture"
	pathkey := CASPathTransformFunc("root", key)
	expectedPathname := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
	if pathkey.Pathname != expectedPathname {
		t.Error(t, "have %s want %s", pathkey.Pathname, expectedPathname)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)
	key := "testingstore"
	data := []byte("some jpeg byte")
	if _, err := s.writestream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	_, r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, _ := io.ReadAll(r)
	if string(b) != string(data) {
		t.Errorf("want %s have %s", data, b)
	}

	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}
