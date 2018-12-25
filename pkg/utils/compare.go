package utils

import (
	"encoding/json"
	"reflect"
)

func GenericTypeConvert(src interface{}, dest interface{}) error {
	bts, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(bts, dest)
}

func CompareAnyValues(cmp1 interface{}, cmp2 interface{}) bool {
	aValue := reflect.ValueOf(cmp1)
	bValue := reflect.ValueOf(cmp2)
	// Convert types and compare
	if bValue.Type().ConvertibleTo(aValue.Type()) {
		return reflect.DeepEqual(cmp1, bValue.Convert(aValue.Type()).Interface())
	}
	return false
}
