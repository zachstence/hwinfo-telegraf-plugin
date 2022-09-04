package hwinfo

// #include "hwisenssm2.h"
import "C"

import (
	"errors"
	"fmt"
	"log"
	"unsafe"

	"github.com/zachstence/hwinfo-telegraf-plugin/plugins/inputs/hwinfo/internal/mutex"
	"github.com/zachstence/hwinfo-telegraf-plugin/plugins/inputs/hwinfo/internal/shmem"
	"github.com/zachstence/hwinfo-telegraf-plugin/plugins/inputs/hwinfo/internal/util"
)

// SharedMemory provides access to the HWiNFO shared memory
type HWiNFO struct {
	data  []byte
	shmem C.PHWiNFO_SENSORS_SHARED_MEM2
}

// ReadSharedMem reads data from HWiNFO shared memory
// creating a copy of the data
func Read() (*HWiNFO, error) {
	data, err := shmem.Read()
	if err != nil {
		return nil, err
	}

	// If first byte is empty, nothing was read
	if data[0] == 0 {
		return nil, errors.New("no data exists in HWiNFO shared memory, do you have it enabled?")
	}

	shmem := &HWiNFO{
		data:  append([]byte(nil), data...),
		shmem: C.PHWiNFO_SENSORS_SHARED_MEM2(unsafe.Pointer(&data[0])),
	}

	// If we read less than expected, we're missing data
	headerLength := C.sizeof_HWiNFO_SENSORS_SHARED_MEM2
	sensorsLength := shmem.NumSensorElements() * C.sizeof_HWiNFO_SENSORS_SENSOR_ELEMENT
	readingsLength := shmem.NumReadingElements() * C.sizeof_HWiNFO_SENSORS_READING_ELEMENT
	expectedBytes := headerLength + sensorsLength + readingsLength

	actualBytes := len(data)
	if actualBytes < expectedBytes {
		// TODO how to resolve this? Config options?
		return nil, errors.New("didn't read full shared memory buffer")
	}

	return shmem, nil
}

func UnlockMutex() {
	mutex.Unlock()
}

// Signature "HWiS" if active, 'DEAD' when inactive
func (hwinfo *HWiNFO) Signature() string {
	return util.DecodeCharPtr(unsafe.Pointer(&hwinfo.shmem.dwSignature), C.sizeof_DWORD)
}

// Version v1 is latest
func (hwinfo *HWiNFO) Version() int {
	return int(hwinfo.shmem.dwVersion)
}

// Revision revision of version
func (hwinfo *HWiNFO) Revision() int {
	return int(hwinfo.shmem.dwRevision)
}

// PollTime last polling time
func (hwinfo *HWiNFO) PollTime() uint64 {
	addr := unsafe.Pointer(uintptr(unsafe.Pointer(&hwinfo.shmem.dwRevision)) + C.sizeof_DWORD)
	return uint64(*(*C.__time64_t)(addr))
}

// OffsetOfSensorSection offset of the Sensor section from beginning of HWiNFO_SENSORS_SHARED_MEM2
func (hwinfo *HWiNFO) OffsetOfSensorSection() int {
	return int(hwinfo.shmem.dwOffsetOfSensorSection)
}

// SizeOfSensorElement size of each sensor element = sizeof( HWiNFO_SENSORS_SENSOR_ELEMENT )
func (hwinfo *HWiNFO) SizeOfSensorElement() int {
	return int(hwinfo.shmem.dwSizeOfSensorElement)
}

// NumSensorElements number of sensor elements
func (hwinfo *HWiNFO) NumSensorElements() int {
	return int(hwinfo.shmem.dwNumSensorElements)
}

// OffsetOfReadingSection offset of the Reading section from beginning of HWiNFO_SENSORS_SHARED_MEM2
func (hwinfo *HWiNFO) OffsetOfReadingSection() int {
	return int(hwinfo.shmem.dwOffsetOfReadingSection)
}

// SizeOfReadingElement size of each Reading element = sizeof( HWiNFO_SENSORS_READING_ELEMENT )
func (hwinfo *HWiNFO) SizeOfReadingElement() int {
	return int(hwinfo.shmem.dwSizeOfReadingElement)
}

// NumReadingElements number of Reading elements
func (hwinfo *HWiNFO) NumReadingElements() int {
	return int(hwinfo.shmem.dwNumReadingElements)
}

func (hwinfo *HWiNFO) dataForSensor(pos int) ([]byte, error) {
	if pos >= hwinfo.NumSensorElements() {
		return nil, fmt.Errorf("dataForSensor pos out of range, %d for size %d", pos, hwinfo.NumSensorElements())
	}
	start := hwinfo.OffsetOfSensorSection() + (pos * hwinfo.SizeOfSensorElement())
	end := start + hwinfo.SizeOfSensorElement()
	return hwinfo.data[start:end], nil
}

// IterSensors iterate over each sensor
func (hwinfo *HWiNFO) IterSensors() <-chan Sensor {
	ch := make(chan Sensor)
	go func() {
		for i := 0; i < hwinfo.NumSensorElements(); i++ {
			data, err := hwinfo.dataForSensor(i)
			if err != nil {
				log.Fatalf("TODO: failed to read dataForSensor: %v", err)
			}
			ch <- NewSensor(data)
		}
		close(ch)
	}()
	return ch
}

func (hwinfo *HWiNFO) dataForReading(pos int) ([]byte, error) {
	if pos >= hwinfo.NumReadingElements() {
		return nil, fmt.Errorf("dataForReading pos out of range, %d for size %d", pos, hwinfo.NumSensorElements())
	}
	start := hwinfo.OffsetOfReadingSection() + (pos * hwinfo.SizeOfReadingElement())
	end := start + hwinfo.SizeOfReadingElement()
	return hwinfo.data[start:end], nil
}

// IterReadings iterate over each sensor
func (hwinfo *HWiNFO) IterReadings() <-chan Reading {
	ch := make(chan Reading)
	go func() {
		for i := 0; i < hwinfo.NumReadingElements(); i++ {
			data, err := hwinfo.dataForReading(i)
			if err != nil {
				log.Fatalf("TODO: failed to read dataForReading: %v", err)
			}
			ch <- NewReading(data)
		}
		close(ch)
	}()
	return ch
}
