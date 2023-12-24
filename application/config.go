package application

import (
	"os"
	"strconv"
)

type Config struct {
	RedisAdress string
	ServerPort  uint16
}

func LoadConfig() Config {
	cfg := Config{
		RedisAdress: "localhost:6379",
		ServerPort:  3000,
	}

	if redisAddres, exist := os.LookupEnv("REDIS_ADDRESS"); exist {
		cfg.RedisAdress = redisAddres
	}

	if serverPort, exist := os.LookupEnv("SERVER_PORT"); exist {
		if port, err := strconv.ParseUint(serverPort, 10, 16); err == nil {
			cfg.ServerPort = uint16(port)
		}
	}

	return cfg
}
