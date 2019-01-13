package main

import "fmt"

var (
	Verbose   = false
	vPrefixes = []string{}
)

func pushVPrefix(pfx string) {
	vPrefixes = append(vPrefixes, pfx)
}

func popVPrefix() {
	if len(vPrefixes) > 0 {
		vPrefixes = vPrefixes[0 : len(vPrefixes)-1]
	}
}

func vPrefix() (pfx string) {
	if len(vPrefixes) > 0 {
		pfx = vPrefixes[len(vPrefixes)-1]
	}
	return
}

func vPrintln(a ...interface{}) {
	if Verbose {
		fmt.Printf("%s", vPrefix())
		fmt.Println(a...)
	}
}

func vPrintf(format string, a ...interface{}) {
	if Verbose {
		fmt.Printf("%s", vPrefix())
		fmt.Printf(format, a...)
	}
}
