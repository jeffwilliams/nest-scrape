package main

import "testing"

func TestConvert(t *testing.T) {

	input := []string{"TEMPERATURE SENSORS", "Dining Room Thermostat", "19.5°", "Bedroom 1", "17.5°", "Bedroom 2", "17.5°", "Upstairs Hallway", "19°", "INSIDE HUMIDITY", "Dining Room", "31%", "OUTSIDE TEMP.", "Ottawa", "-16°"}

	expected := Measurements{
		InternalTemperatures: []Measurement{
			Measurement{"Dining Room Thermostat", 19.5},
			Measurement{"Bedroom 1", 17.5},
			Measurement{"Bedroom 2", 17.5},
			Measurement{"Upstairs Hallway", 19},
		},
		Humidities: []Measurement{
			Measurement{"Dining Room", 31},
		},
		ExternalTemperatures: []Measurement{
			Measurement{"Ottawa", -16},
		},
	}

	actual, err := convertSensorInfo(input)

	validateMeasurements := func(expected, actual []Measurement) {
		if len(expected) != len(actual) {
			t.Fatalf("Expected %d measurements but got %d: %v", len(expected), len(actual), actual)
		}

		for i, e := range expected {
			a := actual[i]
			if e.Label != a.Label || e.Value != a.Value {
				t.Fatalf("Expected measurement %v but got %v", e, a)
			}
		}
	}

	if err != nil {
		t.Fatal("Parsing failed:", err)
	}

	validateMeasurements(expected.InternalTemperatures, actual.InternalTemperatures)
	validateMeasurements(expected.Humidities, actual.Humidities)
	validateMeasurements(expected.ExternalTemperatures, actual.ExternalTemperatures)

}
