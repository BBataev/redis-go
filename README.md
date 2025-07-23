CodeCrafters tasks

Data Base structure

type DB struct {
	Value interface{}
	TTL   time.Time
}

var db = make(map[string]DB)

All commands handlers in internal/handler/commands.go