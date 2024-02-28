package main

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This will examine the Config global, and ensure the values match config.yaml
func TestLoadConfig(t *testing.T) {

	assert.IsType(t, GlobalConfig{}, Config, "type check")

	if reflect.DeepEqual(Config, new(GlobalConfig)) {
		t.Error("Config is empty.")
	}

	assert.True(t, Config.Debug, "Configuration - Debug")

	assert.NotEmpty(t, Config.Hostdb.Pass, "Configuration - Hostdb.Pass")
	assert.Equal(t, "https://hostdb.pdxfixit.com/v0", Config.Hostdb.Url, "Configuration - Hostdb.Url")
	assert.Equal(t, "writer", Config.Hostdb.User, "Configuration - Hostdb.User")

	assert.NotEmpty(t, Config.Inventory.Url, "Configuration - Inventory.Url")

	assert.NotEmpty(t, Config.Vcenter.Pass, "Configuration - Vcenter.Pass")
	assert.Equal(t, "username", Config.Vcenter.User, "Configuration - Vcenter.User")

}
