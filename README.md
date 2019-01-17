# nest-scrape

This program logs into to the Nest website and retrieves the thermostat, temperature sensor, and humidity measurements, and the external temperature. It's a workaround for the fact that the Nest API doesn't allow retrieval of the temperature sensor measurements.

It works by starting the firefox browser in headless mode, and controlling it via the marionette protocol to perform the login. After login, it scrapes the web UI to retrieve the sensor information and prints it out two lines: a line of comma-separated headings, and a line of comma separated measurements, like so:
  
    Time, Dining Room Thermostat Int. Temp., Bedroom 1 Int. Temp., Bedroom 2 Int. Temp., Upstairs Hallway Int. Temp., Dining Room Humid., Home Ext. Temp.
    Jan 14 14:54:16 2019, 20, 18.5, 18.5, 19.5, 31, -8


# Requirements

  * Linux
  * Firefox (tested with 64.0.2)
  * Go (for compiling. Compiles with Go 1.11)

# Instructions

First, download firefox.

Then, compile the program using Go. If Go is installed, this should be as easy as `go get github.com/jeffwilliams/nest-scrape`. This'll produce a binary named `nest-scrape`.

Next, run `nest-scrape --generate`. This will output a sample config file named `nest.yaml`. Change the permissions on the file so that group and other cannot read or write it (`chmod go-rw nest.yaml`) because it'll contain your password. Then edit the file and enter the correct settings.

Finally, run `nest-scrape` to collect the measurements and print them to stdout.

See `nest-scrape --help` for other options.
