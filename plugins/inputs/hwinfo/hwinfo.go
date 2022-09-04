package hwinfo

import (
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/rs/zerolog/log"

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
	return "Gather Windows system hardware information from HWiNFO"
}

func (input *HWiNFOInput) Gather(a telegraf.Accumulator) error {
	log.Debug().Msg("Gathering metrics...")

	// Gather data
	data := input.gather()

	// Convert raw data to telegraf fields/tags
	for _, datum := range data {
		metrics := buildFieldsAndTags(datum)
		for _, metric := range metrics {
			a.AddFields("hwinfo", metric.fields, metric.tags)
		}
	}

	log.Debug().Msg("Done gathering metrics")
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

func (input *HWiNFOInput) gather() []SensorReadings {
	rawData, err := hwinfoInternal.Read()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Debug().Msgf("Found %d sensors and %d readings", rawData.NumSensorElements(), rawData.NumReadingElements())

	data := []SensorReadings{}

	// Get sensors
	sensors, errs := rawData.IterSensors()
	for e := range errs {
		log.Error().Err(e).Send()
	}

	for s := range sensors {
		data = append(data, SensorReadings{
			sensor:   s,
			readings: []hwinfoInternal.Reading{},
		})
	}

	// Get readings
	readings, errs := rawData.IterReadings()
	for e := range errs {
		log.Error().Err(e).Send()
	}

	for r := range readings {
		sensorIndex := int(r.SensorIndex())
		if sensorIndex < len(data) {
			data[sensorIndex].readings = append(data[sensorIndex].readings, r)
		} else {
			log.Error().Msgf("sensor index out of range, attempting to access index %d, but %d sensors found ", sensorIndex, len(data))
		}
	}

	return data
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
			"sensorInst":     strconv.FormatUint(sensor.SensorInst(), 10),
			"sensorType":     string(sensor.SensorType()),
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
