package hwinfoShMem

// #include "hwisenssm2.h"
import "C"

import (
	"strconv"
	"unsafe"
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
	return DecodeCharPtr(unsafe.Pointer(&s.cs.szSensorNameOrig), C.HWiNFO_SENSORS_STRING_LEN2)
}

// NameUser sensor name displayed, which might have been renamed by user
func (s *Sensor) NameUser() string {
	return DecodeCharPtr(unsafe.Pointer(&s.cs.szSensorNameUser), C.HWiNFO_SENSORS_STRING_LEN2)
}

// TODO I wish there was a better way to do this, ideally something provided explicitly bw HWiNFO
// A dynamic value computed by looking at other fields of the sensor
func (s *Sensor) SensorType() SensorType {
	name := s.NameOrig()

	if StartsWithLower(name, "system") {
		return System
	} else if StartsWithLower(name, "cpu") {
		return CPU
	} else if StartsWithLower(name, "s.m.a.r.t.") {
		return SMART
	} else if StartsWithLower(name, "drive") {
		return Drive
	} else if StartsWithLower(name, "gpu") {
		return GPU
	} else if StartsWithLower(name, "network") {
		return Network
	} else if StartsWithLower(name, "windows") {
		return Windows
	} else if StartsWithLower(name, "memory timings") {
		return MemoryTimings
	} else {
		return Unknown
	}
}
