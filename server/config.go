package server

const (
	DefPort 					= 80
	DefNumberOfServerGoroutines = 100
)

// submit server configuration
type Config struct {
	Port						int
	NumberOfServerGoroutines	int
}
