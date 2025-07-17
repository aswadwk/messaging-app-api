package config

import (
	// "log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

type Config struct {
	AppName  string
	AppEnv   string
	AppPort  string
	Debug    bool
	LogLevel string
	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUserName string
	DBPassword string

	RabbitMQURL string

	JWTSecret          string
	JWTAccessTokenTTL  string
	JWTRefreshTokenTTL string
}

var Cfg Config

func LoadConfig() {
	_ = godotenv.Load() // silently load .env (ignore error if missing)

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	Cfg = Config{
		AppName:    getEnv("APP_NAME", "App Server"),
		AppEnv:     getEnv("APP_ENV", env),
		AppPort:    getEnv("APP_PORT", "8080"),
		Debug:      getEnv("DEBUG", "false") == "true",
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_DATABASE", "messaging_task_db"),
		DBUserName: getEnv("DB_USERNAME", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),

		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),

		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key"), // Default secret key, sebaiknya diganti di production
		JWTAccessTokenTTL:  getEnv("JWT_ACCESS_TOKEN_TTL", "1h"),
		JWTRefreshTokenTTL: getEnv("JWT_REFRESH_TOKEN_TTL", "24h"),
	}
}

func getEnv(key string, fallback string) string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func NewFiberApp() *fiber.App {
	app := fiber.New()
	return app
}
