package core

import "time"

var database store = store{
	data: make(map[string]*entry),
}

func GET(key []byte) []byte {
	// Check the key exists
	database.mutex.RLock()
	value, ok := database.data[string(key)]

	if !ok || value == nil {
		database.mutex.RUnlock()
		return nil
	}

	// If there's no expiration or the entry is
	// not expired, then return the value
	if value.expiresAt == 0 || time.Now().Unix() <= value.expiresAt {
		database.mutex.RUnlock()
		return value.bytes
	}

	// Value is expired at this point
	database.mutex.RUnlock()

	database.mutex.Lock()

	// Doble check to look whether another thread has deleted or updated the value
	value, ok = database.data[string(key)]

	if ok && value.expiresAt != 0 && time.Now().Unix() > value.expiresAt {
		delete(database.data, string(key))
	}

	database.mutex.Unlock()
	return nil
}

func SET(key, value []byte) {
	database.mutex.Lock()
	database.data[string(key)] = &entry{
		bytes: value,
	}
	database.mutex.Unlock()
}

func EXPIRE(key []byte, seconds int64) bool {
	// Check the key exists
	database.mutex.RLock()
	value, ok := database.data[string(key)]

	if !ok || value == nil {
		database.mutex.RUnlock()
		return false
	}

	// The key exists, then set expiration
	database.mutex.RUnlock()

	database.mutex.Lock()
	value.expiresAt = time.Now().Unix() + seconds

	database.mutex.Unlock()
	return true
}

func DEL(key []byte) bool {
	database.mutex.RLock()
	value, ok := database.data[string(key)]

	if !ok || value == nil {
		database.mutex.RUnlock()
		return false
	}

	database.mutex.RUnlock()
	database.mutex.Lock()
	delete(database.data, string(key))

	database.mutex.Unlock()

	return true
}
