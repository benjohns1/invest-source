package app

// App provides methods for all application use-cases.
type App struct {
	Cache Cache
	Provider Provider
	Log Log
}

// Cache caches API data when multiple use-cases are run for the same dataset without having to re-query the source API.
type Cache interface {
	ReadCurrent() ([]byte, error)
	WriteCurrent(data []byte) error
}

// Provider implements a source provider for retrieving external data.
type Provider interface {
	QueryLatest() ([]byte, error)
}

// Log interface.
type Log interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}
