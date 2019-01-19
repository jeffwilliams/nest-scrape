package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type formatter func(measurements *Measurements) (string, error)

func formatAsCsvNoHeader(measurements *Measurements) (string, error) {
	var buf bytes.Buffer

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
	buf.Truncate(buf.Len() - 2) // Undo last comma and space

	return buf.String(), nil
}

func formatAsCsv(measurements *Measurements) (string, error) {
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
	buf.Truncate(buf.Len() - 2) // Undo last comma and space
	buf.WriteRune('\n')

	s, _ := formatAsCsvNoHeader(measurements)
	buf.Write([]byte(s))

	return buf.String(), nil
}

func formatAsJson(measurements *Measurements) (s string, err error) {
	b, err := json.Marshal(measurements)
	return string(b), err
}

func formatterFromName(name string) (formatter, error) {
	switch name {
	case "csv+hdr":
		return formatAsCsv, nil
	case "csv":
		return formatAsCsvNoHeader, nil
	case "json":
		return formatAsJson, nil
	default:
		return nil, fmt.Errorf("Unknown formatter '%s'", name)
	}
}
