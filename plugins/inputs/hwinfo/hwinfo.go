package hwinfo

import (
	"runtime/debug"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/abdfnx/gosh"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/rs/zerolog/log"

	shmem "github.com/zachstence/hwinfo-telegraf-plugin/plugins/inputs/hwinfo/hwinfoShMem"
)

// ============================================================================
// Public input plugin interface
// ============================================================================

type HWiNFOInputPlugin struct {
	hwinfoVersion string
	pluginVersion string
	shmemVersion  string
}

func init() {
	inputs.Add("hwinfo", func() telegraf.Input {
		return &HWiNFOInputPlugin{
			hwinfoVersion: HWiNFOVersion(),
			pluginVersion: PluginVersion(),
		}
	})
}

func (input *HWiNFOInputPlugin) Init() error {
	return nil
}

func (input *HWiNFOInputPlugin) Stop() {
	shmem.UnlockMutex()
}

func (input *HWiNFOInputPlugin) SampleConfig() string {
	return `
[[inputs.hwinfo]]
	# no config
`
}

func (input *HWiNFOInputPlugin) Description() string {
	return "Gather Windows system hardware information from HWiNFO"
}

func (input *HWiNFOInputPlugin) Gather(a telegraf.Accumulator) error {
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

// ============================================================================
// Private helpers
// ============================================================================

type Metric struct {
	fields map[string]interface{}
	tags   map[string]string
}

type SensorReadings struct {
	sensor   shmem.Sensor
	readings []shmem.Reading
}

func (input *HWiNFOInputPlugin) gather() []SensorReadings {
	rawData, err := shmem.Read()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	input.shmemVersion = rawData.Version()

	data := []SensorReadings{}

	// Get sensors
	for _, s := range rawData.Sensors() {
		data = append(data, SensorReadings{
			sensor:   s,
			readings: []shmem.Reading{},
		})
	}
	log.Debug().Msgf("Processed %d sensors", rawData.Header().NumSensorElements())

	// Get readings
	for _, r := range rawData.Readings() {
		sensorIndex := int(r.SensorIndex())
		if sensorIndex >= len(data) {
			log.Error().Msgf("sensor index out of range, attempting to access index %d, but %d sensors found ", sensorIndex, len(data))
			continue
		}

		data[sensorIndex].readings = append(data[sensorIndex].readings, r)
	}
	log.Debug().Msgf("Processed %d readings", rawData.Header().NumReadingElements())

	return data
}

func (hwinfo *HWiNFOInputPlugin) buildFieldsAndTags(sensorReadings SensorReadings) []Metric {
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
