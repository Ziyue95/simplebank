package util

import "github.com/spf13/viper"

// Config stores all configuration of the application
// The values are read by viper from a config file or environment variables.
type Config struct {
	// Viper uses the mapstructure package under the hood for unmarshaling values,
	// so we use the mapstructure tags to specify the name of each config field
	// must use the exact name of each variable as being declared in the app.env
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

// LoadConfig reads configurations from a config file inside the path if it exists,
// or override their values with environment variables if they are provided;\
func LoadConfig(path string) (config Config, err error) {
	// 1. tell Viper the location of the config file
	viper.AddConfigPath(path)
	// tell Viper to look for config file with name "app" and type "env"
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	// 2. tell Viper to read values from environment variables
	// viper.AutomaticEnv() to tell viper to automatically override values
	// that it has read from config file with the values of the corresponding environment variables if they exist
	viper.AutomaticEnv()

	// reading config values using viper.ReadInConfig()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	// unmarshals the values into the target config object
	err = viper.Unmarshal(&config)
	return
}
