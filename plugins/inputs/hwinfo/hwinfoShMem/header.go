package hwinfoShMem

// #include "hwisenssm2.h"
import "C"

import (
	"unsafe"
)

const headerLength = C.sizeof_HWiNFO_SENSORS_SHARED_MEM2

type Header struct {
	c C.PHWiNFO_SENSORS_SHARED_MEM2
}

func NewHeader(data []byte) Header {
	return Header{
		c: C.PHWiNFO_SENSORS_SHARED_MEM2(unsafe.Pointer(&data[0])),
	}
}

// Signature "HWiS" if active, 'DEAD' when inactive
func (header *Header) Signature() string {
	return DecodeCharPtr(unsafe.Pointer(&header.c.dwSignature), C.sizeof_DWORD)
}

// Version version of shared memory
func (header *Header) Version() int {
	return int(header.c.dwVersion)
}

// Revision revision of version
func (header *Header) Revision() int {
	return int(header.c.dwRevision)
}

// PollTime last polling time
func (header *Header) PollTime() uint64 {
	addr := unsafe.Pointer(uintptr(unsafe.Pointer(&header.c.dwRevision)) + C.sizeof_DWORD)
	return uint64(*(*C.__time64_t)(addr))
}

// OffsetOfSensorSection offset of the Sensor section from beginning of HWiNFO_SENSORS_SHARED_MEM2
func (header *Header) OffsetOfSensorSection() int {
	return int(header.c.dwOffsetOfSensorSection)
}

// SizeOfSensorElement size of each sensor element = sizeof( HWiNFO_SENSORS_SENSOR_ELEMENT )
func (header *Header) SizeOfSensorElement() int {
	return int(header.c.dwSizeOfSensorElement)
}

// NumSensorElements number of sensor elements
func (header *Header) NumSensorElements() int {
	return int(header.c.dwNumSensorElements)
}

// OffsetOfReadingSection offset of the Reading section from beginning of HWiNFO_SENSORS_SHARED_MEM2
func (header *Header) OffsetOfReadingSection() int {
	return int(header.c.dwOffsetOfReadingSection)
}

// SizeOfReadingElement size of each Reading element = sizeof( HWiNFO_SENSORS_READING_ELEMENT )
func (header *Header) SizeOfReadingElement() int {
	return int(header.c.dwSizeOfReadingElement)
}

// NumReadingElements number of Reading elements
func (header *Header) NumReadingElements() int {
	return int(header.c.dwNumReadingElements)
}
