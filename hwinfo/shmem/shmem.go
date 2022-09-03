package shmem

/*
#include <windows.h>
#include "../hwisenssm2.h"
*/
import "C"
import (
	"reflect"
	"unsafe"

	"github.com/hidez8891/shm"

	"hwinfo64-telegraf-plugin/hwinfo/mutex"
)

var buf = make([]byte, 200000)

func copyBytes(addr uintptr) []byte {
	headerLen := C.sizeof_HWiNFO_SENSORS_SHARED_MEM2

	var d []byte
	dh := (*reflect.SliceHeader)(unsafe.Pointer(&d))

	dh.Data = addr
	dh.Len, dh.Cap = headerLen, headerLen

	cheader := C.PHWiNFO_SENSORS_SHARED_MEM2(unsafe.Pointer(&d[0]))
	fullLen := int(cheader.dwOffsetOfReadingSection + (cheader.dwSizeOfReadingElement * cheader.dwNumReadingElements))

	if fullLen > cap(buf) {
		buf = append(buf, make([]byte, fullLen-cap(buf))...)
	}

	dh.Len, dh.Cap = fullLen, fullLen

	copy(buf, d)

	return buf[:fullLen]
}

func ReadBytes() ([]byte, error) {
	// Lock mutex and unlock after we are done reading
	err := mutex.Lock()
	defer mutex.Unlock()
	if err != nil {
		return nil, err
	}

	// Open and read shared memory
	r, err := shm.Open(C.HWiNFO_SENSORS_MAP_FILE_NAME2, 200000)
	if err != nil {
		return nil, err
	}

	r.Read(buf)

	r.Close()

	return buf, nil
}
