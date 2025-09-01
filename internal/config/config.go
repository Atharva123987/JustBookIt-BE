package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	MongoURI     string
	JWTSecret    string
	StripeSecret string
}

// Global variable (exported so other packages can use it)
var AppConfig *Config

func InitConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, using system environment")
	}

	AppConfig = &Config{
		Port:         getEnv("PORT", "3000"),
		MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
		JWTSecret:    getEnv("JWT_SECRET", "defaultsecret"),
		StripeSecret: getEnv("STRIPE_SECRET", "your-stripe-secret-key-here"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
