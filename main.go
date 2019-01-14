package main

import (
	"bytes"
	"fmt"
	"time"

	flag "github.com/ogier/pflag"
)

var verbose = flag.IntP("verbose", "v", 0, "Amount of verbosity")
var showBrowser = flag.BoolP("show", "s", false, "Show the web browser being controlled, and don't close it when done.")
var scrshotOnFailure = flag.StringP("failshot", "r", "", "On failure, save a screenshot of the browser to the specified file.")
var onlyLogin = flag.BoolP("login-only", "l", false, "Stop after logging into the bank")

func main() {
	flag.Parse()

	Verbose = *verbose > 0

	conf, err := LoadConfig()
	if err != nil {
		fmt.Printf("Loading config failed: %v\n", err)
		return
	}

	_, err = StartBrowser(
		BrowserStartOpts{
			BrowserPath: conf.BrowserPath,
			ShowBrowser: *showBrowser,
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

	var format formatter
	format = formatAsCsv

	fmt.Println(format(measurements))

}

type formatter func(measurements *Measurements) string

func formatAsCsv(measurements *Measurements) string {
	var buf bytes.Buffer

	// Header
	fmt.Fprintf(&buf, "Time, ")
	for _, v := range measurements.InternalTemperatures {
		fmt.Fprintf(&buf, "%s Int. Temp., ", v.Label)
	}
	for _, v := range measurements.Humidities {
		fmt.Fprintf(&buf, "%s Humid., ", v.Label)
	}
	for _, v := range measurements.ExternalTemperatures {
		fmt.Fprintf(&buf, "%s Ext. Temp., ", v.Label)
	}
	buf.WriteRune('\n')

	fmt.Fprintf(&buf, "%v, ", time.Now().Format("Jan 2 15:04:05 2006"))
	for _, v := range measurements.InternalTemperatures {
		fmt.Fprintf(&buf, "%v, ", v.Value)
	}
	for _, v := range measurements.Humidities {
		fmt.Fprintf(&buf, "%v, ", v.Value)
	}
	for _, v := range measurements.ExternalTemperatures {
		fmt.Fprintf(&buf, "%v, ", v.Value)
	}
	buf.WriteRune('\n')

	return buf.String()
}
