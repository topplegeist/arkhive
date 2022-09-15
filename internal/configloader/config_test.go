package configloader_test

import (
	"os"
	"testing"

	"arkhive.dev/launcher/internal/configloader"
)

// Test default configuration loading
func TestLoadDefaultConfiguration(t *testing.T) {
	configuration, err := configloader.LoadConfiguration("unexistent", "")
	if err != nil {
		t.Fatal(err)
	}
	if configuration.LogLevel != "debug" {
		t.Errorf("Default log level is \"%s\", not \"%s\"", configuration.LogLevel, "debug")
	}
}

// Test environment variables configuration loading
func TestLoadEnvironmentVariablesConfiguration(t *testing.T) {
	os.Setenv("LOG_LEVEL", "LOG_LEVEL")

	configuration, err := configloader.LoadConfiguration("unexistent", "")
	if err != nil {
		t.Fatal(err)
	}
	if configuration.LogLevel != "LOG_LEVEL" {
		t.Errorf("Default log level is \"%s\", not \"%s\"", configuration.LogLevel, "LOG_LEVEL")
	}
}
