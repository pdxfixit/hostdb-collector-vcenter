package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

type GlobalConfig struct {
	Debug     bool            `mapstructure:"debug"`
	Hostdb    HostdbConfig    `mapstructure:"hostdb"`
	Inventory InventoryConfig `mapstructure:"inventory"`
	Vcenter   VcenterConfig   `mapstructure:"vcenter"`
}

var Config GlobalConfig

func loadConfig() {

	log.Println("Loading configuration...")

	// load the config
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/hostdb-collector-vcenter")
	viper.AddConfigPath(".")

	// load env vars
	viper.SetEnvPrefix("hostdb_collector_vcenter")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// read the config file, and handle any errors
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// unmarshal into our struct
	if err := viper.Unmarshal(&Config); err != nil {
		log.Fatal(fmt.Errorf("unable to decode into struct, %v", err))
	}

	//
	// validation
	//

	// validate the HostDB url
	if _, err := url.ParseRequestURI(Config.Hostdb.Url); err != nil {
		log.Fatal("HostDB URL is invalid.")
	}

	// validate the inventory url
	if _, err := url.ParseRequestURI(Config.Inventory.Url); err != nil {
		log.Fatal("Inventory URL is invalid.")
	}

}
