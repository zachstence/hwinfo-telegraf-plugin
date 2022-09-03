package shmem

// #include "../hwisenssm2.h"
import "C"
import (
	"github.com/hidez8891/shm"

	"hwinfo64-telegraf-plugin/hwinfo/mutex"
)

// Arbitrary values chosen to somehow bound the size of the buffer we are creating
// TODO is there a better way to do this?
const maxSensors = 50
const maxReadings = 500
const headerLength = C.sizeof_HWiNFO_SENSORS_SHARED_MEM2
const sensorsLength = maxSensors * C.sizeof_HWiNFO_SENSORS_SENSOR_ELEMENT
const readingsLength = maxReadings * C.sizeof_HWiNFO_SENSORS_READING_ELEMENT
const totalLength = headerLength + sensorsLength + readingsLength

func Read() ([]byte, error) {
	// Lock mutex and unlock after we are done reading
	err := mutex.Lock()
	defer mutex.Unlock()
	if err != nil {
		return nil, err
	}

	// Open and read shared memory
	r, err := shm.Open(C.HWiNFO_SENSORS_MAP_FILE_NAME2, totalLength)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, totalLength)
	r.Read(buf)
	r.Close()

	return buf, nil
}
