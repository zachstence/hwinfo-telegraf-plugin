package hwinfo

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/abdfnx/gosh"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/rs/zerolog/log"

	hwinfoInternal "github.com/zachstence/hwinfo-telegraf-plugin/plugins/inputs/hwinfo/internal"
)

// ============================================================================
// Public input plugin interface
// ============================================================================

type HWiNFOInput struct {
	hwinfoVersion string
	pluginVersion string
	shmemVersion  string
}

func init() {
	inputs.Add("hwinfo", func() telegraf.Input {
		return &HWiNFOInput{
			hwinfoVersion: HWiNFOVersion(),
			pluginVersion: PluginVersion(),
			shmemVersion:  SHMemVersion(),
		}
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
	log.Debug().Msg("Converting metrics...")
	writeCount := 0
	for _, datum := range data {
		metrics := input.buildFieldsAndTags(datum)
		for _, metric := range metrics {
			a.AddFields("hwinfo", metric.fields, metric.tags)
			writeCount++
		}
	}
	log.Debug().Msgf("Wrote %d metrics", writeCount)

	log.Debug().Msg("Done gathering metrics")
	return nil
}

func HWiNFOVersion() string {
	err, out, errout := gosh.PowershellOutput("(Get-Process HWiNFO64 | Select-Object Path | Get-Item).VersionInfo.ProductVersion")
	if err != nil {
		log.Debug().Msgf("Failed to query version of HWiNFO64 process: %v, %s", err, errout)
		return "unknown"
	}
	return strings.TrimSpace(out)
}

func PluginVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	i := slices.IndexFunc(bi.Deps, func(module *debug.Module) bool { return module.Path == "github.com/zachstence/hwinfo-telegraf-plugin" })
	if i == -1 {
		return "unknown"
	}
	return bi.Deps[i].Version
}

func SHMemVersion() string {
	rawData, err := hwinfoInternal.Read()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	return fmt.Sprintf("v%d rev%d", rawData.Version(), rawData.Revision())
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
	log.Debug().Msgf("Found %d sensors with a total of %d readings", rawData.NumSensorElements(), rawData.NumReadingElements())

	data := []SensorReadings{}

	// Get sensors
	for s := range rawData.IterSensors() {
		if s.Error != nil {
			log.Error().Err(s.Error).Send()
		} else {
			data = append(data, SensorReadings{
				sensor:   s.Sensor,
				readings: []hwinfoInternal.Reading{},
			})
		}
	}
	log.Debug().Msgf("Processed %d sensors", rawData.NumSensorElements())

	// Get readings
	for r := range rawData.IterReadings() {
		if r.Error != nil {
			log.Error().Err(r.Error).Send()
		} else {
			sensorIndex := int(r.Reading.SensorIndex())
			if sensorIndex < len(data) {
				data[sensorIndex].readings = append(data[sensorIndex].readings, r.Reading)
			} else {
				log.Error().Msgf("sensor index out of range, attempting to access index %d, but %d sensors found ", sensorIndex, len(data))
			}
		}
	}
	log.Debug().Msgf("Processed %d readings", rawData.NumReadingElements())

	return data
}

func (hwinfo *HWiNFOInput) buildFieldsAndTags(sensorReadings SensorReadings) []Metric {
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
			"hwinfoVersion": hwinfo.hwinfoVersion,
			"pluginVersion": hwinfo.pluginVersion,
			"shmemVersion":  hwinfo.shmemVersion,

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
