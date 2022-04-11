package config

import (
	"log"
	"strconv"
)

var Default = &config{
	Database: Database{
		Driver:          "postgres",
		IP:              "localhost",
		Port:            getPort(),
		User:            "postgres",
		Password:        "postgres",
		Name:            "hero",
		ConnMaxIdle:     500,
		ConnMaxOpen:     300,
		ConnMaxLifetime: 50,
		Debug:           true,
		SSLMode:         "disable",
	},
	Server: Server{
		IP:   "127.0.0.1",
		Port: 5310,
	},
}

func getPort() int {
	sPort := "5432"
	port, err := strconv.ParseInt(sPort, 10, 32)
	if err != nil {
		log.Fatalln(err)
	}

	return int(port)
}
