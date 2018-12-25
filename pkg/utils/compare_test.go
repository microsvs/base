package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenericTypeConvert(t *testing.T) {
	var err error
	// bool
	var src01 bool = true
	var dest01 bool
	if err = GenericTypeConvert(src01, &dest01); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, src01, dest01, "bool is not equal")
	// int
	var src02 int = 10
	var dest02 int
	if err = GenericTypeConvert(src02, &dest02); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, src02, dest02, "int is not equal")
	// string
	var src03 string = "hello,world"
	var dest03 string
	if err = GenericTypeConvert(src03, &dest03); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, src03, dest03, "string is not equal")
	// slice
	var src04 = []int{1, 2, 3}
	var dest04 []int
	if err = GenericTypeConvert(src04, &dest04); err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, src04, dest04, "slice is not equal")
	// map
	var src05 = map[string]interface{}{
		"a": struct {
			Name string
			Age  int
		}{
			Name: "lily",
			Age:  05,
		},
		"b": "tmp",
		"c": 10,
	}
	var dest05 map[string]interface{}
	if err = GenericTypeConvert(src05, &dest05); err != nil {
		t.Error(err.Error())
	}
	fmt.Println(dest05)
}
