package core

func GET(key string) string {
	return database[key]
}

func SET(key, value string) {
	database[key] = value
}

func DEL(key string) {
	delete(database, key)
}
