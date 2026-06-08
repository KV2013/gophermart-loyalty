package config

import "github.com/caarlos0/env/v6"

type Config struct {
	ServerAddress        string `env:"RUN_ADDRESS"`
	BaseURL              string `env:"BASE_URL"`
	LogLevel             string `env:"LOG_LEVEL"`
	DatabaseURI          string `env:"DATABASE_URI"`
	JWTSecretKey         string `env:"JWT_SECRET_KEY"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewConfig() (*Config, error) {

	cfg := Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		LogLevel:      "info",
		JWTSecretKey:  "default-secret-change-me",
	}

	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

	parseFlags()

	if flagServerAddr != "" {
		cfg.ServerAddress = flagServerAddr
	}
	if flagBaseURL != "" {
		cfg.BaseURL = flagBaseURL
	}
	if flagLogLevel != "" {
		cfg.LogLevel = flagLogLevel
	}
	if flagDatabaseURI != "" {
		cfg.DatabaseURI = flagDatabaseURI
	}
	if flagJWTSecretKey != "" {
		cfg.JWTSecretKey = flagJWTSecretKey
	}
	if flagAccrualSystemAddress != "" {
		cfg.AccrualSystemAddress = flagAccrualSystemAddress
	}

	return &cfg, nil
}
