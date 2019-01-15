package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"unicode"

	mcl "github.com/njasm/marionette_client"
)

const (
	nestUrl         = "https://home.nest.com"
	loginFieldId    = "email"
	passwordFieldId = "pass"
	loginButtonId   = "signin"
)

var waitTime = 8 * time.Second

type findFlag uint

const (
	flagWait findFlag = 1 << iota
	flagClick
	flagVisible
)

type ScraperParams struct {
	Login    string
	Password string
}

type Scraper struct{}

func ElementIsVisible(by mcl.By, value string) func(f mcl.Finder) (bool, *mcl.WebElement, error) {
	return func(f mcl.Finder) (bool, *mcl.WebElement, error) {
		result := true
		v, e := f.FindElement(by, value)
		if e != nil || v == nil {
			result = false
		}

		if !v.Displayed() {
			result = false
		}

		return result, v, e
	}
}

func waitBy(client *mcl.Client, fieldNameForLog, val string, by mcl.By) {
	vPrintf("Waiting for %s to load (element %s %s)\n", fieldNameForLog, by, val)
	mcl.Wait(client).For(waitTime).Until(mcl.ElementIsPresent(by, val))
}

func findById(client *mcl.Client, fieldNameForLog, id string, flags findFlag) (elem *mcl.WebElement, err error) {
	return findBy(client, fieldNameForLog, id, flags, mcl.ID)
}

func findBySelector(client *mcl.Client, fieldNameForLog, sel string, flags findFlag) (elem *mcl.WebElement, err error) {
	return findBy(client, fieldNameForLog, sel, flags, mcl.CSS_SELECTOR)
}

func findBy(client *mcl.Client, fieldNameForLog, val string, flags findFlag, by mcl.By) (elem *mcl.WebElement, err error) {
	if flags&flagWait > 0 {
		waitBy(client, fieldNameForLog, val, by)
	}

	if flags&flagVisible > 0 {
		mcl.Wait(client).For(waitTime).Until(ElementIsVisible(by, val))
	}

	vPrintf("Finding %s (element with id %s)\n", fieldNameForLog, val)
	elem, err = client.FindElement(by, val)
	if err != nil {
		err = fmt.Errorf("When finding %s by %s %s, got an error: %s", fieldNameForLog, by, val, err.Error())
		return
	}

	if flags&flagClick > 0 {
		vPrintf("Clicking %s (element %s %s)\n", fieldNameForLog, by, val)
		elem.Click()
	}

	return
}

func sendKeys(elem *mcl.WebElement, what, keys string) (err error) {
	vPrintf("Entering the string '%s' into %s\n", keys, what)
	err = elem.SendKeys(keys)
	if err != nil {
		err = fmt.Errorf("When entering %s, got an error: %s", what, err.Error())
		return
	}
	return
}

func setValueOfActive(client *mcl.Client, value string) (err error) {
	vPrintf("Setting the value of currently active element (i.e. last clicked) '%s'\n", value)
	_, err = client.ExecuteScript(fmt.Sprintf("document.activeElement.value=\"%s\"", value), nil, 5, false)
	return

}

func (s Scraper) Login(client *mcl.Client, parms ScraperParams) (err error) {
	// Handle panics
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	checkErr := func() {
		if err != nil {
			fmt.Println("Error: ", err)
			panic(1)
		}
	}

	vPrintln("Navigating to site")
	client.Navigate(nestUrl)

	vPrintln("Checking if login is required")
	_, err = findBySelector(client, "thermostat location", ".puck-item > a", flagWait)
	if err == nil {
		// The element with class .puck-item already exists, meaning we didn't need to log in.
		vPrintln("Login was not required")
		return
	}

	_, err = findById(client, "login field", loginFieldId, flagWait|flagClick)
	checkErr()

	// For some reason we can't just use SendKeys here because the email field is not visible.
	err = setValueOfActive(client, parms.Login)
	checkErr()

	_, err = findById(client, "password field", passwordFieldId, flagClick)
	checkErr()

	err = setValueOfActive(client, parms.Password)
	checkErr()

	_, err = findById(client, "sign-in button", loginButtonId, flagClick)
	checkErr()

	return
}

func (s Scraper) GetTemperatures(client *mcl.Client, parms ScraperParams) (measurements *Measurements, err error) {
	// Handle panics
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	checkErr := func() {
		if err != nil {
			fmt.Println("Error: ", err)
			panic(1)
		}
	}

	_, err = findBySelector(client, "thermostat location", ".puck-item > a", flagWait|flagClick)
	checkErr()

	vPrintln("Scraping thermostat info")

	// The temperature sensors, inside humidity, and outside temperature are all layed out as HTML element
	// siblings, with headers as <header> elements, and everything else as <divs>. There is nothing to
	// distinguish a thermostat temperature div from a humidity div, or outside temperature div.
	//
	// To get useful information, then, what I do here is take all the siblings and flatten the text under
	// them into an ordered array of values. This is done by selecting the right types of elements all together,
	// then mapping a function to that array that converts the array elements to their contained text.
	//
	// The result is something like:
	// ["TEMPERATURE SENSORS","Dining Room Thermostat","19.5°","Bedroom 1","17.5°","Bedroom 2","17.5°","Upstairs Hallway","19°","INSIDE HUMIDITY","Dining Room","31%","OUTSIDE TEMP.","Ottawa","-16°"]
	//

	resp, err := client.ExecuteScript(`return Array.from(document.querySelectorAll('.card.type-thermostat div[class*="style_title"],div[class*="style_value"],header')).map(function(val){ return val.textContent; })`, nil, 5, false)
	checkErr()

	var result struct {
		Value []string
	}

	vPrintln("Decoding thermostat info")
	err = json.Unmarshal([]byte(resp.Value), &result)
	checkErr()

	vPrintln("Thermostat raw info: ", result.Value)

	vPrintln("Converting sensor info to internal format")
	measurements, err = convertSensorInfo(result.Value)
	checkErr()

	return
}

type Measurement struct {
	Label string
	Value float32
}

func (m *Measurement) parseValue(s string) (err error) {
	var b bytes.Buffer

	// Take characters until the first non-digit
	for _, r := range []rune(s) {
		if !unicode.IsDigit(r) && r != '-' && r != '.' {
			break
		}
		b.WriteRune(r)
	}

	f64, err := strconv.ParseFloat(b.String(), 32)
	m.Value = float32(f64)
	return
}

type Measurements struct {
	InternalTemperatures []Measurement
	ExternalTemperatures []Measurement
	Humidities           []Measurement
}

func convertSensorInfo(raw []string) (measurements *Measurements, err error) {

	const (
		stateUnknown = iota
		stateInternalTemps
		stateHumidity
		stateExternalTemps
	)

	// Information about the section of the raw measurements we are in (external temps, humidity, etc)
	type section struct {
		header string
		// Which slice of measurements we should add the measurements of this section into
		slice *[]Measurement
	}

	measurements = &Measurements{[]Measurement{}, []Measurement{}, []Measurement{}}

	sections := []section{
		stateUnknown:       section{"unknown", nil},
		stateInternalTemps: section{"TEMPERATURE SENSORS", &measurements.InternalTemperatures},
		stateHumidity:      section{"INSIDE HUMIDITY", &measurements.Humidities},
		stateExternalTemps: section{"OUTSIDE TEMP.", &measurements.ExternalTemperatures},
	}

	// Only call this on possible header entries
	nextState := func(hdr string) int {
		for i, sec := range sections {
			if sec.header == hdr {
				return i
			}
		}

		return stateUnknown
	}

	state := stateUnknown
	pairIndex := 0
	var pair [2]string

	for _, v := range raw {
		if state == stateUnknown {
			// Searching for a header
			state = nextState(v)
			pairIndex = 0
			continue
		}

		// If we are on a header, change state
		if ns := nextState(v); ns != stateUnknown {
			state = ns
			pairIndex = 0
			continue
		}

		// Suck up two values
		pair[pairIndex] = v
		pairIndex = 1 - pairIndex
		if pairIndex == 0 {
			// Got both parts of the pair
			msr := Measurement{Label: pair[0]}
			err = msr.parseValue(pair[1])
			if err != nil {
				vPrintf("Error decoding sensor measurement '%s': %v", pair[1], err)
				continue
			}
			*sections[state].slice = append(*sections[state].slice, msr)
		}
	}

	return
}
