package cache

import (
	"bytes"
	"encoding/gob"
)

var c Cache

type Cache struct {
	c map[string][]byte
}

func New() (c *Cache) {
	return &Cache{}
}

func (c *Cache) Set(key string, data interface{}) (err error) {
	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(data)
	if err != nil {
		return
	}
	(*c).c[key] = buf.Bytes()
	return
}

func (c Cache) Get(key string, data interface{}) {
	gob.NewDecoder(bytes.NewReader(c.c[key])).Decode(data)
}
