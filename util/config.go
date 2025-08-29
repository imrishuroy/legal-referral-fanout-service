package util

import "github.com/spf13/viper"

type Config struct {
	DBDriver         string `mapstructure:"DB_DRIVER"`
	DBSource         string `mapstructure:"DB_SOURCE"`
	ServerAddress    string `mapstructure:"SERVER_ADDRESS"`
	BootStrapServers string `mapstructure:"BOOTSTRAP_SERVERS"`
	SecurityProtocol string `mapstructure:"SECURITY_PROTOCOL"`
	SASLMechanism    string `mapstructure:"SASL_MECHANISM"`
	SASLUsername     string `mapstructure:"SASL_USERNAME"`
	SASLPassword     string `mapstructure:"SASL_PASSWORD"`
	Topic            string `mapstructure:"TOPIC"`
	SQSURL           string `mapstructure:"SQS_URL"`
}

// LoadConfig reads configuration from file or environment variables
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	// Always allow environment variables to override
	viper.AutomaticEnv()

	// Try to read config file if present; ignore missing file
	_ = viper.ReadInConfig()

	err = viper.Unmarshal(&config)
	return
}
