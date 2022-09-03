package main

import (
	"fmt"

	"hwinfo-telegraf-plugin/plugins/inputs/hwinfo"
)

type SensorReadings struct {
	sensor   hwinfo.Sensor
	readings []hwinfo.Reading
}

func main() {
	rawData, err := hwinfo.Read()
	if err != nil {
		fmt.Printf("ReadSharedMem failed: %v\n", err)
	}
	fmt.Printf("Found %d sensors and %d total readings", rawData.NumSensorElements(), rawData.NumReadingElements())

	data := []SensorReadings{}

	// Get sensors
	for s := range rawData.IterSensors() {
		data = append(data, SensorReadings{
			sensor:   s,
			readings: []hwinfo.Reading{},
		})
	}

	// Get readings
	for r := range rawData.IterReadings() {
		sensorIndex := int(r.SensorIndex())
		if sensorIndex < len(data) {
			data[sensorIndex].readings = append(data[sensorIndex].readings, r)
		} else {
			fmt.Printf("sensor index out of range, attempting to access index %d, but %d sensors found ", sensorIndex, len(data))
		}
	}

	fmt.Println(data[1].sensor.NameOrig(), data[1].readings[39].LabelOrig(), data[1].readings[39].Value(), data[1].readings[39].Unit())
	// -> CPU [#0]: AMD Ryzen 5 5600X Total CPU Usage 6.158333333333334 %
}
