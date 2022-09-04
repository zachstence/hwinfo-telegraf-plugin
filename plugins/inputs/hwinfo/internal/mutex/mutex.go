package mutex

// #include "../hwisenssm2.h"
import "C"
import (
	"errors"
	"fmt"
	"sync"
	"unsafe"

	log "github.com/rs/zerolog/log"
)

var ghnd C.HANDLE
var imut = sync.Mutex{}

// Lock the global mutex
func Lock() {
	imut.Lock()
	lpName := C.CString(C.HWiNFO_SENSORS_SM2_MUTEX)
	defer C.free(unsafe.Pointer(lpName))

	ghnd = C.OpenMutex(C.READ_CONTROL, C.FALSE, lpName)
	if ghnd == C.HANDLE(C.NULL) {
		err := handleLastError(uint64(C.GetLastError()))
		log.Fatal().Err(err).Send()
	}
}

// Unlock the global mutex
func Unlock() {
	defer imut.Unlock()
	C.CloseHandle(ghnd)
}

// ErrFileNotFound Windows error
var ErrFileNotFound = errors.New("could not find HWiNFO shared memory file, is HWiNFO running with Shared Memory Support enabled?")

// ErrInvalidHandle Windows error
var ErrInvalidHandle = errors.New("could not read HWiNFO shared memory file, is HWiNFO running with Shared Memory Support enabled?")

// UnknownError unhandled Windows error
type UnknownError struct {
	Code uint64
}

func (e UnknownError) Error() string {
	return fmt.Sprintf("unknown error code: %d", e.Code)
}

// HandleLastError converts C.GetLastError() to golang error
func handleLastError(code uint64) error {
	switch code {
	case 2: // ERROR_FILE_NOT_FOUND
		return ErrFileNotFound
	case 6: // ERROR_INVALID_HANDLE
		return ErrInvalidHandle
	default:
		return UnknownError{Code: code}
	}
}
