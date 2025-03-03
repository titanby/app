package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestConfig struct {
	Name    string `json:"name"`
	Version string `json:"version" default:"1.0.0"`
	Count   int    `json:"count" default:"10"`
	Enabled bool   `json:"enabled"`
	Nested  struct {
		Value  string  `json:"value" default:"default"`
		Number float64 `json:"number" default:"3.14"`
	} `json:"nested"`
	Items []struct {
		ID   int    `json:"id" default:"1"`
		Name string `json:"name" default:"item"`
	} `json:"items"`
}

func TestReadConfigFile(t *testing.T) {
	t.Run("successful read", func(t *testing.T) {
		// Create temporary config file
		content := `{"name":"test-app","version":"2.0.0","enabled":true}`
		tmpFile, err := os.CreateTemp("", "config-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(content)
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())

		// Test reading the file
		err = readConfigFile(tmpFile.Name())
		assert.NoError(t, err)
		assert.Equal(t, []byte(content), config)
	})

	t.Run("file not found", func(t *testing.T) {
		err := readConfigFile("non-existent-file.json")
		assert.Error(t, err)
	})

	t.Run("invalid file permissions", func(t *testing.T) {
		// Skip on Windows as permissions work differently
		if os.Getenv("GOOS") == "windows" {
			t.Skip("Skipping permissions test on Windows")
		}

		tmpFile, err := os.CreateTemp("", "config-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		require.NoError(t, tmpFile.Close())
		require.NoError(t, os.Chmod(tmpFile.Name(), 0000))
		defer os.Chmod(tmpFile.Name(), 0644)

		err = readConfigFile(tmpFile.Name())
		assert.Error(t, err)
	})
}

func TestGetConfig(t *testing.T) {
	t.Run("valid config unmarshal", func(t *testing.T) {
		// Set mock config data
		config = []byte(`{
			"name": "test-app",
			"version": "2.0.0",
			"count": 5,
			"enabled": true,
			"nested": {
				"value": "custom",
				"number": 2.71
			},
			"items": [
				{"id": 42, "name": "item1"},
				{"id": 0, "name": ""}
			]
		}`)

		var cfg TestConfig
		err := GetConfig(&cfg)

		assert.NoError(t, err)
		assert.Equal(t, "test-app", cfg.Name)
		assert.Equal(t, "2.0.0", cfg.Version)
		assert.Equal(t, 5, cfg.Count)
		assert.True(t, cfg.Enabled)
		assert.Equal(t, "custom", cfg.Nested.Value)
		assert.Equal(t, 2.71, cfg.Nested.Number)
		assert.Len(t, cfg.Items, 2)
		assert.Equal(t, 42, cfg.Items[0].ID)
		assert.Equal(t, "item1", cfg.Items[0].Name)
		assert.Equal(t, 1, cfg.Items[1].ID)        // Default value applied
		assert.Equal(t, "item", cfg.Items[1].Name) // Default value applied
	})

	t.Run("invalid json", func(t *testing.T) {
		config = []byte(`{"name": "broken"}`)

		var cfg TestConfig
		err := GetConfig(&cfg)

		assert.NoError(t, err)
		assert.Equal(t, "broken", cfg.Name)
		assert.Equal(t, "1.0.0", cfg.Version) // Default applied
	})

	t.Run("malformed json", func(t *testing.T) {
		config = []byte(`{"name": "test-app"`)

		var cfg TestConfig
		err := GetConfig(&cfg)

		assert.Error(t, err)
	})
}

func TestSetDefaults(t *testing.T) {
	t.Run("defaults are applied correctly", func(t *testing.T) {
		var cfg TestConfig
		config = []byte(`{"name": "test-app"}`)

		err := GetConfig(&cfg)

		assert.NoError(t, err)
		assert.Equal(t, "test-app", cfg.Name)
		assert.Equal(t, "1.0.0", cfg.Version)
		assert.Equal(t, 10, cfg.Count)
		assert.Equal(t, "default", cfg.Nested.Value)
		assert.Equal(t, 3.14, cfg.Nested.Number)
	})

	t.Run("defaults don't override existing values", func(t *testing.T) {
		var cfg TestConfig
		config = []byte(`{
			"name": "test-app",
			"version": "2.0.0", 
			"count": 5,
			"nested": {
				"value": "custom"
			}
		}`)

		err := GetConfig(&cfg)

		assert.NoError(t, err)
		assert.Equal(t, "test-app", cfg.Name)
		assert.Equal(t, "2.0.0", cfg.Version)       // Not default
		assert.Equal(t, 5, cfg.Count)               // Not default
		assert.Equal(t, "custom", cfg.Nested.Value) // Not default
		assert.Equal(t, 3.14, cfg.Nested.Number)    // Default
	})

	t.Run("test all default types", func(t *testing.T) {
		type AllTypes struct {
			Int       int     `default:"42"`
			Int8      int8    `default:"8"`
			Int16     int16   `default:"16"`
			Int32     int32   `default:"32"`
			Int64     int64   `default:"64"`
			Uint      uint    `default:"42"`
			Uint8     uint8   `default:"8"`
			Uint16    uint16  `default:"16"`
			Uint32    uint32  `default:"32"`
			Uint64    uint64  `default:"64"`
			Float32   float32 `default:"3.2"`
			Float64   float64 `default:"6.4"`
			String    string  `default:"default"`
			NoDefault bool    // No default tag
		}

		var data AllTypes
		config = []byte(`{}`)

		err := GetConfig(&data)

		assert.NoError(t, err)
		assert.Equal(t, 42, data.Int)
		assert.Equal(t, int8(8), data.Int8)
		assert.Equal(t, int16(16), data.Int16)
		assert.Equal(t, int32(32), data.Int32)
		assert.Equal(t, int64(64), data.Int64)
		assert.Equal(t, uint(42), data.Uint)
		assert.Equal(t, uint8(8), data.Uint8)
		assert.Equal(t, uint16(16), data.Uint16)
		assert.Equal(t, uint32(32), data.Uint32)
		assert.Equal(t, uint64(64), data.Uint64)
		assert.Equal(t, float32(3.2), data.Float32)
		assert.Equal(t, 6.4, data.Float64)
		assert.Equal(t, "default", data.String)
		assert.False(t, data.NoDefault) // No default applied
	})

	t.Run("zero values are replaced with defaults", func(t *testing.T) {
		var cfg TestConfig
		config = []byte(`{"count": 0, "version": ""}`)

		err := GetConfig(&cfg)

		assert.NoError(t, err)
		assert.Equal(t, "1.0.0", cfg.Version) // Default applied to empty string
		assert.Equal(t, 10, cfg.Count)        // Default applied to zero int
	})
}

func TestConfigIntegration(t *testing.T) {
	// Integration test with file read and config parsing
	content := `{
		"name": "integration-test",
		"enabled": true,
		"items": [
			{"id": 1, "name": "first"},
			{"id": 2}
		]
	}`

	tmpFile, err := os.CreateTemp("", "config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	err = readConfigFile(tmpFile.Name())
	require.NoError(t, err)

	var cfg TestConfig
	err = GetConfig(&cfg)

	assert.NoError(t, err)
	assert.Equal(t, "integration-test", cfg.Name)
	assert.Equal(t, "1.0.0", cfg.Version) // Default
	assert.Equal(t, 10, cfg.Count)        // Default
	assert.True(t, cfg.Enabled)
	assert.Len(t, cfg.Items, 2)
	assert.Equal(t, 1, cfg.Items[0].ID)
	assert.Equal(t, "first", cfg.Items[0].Name)
	assert.Equal(t, 2, cfg.Items[1].ID)
	assert.Equal(t, "item", cfg.Items[1].Name) // Default
}
