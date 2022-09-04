package hwinfo

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	hwinfoInternal "github.com/zachstence/hwinfo-telegraf-plugin/plugins/inputs/hwinfo/internal"
	"github.com/zachstence/hwinfo-telegraf-plugin/plugins/inputs/hwinfo/internal/mutex"
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
	mutex.Unlock()
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
	fields, tags, err := input.gather()
	if err != nil {
		a.AddError(err)
	}

	a.AddFields("hwinfo", fields, tags)
	return nil
}

// ============================================================================
// Private helpers
// ============================================================================

type SensorReadings struct {
	sensor   hwinfoInternal.Sensor
	readings []hwinfoInternal.Reading
}

func (input *HWiNFOInput) gather() (map[string]interface{}, map[string]string, error) {
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

	fmt.Println(data[1].sensor.NameOrig(), data[1].readings[39].LabelOrig(), data[1].readings[39].Value(), data[1].readings[39].Unit())
	// -> CPU [#0]: AMD Ryzen 5 5600X Total CPU Usage 6.158333333333334 %

	// TODO translate raw data to telegraf fields/tags
	return map[string]interface{}{
		"test": "value",
	}, map[string]string{}, nil
}
