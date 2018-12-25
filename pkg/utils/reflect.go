package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/microsvs/base/pkg/errors"
)

func FindFieldFromStruct(data interface{}, target string) (interface{}, error) {
	v := reflect.ValueOf(data)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%v is not struct type", data)
	}
	if _, ok := v.Type().FieldByName(target); !ok {
		return nil, fmt.Errorf("field `%s` not exist in %v", target, data)
	}
	return v.FieldByName(target).Interface(), nil
}

// decode callservice return data
// path format: methodname.xxx
func Decode(m interface{}, path string, rawVal interface{}) error {
	var (
		idx   int
		param string
		err   error
		data  = m
		bts   []byte
	)
	params := strings.Split(path, ".")
	for idx, param = range params {
		switch val := data.(type) {
		case map[string]interface{}:
			data = val[param]
		default:
			break
		}
	}
	if idx != len(params)-1 {
		return errors.FGEDataParseError
	}
	if bts, err = json.Marshal(data); err != nil {
		return fmt.Errorf("json decode failed. err=%s", err.Error())
	}
	return json.Unmarshal(bts, rawVal)
}
