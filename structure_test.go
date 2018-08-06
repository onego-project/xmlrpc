package xmlrpc

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

const (
	argsInt        = "records/args_int"
	argsFloat      = "records/args_float"
	argsDouble     = "records/args_double"
	argsBoolt      = "records/args_boolt"
	argsBoolf      = "records/args_boolf"
	argsString     = "records/args_string"
	argsTime       = "records/args_time"
	argsBase64     = "records/args_base64"
	argsSlice      = "records/args_slice"
	argsArray      = "records/args_array"
	argsArrayEmpty = "records/args_array_empty"
	argsStruct     = "records/args_struct"
	argsMap        = "records/args_map"
	argsKind       = "records/args_kind"
)

type food struct{}

func Test_Call_args_int(t *testing.T) {
	// test expects fail before connection to the server, no record needed
	res, err := MakeCallAndCreateRecord(t, argsInt, endpointCorrect, "pow", 2, 9)
	if err != nil {
		t.Fatal("Error:", err)
	}
	if res.ResultInt() != 512 {
		t.Fatal("Method Call returns wrong result.")
	}
}

func Test_Call_args_float(t *testing.T) {
	res, err := MakeCallAndCreateRecord(t, argsFloat, endpointCorrect, "div", float32(90.0), float32(8.0))
	if err != nil {
		t.Fatal("Error:", err)
	}
	if res.ResultDouble() != (90.0 / 8.0) {
		t.Fatal("Method Call returns wrong result.")
	}
}

func Test_Call_args_double(t *testing.T) {
	res, err := MakeCallAndCreateRecord(t, argsDouble, endpointCorrect, "div", 90.0, 8.0)
	if err != nil {
		t.Fatal("Error:", err)
	}
	if res.ResultDouble() != (90.0 / 8.0) {
		t.Fatal("Method Call returns wrong result.")
	}
}

func Test_Call_args_string(t *testing.T) {
	res, err := MakeCallAndCreateRecord(t, argsString, endpointCorrect, "poePoe", "pizza", "pancake")
	if err != nil {
		t.Fatal("Error:", err)
	}
	if res.ResultString() != "yummy" {
		t.Fatal("Method Call returns wrong result.")
	}
}

func Test_Call_args_boolt(t *testing.T) {
	res, err := MakeCallAndCreateRecord(t, argsBoolt, endpointCorrect, "get", true)
	if err != nil {
		t.Fatal("Error:", err)
	}
	if !res.ResultBoolean() {
		t.Fatal("Method Call returns wrong result.")
	}
}

func Test_Call_args_boolf(t *testing.T) {
	res, err := MakeCallAndCreateRecord(t, argsBoolf, endpointCorrect, "get", false)
	if err != nil {
		t.Fatal("Error:", err)
	}
	if res.ResultBoolean() {
		t.Fatal("Method Call returns wrong result.")
	}
}

func Test_Call_args_time(t *testing.T) {
	time, err := time.Parse(timeFormat, "1995-01-01T06:38:05-0000")
	if err != nil {
		t.Fatal("Unable to finish test", err)
	}
	res, err := MakeCallAndCreateRecord(t, argsTime, endpointCorrect, "get", time)
	if err != nil {
		t.Fatal("Error:", err)
	}
	if res.ResultDateTime().String() != time.String() {
		t.Fatal("Method Call returns wrong result.")
	}
}

func Test_Call_args_struct(t *testing.T) {
	res, err := MakeCallAndCreateRecord(t, argsStruct, endpointCorrect, "get", food{})
	if err == nil {
		t.Fatal("No error when arg is struct different from time")
	}
	if res != nil {
		t.Fatal("Method Call returns result when arg is struct differect from time..")
	}
}

func Test_Call_args_base64(t *testing.T) {
	base := []byte("I love pancake.")
	res, err := MakeCallAndCreateRecord(t, argsBase64, endpointCorrect, "get", base)
	if err != nil {
		t.Fatal("Error:", err)
	}
	if !bytes.Equal(res.ResultBase64(), base) {
		t.Fatal("Method Call returns wrong result.")
	}
}

func Test_Call_args_slice(t *testing.T) {
	array := []int64{1, 2, 3}
	res, err := MakeCallAndCreateRecord(t, argsSlice, endpointCorrect, "get", array)
	if err != nil {
		t.Fatal("Error:", err)
	}
	if len(res.ResultArray()) != len(array) {
		t.Fatal("Method Call returns wrong result.")
	}
	for i, e := range array {
		if e != res.ResultArray()[i].ResultInt() {
			t.Fatal("Method Call returns wrong result at index", i, "expected:", e, "got", res.ResultArray()[i].ResultInt())
		}
	}
}

func Test_Call_args_array(t *testing.T) {
	array := [3]int64{1, 2, 3}
	res, err := MakeCallAndCreateRecord(t, argsArray, endpointCorrect, "get", array)
	if err != nil {
		t.Fatal("Error:", err)
	}
	if len(res.ResultArray()) != len(array) {
		t.Fatal("Method Call returns wrong result.")
	}
	for i, e := range array {
		if e != res.ResultArray()[i].ResultInt() {
			t.Fatal("Method Call returns wrong result at index", i, "expected:", e, "got", res.ResultArray()[i].ResultInt())
		}
	}
}

func Test_Call_args_arrayEmpty(t *testing.T) {
	var array []int
	res, err := MakeCallAndCreateRecord(t, argsArrayEmpty, endpointCorrect, "get", array)
	if err == nil {
		t.Fatal("No error when array is empty")
	}
	if !strings.Contains(err.Error(), "cannot parse XML RPC response") {
		t.Fatal("Unexpected error:", err)
	}
	if res != nil {
		t.Fatal("Method Call returns result when array is empty.")
	}
}

func Test_Call_args_arrayNil(t *testing.T) {
	// test expects fail before connection to the server, no record needed
	array := []food{{}}
	res, err := MakeCallAndCreateRecord(t, "", endpointCorrect, "get", array)
	if err == nil {
		t.Fatal("No error when array contains nil")
	}
	if !strings.Contains(err.Error(), "payload preparation failed") {
		t.Fatal("Unexpected error:", err)
	}
	if res != nil {
		t.Fatal("Method Call returns result when array contains nil.")
	}
}

func Test_Call_args_map(t *testing.T) {
	foodValue := map[string]int64{
		"donut": 10,
		"steak": 100,
	}
	res, err := MakeCallAndCreateRecord(t, argsMap, endpointCorrect, "get", foodValue)
	if err != nil {
		t.Fatal("Error:", err)
	}
	for k, v := range foodValue {
		if res.ResultStruct()[k] == nil {
			t.Fatal("Method Call returns wrong result.")
		} else {
			if res.ResultStruct()[k].ResultInt() != v {
				t.Fatal("Method Call returns wrong result.")
			}
		}
	}
}

func Test_Call_args_invalidMap(t *testing.T) {
	// test expects fail before connection to the server, no record needed
	foodValue := map[int64]int64{
		5:  10,
		42: 100,
	}
	res, err := MakeCallAndCreateRecord(t, "", endpointCorrect, "get", foodValue)
	if err == nil {
		t.Fatal("No error when map hasn't key type string.")
	}
	if !strings.Contains(err.Error(), "payload preparation failed") {
		t.Fatal("Unexpected error:", err)
	}
	if res != nil {
		t.Fatal("Method Call returns result when map hasn't key type string.")
	}
}

func Test_Call_args_mapNil(t *testing.T) {
	// test expects fail before connection to the server, no record needed
	foodValue := map[string]food{
		"ananas": {},
	}
	res, err := MakeCallAndCreateRecord(t, "", endpointCorrect, "get", foodValue)
	if err == nil {
		t.Fatal("No error when map's value is nil.")
	}
	if !strings.Contains(err.Error(), "payload preparation failed") {
		t.Fatal("Unexpected error:", err)
	}
	if res != nil {
		t.Fatal("Method Call returns result when map's value is nil.")
	}
}

func Test_Call_args_kind(t *testing.T) {
	res, err := MakeCallAndCreateRecord(t, argsKind, endpointCorrect, "get", 5)
	if err != nil {
		t.Fatal("Error:", err)
	}
	if res.Kind() != KindInt {
		t.Fatal("Method Call returns wrong result.")
	}
}
