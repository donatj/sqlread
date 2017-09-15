package mapcache

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/donatj/sqlread"
)

type MapCache struct {
	sqlfile *os.File
}

func New(sqlfile *os.File) *MapCache {
	return &MapCache{
		sqlfile: sqlfile,
	}
}

var (
	ErrCacheMiss = errors.New("cache miss")
)

func (m *MapCache) Get() (sqlread.SummaryTree, error) {
	cf, err := os.Open(m.getCacheFile())
	defer cf.Close()
	if os.IsNotExist(err) {
		return nil, ErrCacheMiss
	} else if err != nil {
		return nil, err
	}

	d := json.NewDecoder(cf)

	v := cacheInfo{}
	err = d.Decode(&v)
	if err != nil {
		return nil, err
	}

	if v.FileSize != m.getFileSize() { // @todo more validation
		return nil, ErrCacheMiss
	}

	return v.Tree, nil
}

func (m *MapCache) Store(s sqlread.SummaryTree) error {
	cf, err := os.OpenFile(m.getCacheFile(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer cf.Close()
	if err != nil {
		return err
	}

	e := json.NewEncoder(cf)
	err = e.Encode(cacheInfo{
		FileSize: m.getFileSize(),
		Tree:     s,
	})

	if err != nil {
		return err
	}

	return nil
}

func (m *MapCache) getCacheFile() string {
	return m.sqlfile.Name() + ".sqlmap"
}

func (m *MapCache) getFileSize() int64 {
	f, err := m.sqlfile.Stat()
	if err != nil {
		return -1
	}

	return f.Size()
}

type cacheInfo struct {
	FileSize int64
	Tree     sqlread.SummaryTree
}
