package main

import (
	"encoding/json"
	"io/ioutil"
)

type hostConfig struct {
	Host string `json:"host"`
}

type proxyConfig struct {
	DefaultHost hostConfig            `json:"default_host"`
	Paths       map[string]hostConfig `json:"paths"`
	Filename    string                `json:"filename"`
}

var (
	httprConfig proxyConfig
)

func loadConfig() error {
	var b []byte
	var err error

	if b, err = ioutil.ReadFile("/etc/httprecorder/httprecorder.json"); err != nil {
		if b, err = ioutil.ReadFile("./httprecorder.json"); err != nil {
			return err
		}
	}
	log.Debug().Println("config content", string(b))
	return json.Unmarshal(b, &httprConfig)
}
