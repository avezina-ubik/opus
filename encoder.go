// Copyright © 2015 Hraban Luyat <hraban@0brg.net>
//
// License for use of this code is detailed in the LICENSE file

package opus

import (
	"fmt"
	"unsafe"
)

/*
#cgo CFLAGS: -std=c99 -Wall -Werror -pedantic -Ibuild/include
#include <opus/opus.h>
*/
import "C"

var errEncUninitialized = fmt.Errorf("opus encoder uninitialized")

// Encoder contains the state of an Opus encoder for libopus.
type Encoder struct {
	p *C.struct_OpusEncoder
	// Memory for the encoder struct allocated on the Go heap to allow Go GC to
	// manage it (and obviate need to free())
	mem []byte
}

// NewEncoder allocates a new Opus encoder and initializes it with the
// appropriate parameters. All related memory is managed by the Go GC.
func NewEncoder(sample_rate int, channels int, application Application) (*Encoder, error) {
	var enc Encoder
	err := enc.Init(sample_rate, channels, application)
	if err != nil {
		return nil, err
	}
	return &enc, nil
}

// Init initializes a pre-allocated opus encoder. Must be called exactly once in
// the life-time of this object, before calling any other methods.
func (enc *Encoder) Init(sample_rate int, channels int, application Application) error {
	if enc.p != nil {
		return fmt.Errorf("opus encoder already initialized")
	}
	if channels != 1 && channels != 2 {
		return fmt.Errorf("Number of channels must be 1 or 2: %d", channels)
	}
	size := C.opus_encoder_get_size(C.int(channels))
	enc.mem = make([]byte, size)
	enc.p = (*C.OpusEncoder)(unsafe.Pointer(&enc.mem[0]))
	errno := int(C.opus_encoder_init(
		enc.p,
		C.opus_int32(sample_rate),
		C.int(channels),
		C.int(application)))
	if errno != 0 {
		return opuserr(int(errno))
	}
	return nil
}

func (enc *Encoder) Encode(pcm []int16) ([]byte, error) {
	if enc.p == nil {
		return nil, errEncUninitialized
	}
	if pcm == nil || len(pcm) == 0 {
		return nil, fmt.Errorf("opus: no data supplied")
	}
	data := make([]byte, maxEncodedFrameSize)
	n := int(C.opus_encode(
		enc.p,
		(*C.opus_int16)(&pcm[0]),
		C.int(len(pcm)),
		(*C.uchar)(&data[0]),
		C.opus_int32(cap(data))))
	if n < 0 {
		return nil, opuserr(n)
	}
	return data[:n], nil
}

func (enc *Encoder) EncodeFloat32(pcm []float32) ([]byte, error) {
	if enc.p == nil {
		return nil, errEncUninitialized
	}
	if pcm == nil || len(pcm) == 0 {
		return nil, fmt.Errorf("opus: no data supplied")
	}
	data := make([]byte, maxEncodedFrameSize)
	n := int(C.opus_encode_float(
		enc.p,
		(*C.float)(&pcm[0]),
		C.int(len(pcm)),
		(*C.uchar)(&data[0]),
		C.opus_int32(cap(data))))
	if n < 0 {
		return nil, opuserr(n)
	}
	return data[:n], nil
}