package configloader

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Structure to bind application parameters
type Config struct {
	LogLevel string `mapstructure:"LOG_LEVEL"` // logrus library log level to be assigned
}

// Initialize default parameters values
func initDefaultConfiguration() {
	viper.SetDefault("LOG_LEVEL", "debug")
}

// Load configuration from env file
func LoadConfiguration(applicationName string, configurationFilePath string) (config Config, err error) {
	initDefaultConfiguration()

	if configurationFilePath == "" {
		// Read the volume root path
		root := filepath.VolumeName(".")
		if root == "" {
			root = string(filepath.Separator)
		}

		// Set configuration named config from etc/*appName*, $HOME/.*appName* or current folders
		viper.AddConfigPath(filepath.Join(root, "etc", applicationName))
		viper.AddConfigPath(filepath.Join("$HOME", "."+applicationName))
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	} else {
		// Set the configuration file path
		viper.SetConfigFile(configurationFilePath)
	}

	// Get configuration from environment variables, if set
	viper.AutomaticEnv()

	// Get configuration from configuration file, if set
	if configError := viper.ReadInConfig(); configError != nil {
		logrus.Warn(configError.Error())
	}
	err = viper.Unmarshal(&config)

	return
}
