package bser

import (
	"errors"
	"unsafe"
)

func Encode(o interface{}) ([]uint8, error) {
	buffer := []uint8{0, 1, int64Marker, 0, 0, 0, 0, 0, 0, 0, 0}
	buffer, err := encodeInterface(o, buffer)
	pduSize := int64(len(buffer)) - 11 // header
	*(*int64)(unsafe.Pointer(&buffer[3])) = pduSize

	return buffer, err
}

func encodeInterface(o interface{}, buffer []uint8) ([]uint8, error) {
	var err error
	switch obj := o.(type) {
	case int:
		buffer, err = encodeInt(int64(obj), buffer)
	case float64:
		buffer, err = encodeFloat(obj, buffer)
	case map[string]interface{}:
		buffer, err = encodeMap(obj, buffer)
	case []interface{}:
		buffer, err = encodeArray(obj, buffer)
	case string:
		buffer, err = encodeString(obj, buffer)
	default:
		buffer, err = nil, errors.New("unsupported type")
	}

	return buffer, err
}

func encodeString(str string, buffer []uint8) ([]uint8, error) {
	buffer = append(buffer, []uint8{stringMarker}...)
	buffer, _ = encodeInt(int64(len(str)), buffer)
	return append(buffer, str...), nil
}

func encodeInt(i int64, buffer []uint8) ([]uint8, error) {
	tmp := []uint8{int64Marker, 0, 0, 0, 0, 0, 0, 0, 0}
	*(*int64)(unsafe.Pointer(&tmp[1])) = i
	buffer = append(buffer, tmp...)
	return buffer, nil
}

func encodeFloat(f float64, buffer []uint8) ([]uint8, error) {
	tmp := []uint8{floatMarker, 0, 0, 0, 0, 0, 0, 0, 0}
	*(*float64)(unsafe.Pointer(&tmp[1])) = f
	buffer = append(buffer, tmp...)
	return buffer, nil
}

func encodeArray(array []interface{}, buffer []uint8) ([]uint8, error) {
	buffer = append(buffer, []uint8{arrayMarker}...)
	buffer, err := encodeInt(int64(len(array)), buffer)
	for _, v := range array {
		buffer, err = encodeInterface(v, buffer)
		if err != nil {
			return nil, err
		}
	}

	return buffer, nil
}

func encodeMap(m map[string]interface{}, buffer []uint8) ([]uint8, error) {
	buffer = append(buffer, []uint8{mapMarker}...)
	buffer, err := encodeInt(int64(len(m)), buffer)
	for k, v := range m {
		buffer, err = encodeString(k, buffer)
		buffer, err = encodeInterface(v, buffer)
		if err != nil {
			return nil, err
		}
	}

	return buffer, nil
}
