package base

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"shortener/internal/storage"
	"shortener/internal/utils"
)

type Storage struct {
	uris      map[string]string
	urisMutex sync.RWMutex

	users      map[string]map[string]struct{}
	usersMutex sync.RWMutex

	File   *os.File
	writer *bufio.Writer
}

type option func(s *Storage) error

func WithFile(filename string) option {
	return func(s *Storage) error {
		file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
		if err != nil {
			return fmt.Errorf("cannot open or create File %s (%s), continue without holding", filename, err.Error())
		}

		s.urisMutex.Lock()
		defer func() {
			s.urisMutex.Unlock()
		}()

		scanner := bufio.NewScanner(file)
		n := &storage.Users{}

		for scanner.Scan() {
			if err := json.Unmarshal(scanner.Bytes(), n); err != nil {
				return fmt.Errorf("cannot unmarshar stored data from File %s (%s), continue without holding", filename, err.Error())
			}

			s.uris[n.Short] = n.URI
		}

		s.File = file
		s.writer = bufio.NewWriter(s.File)
		return nil
	}
}

func New(opts ...option) *Storage {
	st := &Storage{
		uris:      map[string]string{},
		urisMutex: sync.RWMutex{},

		users:      map[string]map[string]struct{}{},
		usersMutex: sync.RWMutex{},
	}

	for _, opt := range opts {
		if err := opt(st); err != nil {
			log.Printf("warn>: %s", err.Error())
		}
	}

	if st.File != nil {
		log.Println(st.File.Name())
	}

	return st
}

func (st *Storage) store(uri, short string) error {
	n := &storage.Users{
		URI:   uri,
		Short: short,
	}

	data, err := json.Marshal(n)
	if err != nil {
		return err
	}

	if _, err := st.writer.Write(append(data, '\n')); err != nil {
		return err
	}

	if err := st.writer.Flush(); err != nil {
		return err
	}

	return nil
}

func (st *Storage) Ping(ctx context.Context) error {
	return nil
}

func (st *Storage) Get(short string) (string, bool) {
	st.urisMutex.RLock()
	fmt.Println(st.uris)
	defer st.urisMutex.RUnlock()

	uri, ok := st.uris[short]
	return uri, ok
}

func (st *Storage) Users(base, hash string) ([]storage.Users, error) {
	st.usersMutex.RLock()
	defer st.usersMutex.RUnlock()

	u := make([]storage.Users, 0)
	if shortMap, ok := st.users[hash]; ok {
		for short := range shortMap {
			if uri, ok := st.uris[short]; ok {
				u = append(u, storage.Users{
					URI:   uri,
					Short: fmt.Sprintf("%s/%s", base, short),
				})
			}
		}
	}

	return u, nil
}

func (st *Storage) Push(uri, hash string) (string, error) {
	short, err := utils.Shorty(st, uri) // shorty(Storage)
	if err != nil {
		return short, err
	}

	st.urisMutex.Lock()
	defer st.urisMutex.Unlock()
	if _, ok := st.uris[short]; !ok && st.File != nil {
		if err := st.store(uri, short); err != nil {
			return "", err
		}
	}
	st.uris[short] = uri

	st.usersMutex.Lock()
	defer st.usersMutex.Unlock()
	if _, ok := st.users[hash]; !ok {
		st.users[hash] = map[string]struct{}{}
	}
	st.users[hash][short] = struct{}{}

	return short, nil
}

func (st *Storage) Close() error {
	if st.File != nil {
		return st.File.Close()
	}
	return nil
}
