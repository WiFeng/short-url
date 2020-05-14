package service

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis"
	"github.com/wifeng/short-url/pkg/dao"
)

const (
	shortDomain = "http://sh.url/"
)

var (
	base62    int64 = 62
	base62Map       = []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",

		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J",
		"K", "L", "M", "N", "O", "P", "Q", "R", "S", "T",
		"U", "V", "W", "X", "Y", "Z",

		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
		"k", "l", "m", "n", "o", "p", "q", "r", "s", "t",
		"u", "v", "w", "x", "y", "z",
	}
)

// Service describes a service that adds things together.
type Service interface {
	Create(ctx context.Context, longURL string) (string, error)
	Query(ctx context.Context, shortURL string) (string, error)
}

// New returns a basic Service with all of the expected middlewares wired in.
func New(db *sql.DB, re *redis.Client, logger log.Logger) Service {
	var svc Service
	{
		svc = NewBasicService(db, re, logger)
		svc = LoggingMiddleware(logger)(svc)
	}
	return svc
}

// NewBasicService returns a native, stateless implementation of Service.
func NewBasicService(db *sql.DB, re *redis.Client, logger log.Logger) Service {

	return &basicService{
		dao:    dao.New(db, re),
		logger: logger,
	}
}

type basicService struct {
	dao    *dao.Dao
	logger log.Logger
}

func (s *basicService) convertToBase62Str(id int64) string {

	var mod int64
	var base62Str string

	for id > 0 {
		mod = id % base62
		base62Str = base62Map[mod] + base62Str
		id = id / base62
	}

	return base62Str
}

func (s *basicService) Create(_ context.Context, longURL string) (string, error) {

	shortIDKey := fmt.Sprintf("%x", md5.Sum([]byte(longURL)))
	if shortURL, err := s.dao.GetShortURL(shortIDKey); err != nil {
		return "", err
	} else if shortURL != "" {
		return shortDomain + shortURL, nil
	}

	nextID, err := s.dao.GenerateID()
	if err != nil {
		return "", err
	}

	longIDKey := s.convertToBase62Str(nextID)
	if err := s.dao.SetLongURL(longIDKey, longURL); err != nil {
		return "", err
	}

	if err := s.dao.SetShortURL(shortIDKey, longIDKey); err != nil {
		return "", err
	}

	return shortDomain + longIDKey, nil
}

func (s *basicService) Query(_ context.Context, shortURL string) (string, error) {

	longIDKey := shortURL
	longURL, err := s.dao.GetLongURL(longIDKey)
	if err != nil {
		return "", err
	}
	return longURL, nil
}
