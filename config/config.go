package config

type config struct {
	Database Database
	Server   Server
}

type Database struct {
	Driver          string
	IP              string
	Port            int
	User            string
	Password        string `json13:"-"`
	Name            string
	ConnMaxIdle     int
	ConnMaxOpen     int
	ConnMaxLifetime int
	Debug           bool
	SSLMode         string
}

type Server struct {
	IP   string
	Port int
}
