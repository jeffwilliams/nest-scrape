package main

import (
	"fmt"
	"time"

	mcl "github.com/njasm/marionette_client"
)

const (
	nestUrl         = "https://home.nest.com"
	loginFieldId    = "email"
	passwordFieldId = "pass"
	loginButtonId   = "signin"
	waitTime        = 5 * time.Second
)

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

// waitById waits for the element with ID `id` to be loaded.
func waitById(client *mcl.Client, fieldNameForLog, id string) {
	vPrintf("Waiting for %s to load (element with id %s)\n", fieldNameForLog, id)
	mcl.Wait(client).For(waitTime).Until(mcl.ElementIsPresent(mcl.By(mcl.ID), id))
}

func findById(client *mcl.Client, fieldNameForLog, id string, flags findFlag) (elem *mcl.WebElement, err error) {
	if flags&flagWait > 0 {
		waitById(client, fieldNameForLog, id)
	}

	if flags&flagVisible > 0 {
		mcl.Wait(client).For(waitTime).Until(ElementIsVisible(mcl.By(mcl.ID), id))
	}

	vPrintf("Finding %s (element with id %s)\n", fieldNameForLog, id)
	elem, err = client.FindElement(mcl.ID, id)
	if err != nil {
		err = fmt.Errorf("When finding %s (element with id %s), got an error: %s", fieldNameForLog, id, err.Error())
		return
	}

	if flags&flagClick > 0 {
		vPrintf("Clicking %s (element with id %s)\n", fieldNameForLog, id)
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

	_, err = findById(client, "login field", loginFieldId, flagWait|flagClick)
	checkErr()

	// For some reason we can't just use SendKeys here because the email field is not visible.
	err = setValueOfActive(client, parms.Login)
	checkErr()

	_, err = findById(client, "password field", passwordFieldId, flagClick)
	checkErr()

	err = setValueOfActive(client, parms.Password)
	checkErr()

	//_, err = client.ExecuteScript(
	//	fmt.Sprintf("document.getElementById('pass').value='%s'", parms.Password), nil, 5, false)
	//checkErr()

	_, err = findById(client, "sign-in button", loginButtonId, flagClick)
	checkErr()

	// document.querySelector(".puck-item > a").click()

	/*
		vPrintln("sleeping 5 seconds to make sure everything loaded")
		time.Sleep(5000 * time.Millisecond)

		vPrintln("checking if secret question prompt is present")
		elem, err = client.FindElement(mcl.ID, "ngdialog1-aria-labelledby")
		if err == nil {
			if strings.ToLower(strings.TrimSpace(elem.Text())) == "confirm your identity" {
				vPrintln("detected secret question prompt. Finding secret question")

				elem, err = client.FindElement(mcl.ID, "labelWrap_301 ")
				if err != nil {
					fmt.Println("Error finding secret question:", err)
					return
				}

				vPrintln("secret question is:", elem.Text())

				q := strings.TrimSpace(elem.Text())
				ans, ok := parms.SecurityQuestions[q]
				if !ok {
					err = fmt.Errorf("No answer configured for the security question '%s'", q)
					fmt.Println(err)
					return
				}

				vPrintln("finding element to answer secret question")
				elem, err = client.FindElement(mcl.ID, "answer")
				if err != nil {
					fmt.Println(err)
					return
				}

				vPrintln("entering answer: ", ans)

				err = elem.SendKeys(ans)
				if err != nil {
					fmt.Println(err)
					return
				}

				vPrintln("finding 'Enter' button")
				elem, err = client.FindElement(mcl.CSS_SELECTOR, "form[name=enterAnswerForm] button")
				if err != nil {
					fmt.Println(err)
					return
				}

				time.Sleep(1 * time.Second)
				vPrintln("clicking 'Enter' button")
				elem.Click()
			}
		}
	*/
	return
}
