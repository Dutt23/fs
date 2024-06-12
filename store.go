package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRoot = "sdnetwork"

func CASPathTransformFunc(root, key string) PathKey {
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
		Pathname: fmt.Sprintf("%s/%s", root, strings.Join(paths, "/")),
		Filename: hashStr,
	}
}

type PathKey struct {
	Pathname string
	Filename string
}

type PathTransformFunc func(string, string) PathKey

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
	pathkey := s.PathTransformFunc(s.Root, key)

	_, err := os.Stat(pathkey.FullPath())

	return err == nil
}

func (s *Store) Delete(key string) error {
	pathkey := s.PathTransformFunc(s.Root, key)

	// if err := os.RemoveAll(pathkey.FullPath()); err != nil {
	// 	return err
	// }
	return os.RemoveAll(pathkey.FirstPathName())
}

func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writestream(key, r)
}

// TODO : Instead of reading into memory
func (s *Store) Read(key string) (int64, io.Reader, error) {
	return s.readStream(key)
}

func (s *Store) readStream(key string) (int64, io.ReadCloser, error) {
	pathkey := s.PathTransformFunc(s.Root, key)
	f, err := os.Open(pathkey.FullPath())
	if err != nil {
		return 0, nil, err
	}

	stat, err := os.Stat(pathkey.FullPath())
	if err != nil {
		return 0, nil, err
	}
	return stat.Size(), f, nil
}

func (s *Store) WriteDecrypt(encKey []byte, key string, src io.Reader) (int64, error) {
	pathkey := s.PathTransformFunc(s.Root, key)

	if err := os.MkdirAll(pathkey.Pathname, os.ModePerm); err != nil {
		return 0, err
	}

	path := pathkey.FullPath()

	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}

	// Waits for end of file and blocks
	n, err := copyDecrypt(encKey, src, f)
	if err != nil {
		return 0, err
	}

	log.Printf("written (%d) bytes to disk: %s", n, path)
	return int64(n), err
}

func (s *Store) openFileForWriting(key string) (*os.File, error) {
	pathkey := s.PathTransformFunc(s.Root, key)

	if err := os.MkdirAll(pathkey.Pathname, os.ModePerm); err != nil {
		return nil, err
	}

	path := pathkey.FullPath()

	return os.Create(path)
}

func (s *Store) writestream(key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(key)

	if err != nil {
		return 0, err
	}
	// Waits for end of file and blocks
	return io.Copy(f, r)
}
