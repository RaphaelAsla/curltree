package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	SSH      SSHConfig      `json:"ssh"`
	Database DatabaseConfig `json:"database"`
	Logging  LoggingConfig  `json:"logging"`
}

type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	RateLimit    RateLimitConfig `json:"rate_limit"`
}

type SSHConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	HostKeyPath string `json:"host_key_path"`
}

type DatabaseConfig struct {
	Type         string `json:"type"` // sqlite, postgres
	Path         string `json:"path"` // for sqlite
	Host         string `json:"host"` // for postgres
	Port         int    `json:"port"` // for postgres
	Name         string `json:"name"` // for postgres
	User         string `json:"user"` // for postgres
	Password     string `json:"password"` // for postgres
	SSLMode      string `json:"ssl_mode"` // for postgres
	MaxOpenConns int    `json:"max_open_conns"`
	MaxIdleConns int    `json:"max_idle_conns"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	Burst            int `json:"burst"`
}

type LoggingConfig struct {
	Level      string `json:"level"`      // debug, info, warn, error
	Format     string `json:"format"`     // json, text
	Output     string `json:"output"`     // stdout, stderr, file
	OutputFile string `json:"output_file"`
}

func Load() (*Config, error) {
	config := defaultConfig()
	
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		if err := loadFromFile(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	}
	
	loadFromEnv(config)
	
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return config, nil
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			RateLimit: RateLimitConfig{
				RequestsPerMinute: 60,
				Burst:            10,
			},
		},
		SSH: SSHConfig{
			Host:        "localhost",
			Port:        23234,
			HostKeyPath: ".ssh/curltree_host_key",
		},
		Database: DatabaseConfig{
			Type:         "sqlite",
			Path:         "./curltree.db",
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
	}
}

func loadFromFile(config *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return nil
}

func loadFromEnv(config *Config) {
	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
	
	if host := os.Getenv("SSH_HOST"); host != "" {
		config.SSH.Host = host
	}
	if port := os.Getenv("SSH_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.SSH.Port = p
		}
	}
	if hostKeyPath := os.Getenv("HOST_KEY_PATH"); hostKeyPath != "" {
		config.SSH.HostKeyPath = hostKeyPath
	}
	
	if dbType := os.Getenv("DB_TYPE"); dbType != "" {
		config.Database.Type = dbType
	}
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		config.Database.Path = dbPath
	}
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		if p, err := strconv.Atoi(dbPort); err == nil {
			config.Database.Port = p
		}
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		config.Database.Name = dbName
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		config.Database.User = dbUser
	}
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		config.Database.Password = dbPassword
	}
	if dbSSLMode := os.Getenv("DB_SSL_MODE"); dbSSLMode != "" {
		config.Database.SSLMode = dbSSLMode
	}
	
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.Logging.Level = logLevel
	}
	if logFormat := os.Getenv("LOG_FORMAT"); logFormat != "" {
		config.Logging.Format = logFormat
	}
	if logOutput := os.Getenv("LOG_OUTPUT"); logOutput != "" {
		config.Logging.Output = logOutput
	}
	if logOutputFile := os.Getenv("LOG_OUTPUT_FILE"); logOutputFile != "" {
		config.Logging.OutputFile = logOutputFile
	}
	
	if rateLimit := os.Getenv("RATE_LIMIT_PER_MINUTE"); rateLimit != "" {
		if r, err := strconv.Atoi(rateLimit); err == nil {
			config.Server.RateLimit.RequestsPerMinute = r
		}
	}
	if rateBurst := os.Getenv("RATE_LIMIT_BURST"); rateBurst != "" {
		if r, err := strconv.Atoi(rateBurst); err == nil {
			config.Server.RateLimit.Burst = r
		}
	}
}

func (c *Config) validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	
	if c.SSH.Port < 1 || c.SSH.Port > 65535 {
		return fmt.Errorf("invalid SSH port: %d", c.SSH.Port)
	}
	
	if c.Database.Type != "sqlite" && c.Database.Type != "postgres" {
		return fmt.Errorf("unsupported database type: %s", c.Database.Type)
	}
	
	if c.Database.Type == "sqlite" && c.Database.Path == "" {
		return fmt.Errorf("database path is required for SQLite")
	}
	
	if c.Database.Type == "postgres" {
		if c.Database.Host == "" {
			return fmt.Errorf("database host is required for PostgreSQL")
		}
		if c.Database.Name == "" {
			return fmt.Errorf("database name is required for PostgreSQL")
		}
		if c.Database.User == "" {
			return fmt.Errorf("database user is required for PostgreSQL")
		}
	}
	
	if c.Logging.Level != "debug" && c.Logging.Level != "info" && 
	   c.Logging.Level != "warn" && c.Logging.Level != "error" {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}
	
	if c.Logging.Format != "json" && c.Logging.Format != "text" {
		return fmt.Errorf("invalid log format: %s", c.Logging.Format)
	}
	
	if c.Logging.Output != "stdout" && c.Logging.Output != "stderr" && c.Logging.Output != "file" {
		return fmt.Errorf("invalid log output: %s", c.Logging.Output)
	}
	
	if c.Logging.Output == "file" && c.Logging.OutputFile == "" {
		return fmt.Errorf("log output file is required when output is 'file'")
	}
	
	return nil
}

func (c *Config) GetDatabaseURL() string {
	switch c.Database.Type {
	case "sqlite":
		return c.Database.Path + "?_foreign_keys=1"
	case "postgres":
		sslMode := c.Database.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Database.Host, c.Database.Port, c.Database.User, 
			c.Database.Password, c.Database.Name, sslMode)
	default:
		return ""
	}
}