package configuration_test

import (
	"bytes"
	"prevtorrent/internal/preview/platform/configuration"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfiguration(t *testing.T) {
	expectedConfig := configuration.Config{
		ImageDir:              "ImageDir",
		SqlitePath:            "SqlitePath",
		EnableIPv6:            true,
		EnableUTP:             true,
		EnableTorrentDebug:    true,
		LogLevel:              "debug",
		ConnectionsPerTorrent: 100,
		TorrentListeningPort:  1234,
		TorrentStorageDriver:  "TorrentStorageDriver",
		PubSubDriver:          "PubSubDriver",
		GooglePubSubProjectID: "GooglePubSubProjectID",
		AMQPURI:               "AMQPURI",
	}

	config, err := configuration.NewConfig()
	assert.NoError(t, err)

	assert.Equal(t, expectedConfig, config)
}

func TestConfiguration_Print(t *testing.T) {
	config, err := configuration.NewConfig()
	assert.NoError(t, err)

	buf := new(bytes.Buffer)
	config.Print(buf)
	assert.NotEmpty(t, buf)
}

func TestConfiguration_GetTorrentConf(t *testing.T) {
	config, err := configuration.NewConfig()
	assert.NoError(t, err)
	config.TorrentStorageDriver = "inmemory"

	torrentConfig := configuration.GetTorrentConf(config)
	assert.NotNil(t, torrentConfig)
}
