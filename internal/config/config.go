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

	// Vector search (MongoDB Atlas $vectorSearch on knowledge_chunks)
	VectorSearchIndex         string
	VectorSearchLimit         int
	VectorSearchNumCandidates int
	VectorSearchMinScore      float64
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

		VectorSearchIndex:         getEnv("VECTOR_SEARCH_INDEX", "knowledge_chunks_vector_index"),
		VectorSearchLimit:         getEnvInt("VECTOR_SEARCH_LIMIT", 5),
		VectorSearchNumCandidates: getEnvInt("VECTOR_SEARCH_NUM_CANDIDATES", 100),
		VectorSearchMinScore:      getEnvFloat("VECTOR_SEARCH_MIN_SCORE", 0.7),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
		log.Printf("⚠️ Invalid int for %s, using default %d\n", key, fallback)
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
		log.Printf("⚠️ Invalid float for %s, using default %v\n", key, fallback)
	}
	return fallback
}
