package core

import (
	"errors"
	"fmt"
)

type RedisCommand struct {
	Cmd  string
	Args []string
}

func DecodeArrayString(data []byte) ([]string, error) {
	value, _, err := Decode(data)
	if err != nil {
		return nil, err
	}
	ts := value.([]interface{})
	tokens := make([]string, len(ts))
	for i := range tokens {
		tokens[i] = ts[i].(string)
	}
	return tokens, nil
}

func Decode(data []byte) (any, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("no data")
	}
	switch data[0] {
	case '+':
		return readSimpleString(data)
	case '-':
		return readError(data)
	case ':':
		return readInt64(data)
	case '$':
		return readBulkString(data)
	case '*':
		return readArray(data)
	}
	return nil, 0, fmt.Errorf("unknown RESP type: %c", data[0])
}

func readSimpleString(data []byte) (string, int, error) {
	pos := 1
	for ; data[pos] != '\r'; pos++ {
	}

	return string(data[1:pos]), pos + 2, nil
}

func readBulkString(data []byte) (string, int, error) {
	pos := 1

	length, delta := readLength(data[1:])
	pos += delta

	return (string(data[pos : pos+length])), pos + length + 2, nil
}

func readInt64(data []byte) (int64, int, error) {
	var value int64 = 0
	pos := 1
	for ; data[pos] != '\r'; pos++ {
		value = value*10 + int64(data[pos]-'0')
	}
	return value, pos + 2, nil
}

func readArray(data []byte) (interface{}, int, error) {
	pos := 1
	count, delta := readLength(data[1:])
	pos += delta
	elements := make([]interface{}, count)
	for i := range count {
		value, n, err := Decode(data[pos:])

		if err != nil {
			return nil, 0, err
		}
		elements[i] = value
		pos += n
	}

	return elements, pos, nil
}

func readError(data []byte) (string, int, error) {
	return readSimpleString(data)
}

func readLength(data []byte) (int, int) {
	pos, length := 0, 0
	for pos = range data {
		b := data[pos]
		if !(b >= '0' && b <= '9') {
			return length, pos + 2
		}
		length = length*10 + int(b-'0')
	}
	return 0, 0
}

func Encode(value interface{}, isSimple bool) []byte {
	switch v := value.(type) {
	case string:
		if isSimple {
			return fmt.Appendf(nil, "+%s\r\n", v)
		}
		return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(v), v)
	}
	return []byte{}
}
