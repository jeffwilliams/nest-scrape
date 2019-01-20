package main

import (
	"fmt"
	"os"
	"time"

	flag "github.com/ogier/pflag"
)

var generate = flag.BoolP("generate", "g", false, "Generate a sample config file and exit")
var verbose = flag.IntP("verbose", "v", 0, "Amount of verbosity")
var showBrowser = flag.BoolP("show", "s", false, "Show the web browser being controlled, and don't close it when done.")
var scrshotOnFailure = flag.StringP("failshot", "r", "", "On failure, save a screenshot of the browser to the specified file.")
var onlyLogin = flag.BoolP("login-only", "l", false, "Stop after logging into the bank")
var timeout = flag.IntP("timeout", "t", 8, "Number of seconds to wait until each page element loads")
var formatterName = flag.StringP("format", "f", "csv+hdr", "Output format. One of 'csv+hdr', 'csv', or 'json'")
var showVersion = flag.BoolP("version", "e", false, "Output version and exit")

var version = "undefined"

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Log into to the Nest website and retrieve the thermostat, ")
		fmt.Fprintln(os.Stderr, "temperature sensor, and humidity measurements, and the external")
		fmt.Fprintln(os.Stderr, "temperature.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	format, err := formatterFromName(*formatterName)
	if err != nil {
		fmt.Println(err)
		return
	}

	if *generate {
		if err := GenConfig(); err != nil {
			fmt.Printf("Generating config file failed: %v\n", err)
		}
		return
	}

	if *showVersion {
		fmt.Printf("Version %s\n", version)
		return
	}

	Verbose = *verbose > 0

	waitTime = time.Duration(*timeout) * time.Second

	conf, err := LoadConfig()
	if err != nil {
		fmt.Printf("Loading config failed: %v\n", err)
		return
	}

	_, err = StartBrowser(
		BrowserStartOpts{
			BrowserPath: conf.BrowserPath,
			ShowBrowser: *showBrowser,
			ProfileDir:  conf.BrowserProfileDir,
			Verbose:     *verbose > 0,
		})

	if err != nil {
		fmt.Printf("Starting browser failed: %v\n", err)
		return
	}

	client, err := ConnectBrowser(BrowserConnOpts{Debug: *verbose > 1})

	if !*showBrowser {
		defer client.Quit()
	}

	if err != nil {
		fmt.Printf("Connecting to browser for marionette session failed: %v\n", err)
		return
	}

	parms := ScraperParams{
		Login:    conf.Login,
		Password: conf.Password,
	}

	var scraper Scraper

	err = scraper.Login(client, parms)
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}

	if *onlyLogin {
		return
	}

	measurements, err := scraper.GetTemperatures(client, parms)

	if err != nil {
		fmt.Printf("Scraping failed: %v\n", err)
		return
	}

	str, err := format(measurements)
	if err != nil {
		fmt.Printf("Formatting output failed: %v\n", err)
		return
	}

	fmt.Println(str)

	if !*showBrowser {
		client.Quit()
	}

}
