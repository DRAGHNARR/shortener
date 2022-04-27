package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"shortener/internal/utils"
)

type option func(s *Storage) error

type Storage struct {
	Box      map[string]string
	UserBox  map[string]map[string]struct{}
	BoxMutex sync.RWMutex
	File     *os.File
	writer   *bufio.Writer
}

type ByUser struct {
	Orig   string `json:"original_url"`
	Shorty string `json:"short_url"`
}

func New(opts ...option) *Storage {
	s := &Storage{
		Box:      map[string]string{},
		UserBox:  map[string]map[string]struct{}{},
		BoxMutex: sync.RWMutex{},
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			log.Printf("warn>: %s", err.Error())
		}
	}

	if s.File != nil {
		log.Println(s.File.Name())
	}

	return s
}

func WithFile(filename string) option {
	return func(s *Storage) error {
		file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
		if err != nil {
			return fmt.Errorf("cannot open or create File %s (%s), continue without holding", filename, err.Error())
		}

		s.BoxMutex.Lock()
		defer func() {
			s.BoxMutex.Unlock()
		}()

		scanner := bufio.NewScanner(file)
		n := &utils.Node{}

		for scanner.Scan() {
			if err := json.Unmarshal(scanner.Bytes(), n); err != nil {
				return fmt.Errorf("cannot unmarshar stored data from File %s (%s), continue without holding", filename, err.Error())
			}

			s.Box[n.Shorty] = n.Orig
		}

		s.File = file
		s.writer = bufio.NewWriter(s.File)
		return nil
	}
}

func (s *Storage) store(orig, shorty string) error {
	n := &utils.Node{
		Orig:   orig,
		Shorty: shorty,
	}

	data, err := json.Marshal(n)
	if err != nil {
		return err
	}

	if _, err := s.writer.Write(append(data, '\n')); err != nil {
		return err
	}

	if err := s.writer.Flush(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Append(orig, ID string) (string, error) {
	s.BoxMutex.Lock()
	defer func() {
		s.BoxMutex.Unlock()
	}()

	shorty, added, err := utils.Shotifier(&s.Box, orig)
	if err != nil {
		return "", err
	}

	if _, ok := s.UserBox[ID]; !ok {
		s.UserBox[ID] = map[string]struct{}{}
	}
	s.UserBox[ID][shorty] = struct{}{}

	if added && s.File != nil {
		if err := s.store(orig, shorty); err != nil {
			return "", err
		}
	}
	return shorty, nil
}

func (s *Storage) Get(shorty string) (string, bool) {
	s.BoxMutex.RLock()
	defer func() {
		s.BoxMutex.RUnlock()
	}()

	orig, ok := s.Box[shorty]
	return orig, ok
}

func (s *Storage) GetByUser(ID string) []ByUser {
	s.BoxMutex.RLock()
	defer func() {
		s.BoxMutex.RUnlock()
	}()

	byUser := make([]ByUser, 0)
	if shortyMap, ok := s.UserBox[ID]; ok {
		fmt.Println(shortyMap)
		for shorty, _ := range shortyMap {
			if orig, ok := s.Box[shorty]; ok {
				byUser = append(byUser, ByUser{
					Orig:   orig,
					Shorty: fmt.Sprintf("http://localhost:8080/%s", shorty),
				})
			}
		}
	}

	return byUser
}

func (s *Storage) Close() error {
	return s.File.Close()
}
