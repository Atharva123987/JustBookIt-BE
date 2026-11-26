package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	MongoURI     string
	JWTSecret    string
	StripeSecret string

	VectorIndexName string
	VectorMinScore  float64
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

		VectorIndexName: getEnv("VECTOR_INDEX_NAME", "vector_index"),
		VectorMinScore:  getEnvFloat("VECTOR_MIN_SCORE", 0.58),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Printf("⚠️ Invalid float for %s=%q, using %v", key, value, fallback)
		return fallback
	}
	return parsed
}
