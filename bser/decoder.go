package bser

import (
	"errors"
	"io"
	"unsafe"
)

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (dec *Decoder) Decode() (interface{}, error) {
	peek := make([]uint8, 0, peekBufferSize)

	received, err := dec.r.Read(peek[0:sniffBufferSize])
	if err != nil {
		return nil, err
	} else if received != sniffBufferSize {
		return nil, errors.New("failed to sniff PDU header")
	}

	sizes := []int{0, 0, 0, 1, 2, 4, 8}
	sizesIdx := peek[binaryMarkerSize]
	if sizesIdx < int8Marker || sizesIdx > int64Marker {
		return nil, errors.New("bad PDU size marker")
	}
	sizeMarkerSize := sizes[sizesIdx]

	received, err = dec.r.Read(peek[sniffBufferSize : sniffBufferSize+sizeMarkerSize])
	if err != nil {
		return nil, err
	} else if received != sizeMarkerSize {
		return nil, errors.New("failed to peek at PDU header")
	}

	pduSize, _, err := decodeInt(peek)
	if err != nil {
		return nil, err
	}

	buffer := make([]uint8, pduSize)
	received, err = dec.r.Read(buffer)
	if err != nil {
		return nil, err
	} else if received != int(pduSize) {
		return nil, errors.New("failed to load PDU")
	}

	val, _, err := decodeInterface(buffer)
	if err != nil {
		return nil, err
	}

	return val, nil
}

func decodeInterface(buffer []uint8) (interface{}, []uint8, error) {
	if len(buffer) < int8Size {
		return 0, nil, errors.New("unexpected end of input")
	}

	switch buffer[0] {
	case arrayMarker:
		return decodeArray(buffer)
	case mapMarker:
		return decodeMap(buffer)
	case stringMarker:
		return decodeString(buffer)
	case int8Marker:
		num, buffer, err := decodeInt(buffer)
		return float64(num), buffer, err
	case int16Marker:
		num, buffer, err := decodeInt(buffer)
		return float64(num), buffer, err
	case int32Marker:
		num, buffer, err := decodeInt(buffer)
		return float64(num), buffer, err
	case int64Marker:
		num, buffer, err := decodeInt(buffer)
		return float64(num), buffer, err
	case floatMarker:
		return decodeFloat(buffer)
	case trueVal:
		return true, buffer[1:], nil
	case falseVal:
		return false, buffer[1:], nil
	case nilVal:
		return nil, buffer[1:], nil
	case templateMarker:
		return decodeTemplate(buffer)
	default:
		return nil, nil, errors.New("unsupported type")
	}
}

func decodeString(buffer []uint8) (string, []uint8, error) {
	if len(buffer) < int8Size {
		return "", nil, errors.New("unexpected end of input")
	}

	if buffer[0] != stringMarker {
		return "", nil, errors.New("not a string")
	}

	size, buffer, err := decodeInt(buffer)
	if err != nil {
		return "", nil, err
	}

	if size == 0 {
		return "", buffer, nil
	} else if len(buffer) < int(size) {
		return "", nil, errors.New("insufficient string storage")
	}

	str := string(buffer[0:size])
	return str, buffer[size:], nil
}

func decodeFloat(buffer []uint8) (float64, []uint8, error) {
	bufferLen := len(buffer)
	if bufferLen < int8Size {
		return 0, nil, errors.New("unexpected end of input")
	}

	if buffer[0] != floatMarker {
		return 0, nil, errors.New("not a float64")
	}

	if bufferLen < int8Size+float64Size {
		return 0, nil, errors.New("insufficient float64 storage")
	}

	num := *(*float64)(unsafe.Pointer(&buffer[1]))

	return num, buffer[int8Size+float64Size:], nil
}

func decodeArrayHeader(buffer []uint8) (int64, []uint8, error) {
	if len(buffer) < int8Size {
		return 0, nil, errors.New("unexpected end of input")
	}

	if buffer[0] != arrayMarker {
		return 0, nil, errors.New("not an array")
	}

	size, buffer, err := decodeInt(buffer[1:])
	return size, buffer, err
}

func decodeMap(buffer []uint8) (map[string]interface{}, []uint8, error) {
	if len(buffer) < int8Size {
		return nil, nil, errors.New("unexpected end of input")
	}

	if buffer[0] != mapMarker {
		return nil, nil, errors.New("not a map")
	}

	count, buffer, err := decodeInt(buffer[1:])
	if err != nil {
		return nil, nil, err
	}

	decodedMap := make(map[string]interface{}, 0)
	for i := 0; i < int(count); i++ {
		str, buffer, err := decodeString(buffer)
		if err != nil {
			return nil, nil, err
		}

		val, buffer, err := decodeInterface(buffer)
		if err != nil {
			return nil, nil, err
		}

		decodedMap[str] = val
	}

	return decodedMap, buffer, err
}

func decodeTemplate(buffer []uint8) ([]interface{}, []uint8, error) {
	var err error

	if len(buffer) < int8Size {
		return nil, nil, errors.New("unexpected end of input")
	}

	if buffer[0] != mapMarker {
		return nil, nil, errors.New("not a templated array")
	}

	var headerItemsCount int64
	headerItemsCount, buffer, err = decodeArrayHeader(buffer)
	if err != nil {
		return nil, nil, err
	}

	headers := make([]string, headerItemsCount)
	for i := range headers {
		var str string
		str, buffer, err = decodeString(buffer)
		if err != nil {
			return nil, nil, err
		}
		headers[i] = str
	}

	var rowCount int64
	rowCount, buffer, err = decodeInt(buffer)
	if err != nil {
		return nil, nil, err
	}

	rows := make([]interface{}, rowCount)
	for i := range rows {
		row := make(map[string]interface{})
		for j := 0; j < int(headerItemsCount); j++ {
			if len(buffer) == 0 {
				return nil, nil, errors.New("unexpected end of input")
			}

			if buffer[0] == skipMarker {
				buffer = buffer[1:]
			} else {
				var val interface{}
				val, buffer, err = decodeInterface(buffer)
				if err != nil {
					return nil, nil, err
				}
				key := headers[j]
				row[key] = val
			}
		}
		rows[i] = row
	}

	return rows, buffer, nil
}

func decodeArray(buffer []uint8) ([]interface{}, []uint8, error) {
	size, buffer, err := decodeArrayHeader(buffer)
	if err != nil {
		return nil, nil, err
	}

	buffer = buffer[1:]
	array := make([]interface{}, size)
	for i := range array {
		var val interface{}
		val, buffer, err = decodeInterface(buffer)
		if err != nil {
			return nil, nil, err
		}
		array[i] = val
	}

	return array, buffer, nil
}

func decodeInt(buffer []uint8) (int64, []uint8, error) {
	bufferLen := len(buffer)
	offset := int8Size
	val := int64(0)

	if bufferLen < offset {
		return 0, nil, errors.New("insufficient int storage")
	}

	switch buffer[0] {
	case int8Marker:
		offset += int8Size
		if bufferLen < offset {
			return 0, nil, errors.New("overrun extracting int8")
		}
		num := *(*int8)(unsafe.Pointer(&buffer[1]))
		val = int64(num)
	case int16Marker:
		offset += int16Size
		if bufferLen < offset {
			return 0, nil, errors.New("overrun extracting int16")
		}
		num := *(*int16)(unsafe.Pointer(&buffer[1]))
		val = int64(num)
	case int32Marker:
		offset += int8Size
		if bufferLen < offset {
			return 0, nil, errors.New("overrun extracting int32")
		}
		num := *(*int32)(unsafe.Pointer(&buffer[1]))
		val = int64(num)
	case int64Marker:
		offset += int8Size
		if bufferLen < offset {
			return 0, nil, errors.New("overrun extracting int64")
		}
		num := *(*int64)(unsafe.Pointer(&buffer[1]))
		val = int64(num)
	default:
		return 0, nil, errors.New("bad integer marker")
	}

	return val, buffer[offset:], nil
}
