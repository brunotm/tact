package js

import (
	"encoding/json"
	"strconv"
	"time"
	"unsafe"

	"github.com/buger/jsonparser"
)

var (
	nullValue = []byte(`null`)
)

// GetValue fetches the value under the given path as an interface
func GetValue(data []byte, path ...string) (value interface{}, err error) {
	buf, valueType, _, err := jsonparser.Get(data, path...)

	if err != nil {
		return nil, err
	}

	switch valueType {
	case jsonparser.Null:
		value = nil
	case jsonparser.Number:
		if value, err = jsonparser.ParseFloat(buf); err != nil {
			value, err = jsonparser.ParseInt(buf)
		}
	case jsonparser.String:
		value, err = jsonparser.ParseString(buf)
	case jsonparser.Boolean:
		value, err = jsonparser.ParseBoolean(buf)

	default:
		value = buf
	}
	return value, err
}

// Has checks the existence of the given path
func Has(data []byte, path ...string) (exists bool) {
	_, _, _, err := jsonparser.Get(data, path...)
	return err != jsonparser.KeyPathNotFoundError
}

// Get the []byte for the given path
func Get(data []byte, path ...string) (value []byte, err error) {
	value, _, _, err = jsonparser.Get(data, path...)
	return value, err
}

// GetFloat fetches the value for the given path as a float64
func GetFloat(data []byte, path ...string) (value float64, err error) {
	return jsonparser.GetFloat(data, path...)
}

// GetInt fetches the value for the given path as an int64
func GetInt(data []byte, path ...string) (value int64, err error) {
	return jsonparser.GetInt(data, path...)
}

// GetString fetches the value for the given path as a string
func GetString(data []byte, path ...string) (value string, err error) {
	return jsonparser.GetString(data, path...)
}

// GetUnsafeString is like GetString, but unsafe
func GetUnsafeString(data []byte, path ...string) (value string, err error) {
	return jsonparser.GetUnsafeString(data, path...)
}

// GetBoolean fetches the value for the given path as a bool
func GetBoolean(data []byte, path ...string) (value bool, err error) {
	return jsonparser.GetBoolean(data, path...)
}

// GetTime fetches the value for the given path as a time.Time
func GetTime(data []byte, path ...string) (value time.Time, err error) {
	v, err := Get(data, path...)
	if err != nil {
		return value, err
	}

	err = value.UnmarshalText(v)
	if err != nil {
		return value, err
	}

	return value, nil
}

// Delete the given path
func Delete(data []byte, path ...string) (newData []byte) {
	return jsonparser.Delete(data, path...)
}

// SetRawBytes for the given path
func SetRawBytes(data []byte, value []byte, path ...string) (newData []byte, err error) {
	if len(data) == 0 {
		data = make([]byte, 0, 64)
		data = append(data, '{', '}')
	}
	return jsonparser.Set(data, value, path...)
}

// Set marshal and sets the value for the given path
func Set(data []byte, value interface{}, path ...string) (newData []byte, err error) {
	buf := make([]byte, 0, 16)

	switch v := value.(type) {
	case nil:
		buf = nullValue
	case bool:
		buf = strconv.AppendBool(buf, v)
	case int:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int8:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int16:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int32:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int64:
		buf = strconv.AppendInt(buf, v, 10)
	case uint8:
		buf = strconv.AppendUint(buf, uint64(v), 10)
	case uint16:
		buf = strconv.AppendUint(buf, uint64(v), 10)
	case uint32:
		buf = strconv.AppendUint(buf, uint64(v), 10)
	case uint64:
		buf = strconv.AppendUint(buf, v, 10)
	case float32:
		buf = strconv.AppendFloat(buf, float64(v), 'f', -1, 64)
	case float64:
		buf = strconv.AppendFloat(buf, v, 'f', -1, 64)
	case []byte:
		if buf, err = json.Marshal(*(*string)(unsafe.Pointer(&v))); err != nil {
			return nil, err
		}
	case time.Time:
		if buf, err = v.MarshalJSON(); err != nil {
			return nil, err
		}
	// Also catch strings for escaping
	default:
		if buf, err = json.Marshal(v); err != nil {
			return nil, err
		}
	}
	return SetRawBytes(data, buf, path...)
}

// ForEach iterates over the key-value pairs of the JSON object, invoking a given callback for each such entry
func ForEach(data []byte, cb func(key string, value []byte) error, path ...string) (err error) {
	iter := func(key []byte, value []byte, tp jsonparser.ValueType, offset int) error {
		return cb(*(*string)(unsafe.Pointer(&key)), value)
	}
	return jsonparser.ObjectEach(data, iter, path...)
}

// // ArrayEach is used when iterating arrays, accepts a callback function with the same return arguments as `Get`.
// func (j *JSON) ArrayEach(cb func(value []byte, err error), path ...string) (err error) {
// 	iter := func(value []byte, tp jsonparser.ValueType, offset int, err error) {
// 		cb(value, err)
// 	}
// 	_, err = jsonparser.ArrayEach(j.data, iter, path...)
// 	return err
// }
