package config

import "time"

type AuthConfig struct {
	Secret     string
	Expiration time.Duration
}

type AppConfig struct {
	Auth     AuthConfig
	Database struct {
		URL            string
		MigrationsPath string
	}
	Server struct {
		HTTPAddress    string
		GRPCAddress    string
		AllowedOrigins []string
	}
	Log struct {
		Level string
	}
} 