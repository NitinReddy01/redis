package core

import "time"

var store map[string]*Obj

type Obj struct {
	value     interface{}
	expiresAt int64
}

func init() {
	store = make(map[string]*Obj)
}

func NewObj(value interface{}, durationMs int64) *Obj {
	var expiresAt int64 = -1
	if durationMs > 0 {
		expiresAt = time.Now().UnixMilli() + durationMs
	}

	return &Obj{
		value:     value,
		expiresAt: expiresAt,
	}
}

func Put(k string, obj *Obj) {
	store[k] = obj
}

func Get(k string) *Obj {
	return store[k]
}
