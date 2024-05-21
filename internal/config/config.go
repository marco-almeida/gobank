package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	PostgresUser         string        `mapstructure:"POSTGRES_USER"`
	PostgresHost         string        `mapstructure:"POSTGRES_HOST"`
	PostgresPort         int           `mapstructure:"POSTGRES_PORT"`
	PostgresPassword     string        `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDatabase     string        `mapstructure:"POSTGRES_DB"`
	Environment          string        `mapstructure:"ENVIRONMENT"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	RedisAddress         string        `mapstructure:"REDIS_ADDRESS"`
	HTTPServerAddress    string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	GRPCServerAddress    string        `mapstructure:"GRPC_SERVER_ADDRESS"`
	MigrationURL         string        `mapstructure:"MIGRATION_URL"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	EmailSenderName      string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress   string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword  string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (Config, error) {
	config := Config{}
	viper.AddConfigPath(path)

	configName := "development"
	// if MYBANK_ENV is set, use that
	if env, ok := os.LookupEnv("MYBANK_ENV"); ok {
		configName = env
	}
	viper.SetConfigName(configName)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return config, err
	}

	err = viper.Unmarshal(&config)
	fmt.Println("config", config)
	return config, err
}
