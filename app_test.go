package app

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	tests := []struct {
		name          string
		appVersion    string
		buildVersion  string
		expectedValue string
	}{
		{
			name:          "both versions provided",
			appVersion:    "1.0.0",
			buildVersion:  "abc123",
			expectedValue: "1.0.0_abc123",
		},
		{
			name:          "only app version",
			appVersion:    "1.0.0",
			buildVersion:  "",
			expectedValue: "1.0.0",
		},
		{
			name:          "only build version",
			appVersion:    "",
			buildVersion:  "abc123",
			expectedValue: "abc123",
		},
		{
			name:          "no versions",
			appVersion:    "",
			buildVersion:  "",
			expectedValue: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			originalAppInfo := appInfo
			defer func() {
				appInfo = originalAppInfo
			}()

			appInfo.AppVersion = test.appVersion
			appInfo.BuildVersion = test.buildVersion

			// Test
			result := Version()
			assert.Equal(t, test.expectedValue, result)
		})
	}
}

func TestDebugMode(t *testing.T) {
	// Save original state
	originalAppInfo := appInfo
	defer func() {
		appInfo = originalAppInfo
	}()

	// Test with debug mode true
	appInfo.DebugMode = true
	assert.True(t, DebugMode())

	// Test with debug mode false
	appInfo.DebugMode = false
	assert.False(t, DebugMode())
}

func TestName(t *testing.T) {
	// This is harder to test directly because it depends on os.Args[0]
	// But we can test the functionality
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	tests := []struct {
		executable string
		expected   string
	}{
		{"test_app", "test_app"},
		{"test_app.exe", "test_app"},
		{"/path/to/app", "app"},
		{"/path/to/app.bin", "app"},
	}

	for _, test := range tests {
		t.Run(test.executable, func(t *testing.T) {
			os.Args = []string{test.executable}
			result := Name()
			assert.Equal(t, test.expected, result)
		})
	}
}

func captureOutput(f func()) string {
	// Redirect standard output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function
	f()

	// Reset stdout and read captured output
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}

func TestLogger(t *testing.T) {
	t.Run("LogWith creates logger with context", func(t *testing.T) {
		logger := LogWith("key", "value")
		require.NotNil(t, logger)
		require.NotNil(t, logger.l)
	})

	t.Run("GetLogger returns default logger", func(t *testing.T) {
		logger := GetLogger()
		require.NotNil(t, logger)
		require.NotNil(t, logger.l)
	})

	t.Run("Logger methods don't panic", func(t *testing.T) {
		logger := GetLogger()

		// Test all methods don't panic
		assert.NotPanics(t, func() {
			logger.Debug("debug message", "key", "value")
			logger.Info("info message", "key", "value")
			logger.Warn("warn message", "key", "value")
			logger.Error("error message", "key", "value")
			logger.Print("print message")
			logger.Print("print message with", "key", "value")
		})
	})

	t.Run("LogFatal calls os.Exit", func(t *testing.T) {
		// This test uses a helper program to test os.Exit functionality
		if os.Getenv("TEST_LOGGER_EXIT") == "1" {
			LogFatal("fatal message")
			return
		}

		cmd := exec.Command(os.Args[0], "-test.run=TestLogger")
		cmd.Env = append(os.Environ(), "TEST_LOGGER_EXIT=1")
		err := cmd.Run()

		// We expect the process to exit with status 1
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			return
		}
		t.Fatalf("process ran with err %v, want exit status 1", err)
	})

	t.Run("setLogLevel changes log level", func(t *testing.T) {
		// Capture output after setting debug level
		output := captureOutput(func() {
			setLogLevel(levelDebug)
			LogDebug("test debug message")
		})
		assert.Contains(t, strings.ToLower(output), "debug")
		assert.Contains(t, output, "test debug message")

		// Capture output after setting info level (should not log debug)
		output = captureOutput(func() {
			setLogLevel(levelInfo)
			LogDebug("test debug message")
		})
		assert.Empty(t, output)
	})

	t.Run("Custom level names work", func(t *testing.T) {
		output := captureOutput(func() {
			setLogLevel(levelDebug)
			slog.Log(nil, levelFatal, "test fatal message")
		})
		assert.Contains(t, output, "FATAL")
		assert.Contains(t, output, "test fatal message")
	})
}

func TestInit(t *testing.T) {
	// Save original args and restore after test
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	t.Run("Init with debug flag", func(t *testing.T) {
		os.Args = []string{"app", "-d"}

		output := captureOutput(func() {
			Init("1.0.0", "build123")
		})

		assert.Contains(t, output, "DebugMode=true")
		assert.True(t, appInfo.DebugMode)
		assert.Equal(t, "1.0.0", appInfo.AppVersion)
		assert.Equal(t, "build123", appInfo.BuildVersion)
	})
}
