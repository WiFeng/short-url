package config

// Config config
type Config struct {
	Server  Server
	Redis   Redis
	Mysql   Mysql
	General General
}

// Server server config
type Server struct {
	Name string
	HTTP HTTP
	Log  zapLog
}

// HTTP http config
type HTTP struct {
	Addr string
}

// Redis redis config
type Redis struct {
	Host string
	Port int
	Auth string
	Db   int
}

// Mysql mysql config
type Mysql struct {
}
