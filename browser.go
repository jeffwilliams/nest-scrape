package main

// Control firefox through marionette
// See https://developer.mozilla.org/en-US/docs/Mozilla/QA/Marionette/Protocol
// https://firefox-source-docs.mozilla.org/testing/marionette/marionette/Intro.html#how-does-it-work

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"strings"
	"time"

	mcl "github.com/njasm/marionette_client"
)

func browserArgs(headless bool, profilePath string) (args []string) {
	args = []string{"--marionette"}

	if headless {
		args = append(args, "--headless")
	}

	args = append(args, "--profile")
	args = append(args, profilePath)
	return
}

type BrowserStartOpts struct {
	BrowserPath string
	ShowBrowser bool
}

func StartBrowser(o BrowserStartOpts) (cmd *exec.Cmd, err error) {
	profileDir := os.ExpandEnv("$HOME/.banker/td-firefox-profile")
	err = os.MkdirAll(profileDir, 0755)
	if err != nil {
		err = fmt.Errorf("Failed to make firefox profile directory: %v", err)
		return
	}

	args := browserArgs(!o.ShowBrowser, profileDir)

	vPrintf("browser: running browser using cmd '%s %v'\n", o.BrowserPath, strings.Join(args, " "))
	cmd = exec.Command(o.BrowserPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	return
}

type BrowserConnOpts struct {
	Debug bool
}

func ConnectBrowser(o BrowserConnOpts) (client *mcl.Client, err error) {
	mcl.RunningInDebugMode = o.Debug

	client = mcl.NewClient()

	for i := 0; i < 60; i++ {

		err = client.Connect("", 0) // this are the default marionette values for hostname, and port
		if err == nil {
			break
		}
		vPrintln(err)
		vPrintln("browser: retrying connection in 1 second")
		time.Sleep(1. * time.Second)
	}
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = client.NewSession("", nil) // let marionette generate the Session ID with it's default Capabilities
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func Screenshot(client *mcl.Client, outfile string) error {
	rsp, err := client.Screenshot()

	if err != nil {
		return err
	}

	val := make(map[string]string)
	err = json.Unmarshal([]byte(rsp), &val)
	if err != nil {
		return err
	}

	dec, err := base64.StdEncoding.DecodeString(val["value"])
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(dec)

	img, _, err := image.Decode(buf)
	if err != nil {
		return err
	}

	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		return err
	}

	return nil

}
