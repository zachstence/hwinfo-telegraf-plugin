package hwinfo

import (
	"fmt"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	hwinfoInternal "github.com/zachstence/hwinfo-telegraf-plugin/plugins/inputs/hwinfo/internal"
)

// ============================================================================
// Public inpug plugin interface
// ============================================================================

type HWiNFOInput struct{}

func init() {
	inputs.Add("hwinfo", func() telegraf.Input {
		return &HWiNFOInput{}
	})
}

func (input *HWiNFOInput) Init() error {
	return nil
}

func (input *HWiNFOInput) Stop() {
	// Make sure mutex is unlocked before stopping
	hwinfoInternal.UnlockMutex()
}

func (input *HWiNFOInput) SampleConfig() string {
	return `
[[inputs.hwinfo]]
	# no config
`
}

func (input *HWiNFOInput) Description() string {
	return "TODO"
}

func (input *HWiNFOInput) Gather(a telegraf.Accumulator) error {
	// Gather data
	data, err := input.gather()
	if err != nil {
		a.AddError(err)
	}

	// Convert raw data to telegraf fields/tags
	for _, datum := range data {
		metrics := buildFieldsAndTags(datum)
		for _, metric := range metrics {
			a.AddFields("hwinfo", metric.fields, metric.tags)
		}
	}

	return nil
}

// ============================================================================
// Private helpers
// ============================================================================

type Metric struct {
	fields map[string]interface{}
	tags   map[string]string
}

type SensorReadings struct {
	sensor   hwinfoInternal.Sensor
	readings []hwinfoInternal.Reading
}

func (input *HWiNFOInput) gather() ([]SensorReadings, error) {
	rawData, err := hwinfoInternal.Read()
	if err != nil {
		fmt.Printf("ReadSharedMem failed: %v\n", err)
	}
	fmt.Printf("Found %d sensors and %d total readings", rawData.NumSensorElements(), rawData.NumReadingElements())

	data := []SensorReadings{}

	// Get sensors
	for s := range rawData.IterSensors() {
		data = append(data, SensorReadings{
			sensor:   s,
			readings: []hwinfoInternal.Reading{},
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

	return data, nil
}

func buildFieldsAndTags(sensorReadings SensorReadings) []Metric {
	metrics := []Metric{}

	sensor := sensorReadings.sensor
	readings := sensorReadings.readings

	for _, reading := range readings {
		readingType := reading.Type().String()
		readingValue := reading.Value()

		fields := map[string]interface{}{
			(readingType): readingValue,
		}

		tags := map[string]string{
			"sensorId":       sensor.ID(),
			"sensorInst":     strconv.FormatUint(sensor.SensorInst(), 16),
			"sensorNameOrig": sensor.NameOrig(),
			"sensorName":     sensor.NameUser(),

			"readingId":       strconv.FormatInt(int64(reading.ID()), 10),
			"readingNameOrig": reading.LabelOrig(),
			"readingName":     reading.LabelUser(),
			"unit":            reading.Unit(),
		}

		metrics = append(metrics, Metric{
			fields: fields,
			tags:   tags,
		})
	}

	return metrics
}
