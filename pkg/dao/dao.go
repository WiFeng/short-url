package dao

import (
	"database/sql"

	"github.com/go-redis/redis"
)

var db *sql.DB
var re *redis.Client

// Dao struct
type Dao struct {
}

// New Dao
func New(d *sql.DB, r *redis.Client) *Dao {
	db = d
	re = r
	return &Dao{}
}

// GenerateID ...
func (dao *Dao) GenerateID() (int64, error) {
	return _r.generateID()
}

// GetLongURL ...
func (dao *Dao) GetLongURL(idKey string) (string, error) {
	return _r.getLongURL(idKey)
}

// GetShortURL ...
func (dao *Dao) GetShortURL(idKey string) (string, error) {
	return _r.getShortURL(idKey)
}

// SetLongURL ...
func (dao *Dao) SetLongURL(idKey string, val string) error {
	return _r.setLongURL(idKey, val)
}

// SetShortURL ...
func (dao *Dao) SetShortURL(idKey string, val string) error {
	return _r.setShortURL(idKey, val)
}
