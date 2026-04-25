package core

import "sync"

type Store struct {
	mutex sync.RWMutex
	m     map[string][]byte
}

var database Store = Store{
	m: make(map[string][]byte),
}

func GET(key []byte) []byte {
	database.mutex.RLock()
	value := database.m[string(key)]
	database.mutex.RUnlock()

	return value
}

func SET(key, value []byte) {
	database.mutex.Lock()
	database.m[string(key)] = value
	database.mutex.Unlock()
}

func DEL(key []byte) {
	database.mutex.Lock()
	delete(database.m, string(key))
	database.mutex.Unlock()
}
