package hwinfo

// #include "hwisenssm2.h"
import "C"

import (
	"strconv"
	"unsafe"

	"github.com/zachstence/hwinfo-telegraf-plugin/plugins/inputs/hwinfo/internal/util"
)

// Sensor element (e.g. motherboard, cpu, gpu...)
type Sensor struct {
	cs C.PHWiNFO_SENSORS_SENSOR_ELEMENT
}

// Type of sensor (the kind of hardware, i.e. cpu, gpu, drive)
type SensorType string

const (
	System        SensorType = "system"
	CPU           SensorType = "cpu"
	SMART         SensorType = "smart"
	Drive         SensorType = "drive"
	GPU           SensorType = "gpu"
	Network       SensorType = "network"
	Windows       SensorType = "windows"
	MemoryTimings SensorType = "memory-timings"
	Unknown       SensorType = "unknown"
)

// NewSensor constructs a Sensor
func NewSensor(data []byte) Sensor {
	return Sensor{
		cs: C.PHWiNFO_SENSORS_SENSOR_ELEMENT(unsafe.Pointer(&data[0])),
	}
}

// SensorID a unique Sensor ID
func (s *Sensor) SensorID() uint64 {
	return uint64(s.cs.dwSensorID)
}

// SensorInst the instance of the sensor (together with SensorID forms a unique ID)
func (s *Sensor) SensorInst() uint64 {
	return uint64(s.cs.dwSensorInst)
}

// ID a unique ID combining SensorID and SensorInst
func (s *Sensor) ID() string {
	// keeping old method used in legacy steam deck plugin
	return strconv.FormatUint(s.SensorID()*100+s.SensorInst(), 10)
}

// NameOrig original name of sensor
func (s *Sensor) NameOrig() string {
	return util.DecodeCharPtr(unsafe.Pointer(&s.cs.szSensorNameOrig), C.HWiNFO_SENSORS_STRING_LEN2)
}

// NameUser sensor name displayed, which might have been renamed by user
func (s *Sensor) NameUser() string {
	return util.DecodeCharPtr(unsafe.Pointer(&s.cs.szSensorNameUser), C.HWiNFO_SENSORS_STRING_LEN2)
}

// TODO I wish there was a better way to do this, ideally something provided explicitly bw HWiNFO
// A dynamic value computed by looking at other fields of the sensor
func (s *Sensor) SensorType() SensorType {
	name := s.NameOrig()

	if util.StartsWithLower(name, "system") {
		return System
	} else if util.StartsWithLower(name, "cpu") {
		return CPU
	} else if util.StartsWithLower(name, "s.m.a.r.t.") {
		return SMART
	} else if util.StartsWithLower(name, "drive") {
		return Drive
	} else if util.StartsWithLower(name, "gpu") {
		return GPU
	} else if util.StartsWithLower(name, "network") {
		return Network
	} else if util.StartsWithLower(name, "windows") {
		return Windows
	} else if util.StartsWithLower(name, "memory timings") {
		return MemoryTimings
	} else {
		return Unknown
	}
}
