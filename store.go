package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRoot = "sdnetwork"

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blocksize := 5
	slicelen := len(hashStr) / blocksize

	paths := make([]string, slicelen)

	for i := 0; i < slicelen; i++ {
		from, to := i*blocksize, (i*blocksize)+blocksize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		Pathname: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

type PathKey struct {
	Pathname string
	Filename string
}

type PathTransformFunc func(string) PathKey

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		Pathname: key,
		Filename: key,
	}
}

type StoreOpts struct {
	// Root storage name containing all the folders/files of the system
	Root              string
	PathTransformFunc PathTransformFunc
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if len(opts.Root) == 0 {
		opts.Root = defaultRoot
	}
	return &Store{
		StoreOpts: opts,
	}
}

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.Pathname, p.Filename)
}

func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.Pathname, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func (s *Store) Has(key string) bool {
	pathkey := s.PathTransformFunc(key)

	_, err := os.Stat(pathkey.FullPath())

	if err != nil {
		return false
	}
	return true
}

func (s *Store) Delete(key string) error {
	pathkey := s.PathTransformFunc(key)

	// if err := os.RemoveAll(pathkey.FullPath()); err != nil {
	// 	return err
	// }
	return os.RemoveAll(pathkey.FirstPathName())
}

func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	buf := new(bytes.Buffer)
	io.Copy(buf, f)

	return buf, nil
}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathkey := s.PathTransformFunc(key)

	return os.Open(pathkey.FullPath())
}

func (s *Store) writestream(key string, r io.Reader) error {
	pathkey := s.PathTransformFunc(key)

	if err := os.MkdirAll(pathkey.Pathname, os.ModePerm); err != nil {
		return err
	}

	path := pathkey.FullPath()

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("written (%d) bytes to disk: %s", n, path)
	return nil
}
