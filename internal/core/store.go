package core

import "sync"

type store struct {
	mutex sync.RWMutex
	data  map[string]*entry
}
