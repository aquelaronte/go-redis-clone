package core

func GET(key []byte) []byte {
	return database[string(key)]
}

func SET(key, value []byte) {
	database[string(key)] = value
}

func DEL(key []byte) {
	delete(database, string(key))
}
