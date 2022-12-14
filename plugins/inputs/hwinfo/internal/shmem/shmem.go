package shmem

// #include "../hwisenssm2.h"
import "C"
import (
	"fmt"

	"github.com/hidez8891/shm"
	"github.com/rs/zerolog/log"

	"github.com/zachstence/hwinfo-telegraf-plugin/plugins/inputs/hwinfo/internal/mutex"
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
	mutex.Lock()
	defer mutex.Unlock()

	// Open and read shared memory
	r, err := shm.Open(C.HWiNFO_SENSORS_MAP_FILE_NAME2, totalLength)
	if err != nil {
		if isAccessDeniedErr(err) {
			log.Fatal().Err(err).Msg("could not access HWiNFO shared memory, is this plugin running as Administrator?")
		}
		return nil, err
	}
	buf := make([]byte, totalLength)
	r.Read(buf)
	r.Close()

	return buf, nil
}

func isAccessDeniedErr(err error) bool {
	errStr := fmt.Sprintf("%v", err)
	return errStr == "CreateFileMapping: Access is denied."
}
