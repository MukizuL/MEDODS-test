package config

import (
	"flag"
	"log"
	"log/slog"
	"os"
)

type Params struct {
	Addr       string
	PrivateKey string
	DSN        string
}

func GetParams() *Params {
	result := &Params{}

	result.Addr = getEnv("ADDR", "")
	result.PrivateKey = getEnv("PK", "")
	result.DSN = getEnv("DSN", "")

	if result.Addr == "" {
		flag.StringVar(&result.Addr, "a", "localhost:8080", "Sets server address. Default: localhost:8080")
	}
	if result.PrivateKey == "" {
		flag.StringVar(&result.PrivateKey, "pk", "", "Private key is mandatory")
	}
	if result.DSN == "" {
		flag.StringVar(&result.DSN, "d", "", "DSN Format: postgres://user:password@address:port/database")
	}

	flag.Parse()

	if result.PrivateKey == "" {
		flag.Usage()
		log.Fatal("Private key is not specified")
	}

	if result.DSN == "" {
		flag.Usage()
		log.Fatal("DSN is not specified")
	}

	return result
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Key is empty", "key=", key)
	return fallback
}
