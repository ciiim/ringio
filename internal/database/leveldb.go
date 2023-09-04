package database

import (
	"github.com/syndtr/goleveldb/leveldb"
)

func NewLevelDB(path string) (*leveldb.DB, error) {
	return leveldb.OpenFile(path, nil)
}
