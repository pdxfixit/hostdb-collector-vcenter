package main

import "encoding/json"

type HostdbBulkResponse struct {
	Id       string `json:"id"`
	Hostname string `json:"hostname,omitempty"`
	Ok       bool   `json:"ok"`
	Error    string `json:"error,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type HostdbBulkResults struct {
	Results []HostdbBulkResponse `json:"results"`
}

type HostdbConfig struct {
	Pass string `mapstructure:"pass"`
	Url  string `mapstructure:"url"`
	User string `mapstructure:"user"`
}

type HostdbDocument struct {
	Id        string                 `json:"id,omitempty"`
	Type      string                 `json:"type,omitempty"`
	Hostname  string                 `json:"hostname,omitempty"`
	Ip        string                 `json:"ip,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
	Committer string                 `json:"committer,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Data      json.RawMessage        `json:"data,omitempty"`
	Hash      string                 `json:"hash,omitempty"`
}

type HostdbDocumentSet struct {
	Type      string                 `json:"type"`
	Timestamp string                 `json:"timestamp"`
	Context   map[string]interface{} `json:"context"`
	Committer string                 `json:"committer,omitempty"`
	Records   []HostdbDocument       `json:"records"`
}

type Inventory struct {
	Protocols []string        `json:"protocols"`
	Types     []InventoryType `json:"types"`
	Urls      []struct {
		Url      string        `json:"url"`
		Desc     string        `json:"desc"`
		Type     InventoryType `json:"type"`
		Location string        `json:"location"`
		Name     string        `json:"name"`
	} `json:"urls"`
	Locations []string `json:"locations"`
}

type InventoryConfig struct {
	Url string `mapstructure:"url"`
}

type InventoryType struct {
	Name      string   `json:"name"`
	Protocols []string `json:"protocols"`
}

type VcenterConfig struct {
	Pass string `mapstructure:"pass"`
	User string `mapstructure:"user"`
}

type VcenterList struct {
	Value []struct {
		MemorySize int    `json:"memory_size_MiB"`
		Id         string `json:"vm"`
		Name       string `json:"name"`
		PowerState string `json:"power_state"`
		CpuCount   int    `json:"cpu_count"`
	} `json:"value"`
}

type VcenterSessionToken struct {
	Value string `json:"value"`
}

type VcenterVm struct {
	Value json.RawMessage `json:"value"`
}
