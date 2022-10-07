package db

type UnusualDatabase struct {
	data map[string]string
}

func NewUnusualDatabase() *UnusualDatabase {
	data := make(map[string]string)
	data["version"] = "Ken's Key-Value Store 1.0"

	return &UnusualDatabase{
		data: data,
	}
}

func (db *UnusualDatabase) Set(key, value string) {
	if key == "version" {
		return
	}

	db.data[key] = value
}

func (db *UnusualDatabase) Get(key string) (string, bool) {
	value, ok := db.data[key]
	return value, ok
}
