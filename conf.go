package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type BankConfig struct {
	Login    string
	Password string
}

type Config struct {
	BrowserPath string
	Login       string
	Password    string
}

const (
	ConfigPath = "nest.yaml"
)

func LoadConfig() (config *Config, err error) {

	fi, err := os.Stat(ConfigPath)
	if err != nil {
		return
	}

	// Make sure that group and other have no permissions on the file
	perms := fi.Mode().Perm()
	if perms&077 != 0 {
		err = fmt.Errorf("The permissions on the config file must not allow group or other any access")
		return
	}

	raw, err := ioutil.ReadFile(ConfigPath)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(raw, &config)
	if err != nil {
		return
	}

	return
}