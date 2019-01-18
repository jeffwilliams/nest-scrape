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
	BrowserPath       string
	Login             string
	Password          string
	BrowserProfileDir string
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

	err = validateConfig(config)

	return
}

func validateConfig(config *Config) error {

	strs := []struct {
		name string
		val  *string
	}{
		{"BrowserPath", &config.BrowserPath},
		{"Login", &config.Login},
		{"Password", &config.Password},
		{"BrowserProfileDir", &config.BrowserProfileDir},
	}

	for _, v := range strs {
		if len(*v.val) == 0 {
			return fmt.Errorf("The %s setting cannot be empty.", v.name)
		}
	}

	return nil
}

func GenConfig() error {
	if _, err := os.Stat(ConfigPath); err == nil {
		return os.ErrExist
	}

	f, err := os.Create(ConfigPath)
	if err != nil {
		return err
	}

	defer f.Close()

	s := `# Sample nest-scraper config file
# browserpath should be set to the absolute path of the firefox executable to run.
browserpath: /path/to/firefox
login: user@domain.com
password: PASSWORD
# Directory where firefox profile should be stored. Environment variables in this are 
# expanded.
browserprofiledir: $HOME/.nest-scrape/firefox-profile
`

	fmt.Fprintln(f, s)
	return nil
}
