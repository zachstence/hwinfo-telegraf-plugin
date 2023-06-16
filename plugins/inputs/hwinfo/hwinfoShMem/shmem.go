package hwinfoShMem

// #include "hwisenssm2.h"
import "C"

import (
	"fmt"

	"github.com/hidez8891/shm"
	"github.com/rs/zerolog/log"
)

type HWiNFOShMem struct {
	header   *Header
	sensors  []Sensor
	readings []Reading
}

func Read() (*HWiNFOShMem, error) {
	// Lock mutex and unlock after we are done reading
	LockMutex()
	defer UnlockMutex()

	header, err := ReadHeader()
	if err != nil {
		return nil, err
	}

	sensors, err := ReadSensors(header)
	if err != nil {
		return nil, err
	}

	readings, err := ReadReadings(header)
	if err != nil {
		return nil, err
	}

	shmem := &HWiNFOShMem{header, sensors, readings}
	return shmem, nil
}

func ReadHeader() (*Header, error) {
	log.Debug().Msgf("Reading shared memory header")

	bytes, err := readSharedMemory(0, headerLength)
	if err != nil {
		return nil, err
	}

	header := NewHeader(bytes)
	return &header, nil
}

func ReadSensors(header *Header) ([]Sensor, error) {
	offset := header.OffsetOfSensorSection()
	numSensors := header.NumSensorElements()
	length := numSensors * header.SizeOfSensorElement()
	log.Debug().Msgf("Reading %d sensors (%d bytes)", numSensors, length)

	bytes, err := readSharedMemory(offset, length)
	if err != nil {
		return nil, err
	}

	var sensors []Sensor
	for i := 0; i < numSensors; i++ {
		start := i * header.SizeOfSensorElement()
		end := start + header.SizeOfSensorElement()
		sensor := NewSensor(bytes[start:end])
		sensors = append(sensors, sensor)
	}

	return sensors, nil
}

func ReadReadings(header *Header) ([]Reading, error) {
	offset := header.OffsetOfReadingSection()
	numReadings := header.NumReadingElements()
	length := numReadings * header.SizeOfReadingElement()
	log.Debug().Msgf("Reading %d readings (%d bytes)", numReadings, length)

	bytes, err := readSharedMemory(offset, length)
	if err != nil {
		return nil, err
	}

	var readings []Reading
	for i := 0; i < numReadings; i++ {
		start := i * header.SizeOfReadingElement()
		end := start + header.SizeOfReadingElement()
		reading := NewReading(bytes[start:end])
		readings = append(readings, reading)
	}

	return readings, nil
}

func isAccessDeniedErr(err error) bool {
	errStr := fmt.Sprintf("%v", err)
	return errStr == "CreateFileMapping: Access is denied."
}

func readSharedMemory(start int, size int) ([]byte, error) {
	memory, err := shm.Open(C.HWiNFO_SENSORS_MAP_FILE_NAME2, int32(size))
	if err != nil {
		if isAccessDeniedErr(err) {
			log.Fatal().Err(err).Msg("could not access HWiNFO shared memory, is this plugin running as Administrator?")
		}
		return nil, err
	}

	bytes := make([]byte, size)
	memory.ReadAt(bytes, int64(start))
	memory.Close()

	return bytes, nil
}

func (shmem *HWiNFOShMem) Version() string {
	return fmt.Sprintf("v%d rev%d", shmem.header.Version(), shmem.header.Revision())
}

func (shmem *HWiNFOShMem) Header() *Header {
	return shmem.header
}

func (shmem *HWiNFOShMem) Sensors() []Sensor {
	return shmem.sensors
}

func (shmem *HWiNFOShMem) Readings() []Reading {
	return shmem.readings
}
