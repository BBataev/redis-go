package entity

import "time"

type DB struct {
	Value interface{}
	TTL   time.Time
}

var Db = make(map[string]DB)
