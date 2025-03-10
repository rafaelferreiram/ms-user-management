package config

import "os"

type Config struct {
	KeycloakURL      string
	KeycloakRealm    string
	KeycloakUsername string
	KeycloakPassword string
}

func LoadConfig() *Config {
	return &Config{
		KeycloakURL:      getEnv("KEYCLOAK_URL", "http://localhost:8080"),
		KeycloakRealm:    getEnv("KEYCLOAK_REALM", "master"),
		KeycloakUsername: getEnv("KEYCLOAK_USERNAME", "admin"),
		KeycloakPassword: getEnv("KEYCLOAK_PASSWORD", "admin"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
