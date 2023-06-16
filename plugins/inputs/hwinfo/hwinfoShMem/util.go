package hwinfoShMem

import "C"

import (
	"log"
	"strings"
	"unsafe"

	"golang.org/x/text/encoding/charmap"
)

func goStringFromPtr(ptr unsafe.Pointer, len int) string {
	s := C.GoStringN((*C.char)(ptr), C.int(len))
	return s[:strings.IndexByte(s, 0)]
}

// DecodeCharPtr decodes ISO8859_1 string to UTF-8
func DecodeCharPtr(ptr unsafe.Pointer, len int) string {
	s := goStringFromPtr(ptr, len)
	ds, err := decodeISO8859_1(s)
	if err != nil {
		log.Fatalf("TODO: failed to decode: %v", err)
	}
	return ds
}

var isodecoder = charmap.ISO8859_1.NewDecoder()

func decodeISO8859_1(in string) (string, error) {
	return isodecoder.String(in)
}

func StartsWithLower(str string, substr string) bool {
	_str := strings.ToLower(str)
	_substr := strings.ToLower(substr)

	return strings.HasPrefix(_str, _substr)
}
