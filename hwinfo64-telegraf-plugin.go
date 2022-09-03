package main

import (
	"fmt"

	"hwinfo64-telegraf-plugin/hwinfo"
)

type SensorReadings struct {
	sensor   hwinfo.Sensor
	readings []hwinfo.Reading
}

func main() {
	shmem, err := hwinfo.ReadSharedMem()
	if err != nil {
		fmt.Printf("ReadSharedMem failed: %v\n", err)
	}

	data := []SensorReadings{}

	// Get sensors
	for s := range shmem.IterSensors() {
		data = append(data, SensorReadings{
			sensor:   s,
			readings: []hwinfo.Reading{},
		})
	}

	// Get readings
	for r := range shmem.IterReadings() {
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
