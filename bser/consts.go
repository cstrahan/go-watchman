package bser

import "unsafe"

const (
	binaryMarker = "\x00\x01"

	arrayMarker    = 0x00
	mapMarker      = 0x01
	stringMarker   = 0x02
	int8Marker     = 0x03
	int16Marker    = 0x04
	int32Marker    = 0x05
	int64Marker    = 0x06
	floatMarker    = 0x07
	trueVal        = 0x08
	falseVal       = 0x09
	nilVal         = 0x0a
	templateMarker = 0x0b
	skipMarker     = 0x0c

	int8Size    = int(unsafe.Sizeof(int8(0)))
	int16Size   = int(unsafe.Sizeof(int16(0)))
	int32Size   = int(unsafe.Sizeof(int32(0)))
	int64Size   = int(unsafe.Sizeof(int64(0)))
	float64Size = int(unsafe.Sizeof(float64(0)))

	binaryMarkerSize = int8Size * 2
	sniffBufferSize  = binaryMarkerSize + int8Size
	peekBufferSize   = sniffBufferSize + int64Size
)
