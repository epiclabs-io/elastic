package elastic_test

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/epiclabs-io/elastic"

	"github.com/epiclabs-io/ut"
)

var ErrAny = errors.New("Any error")

type StringAlias string
type FloatAlias float64
type IntAlias int

type TestStruct struct {
	X int
	Y int
}

func (ts *TestStruct) String() string {
	return fmt.Sprintf("(%d, %d)", ts.X, ts.Y)
}

// implement the converter interface
func (ts *TestStruct) ConvertTo(targetType reflect.Type) (interface{}, error) {
	if targetType.Kind() == reflect.Float64 {
		return math.Sqrt(float64(ts.X*ts.X) + float64(ts.Y*ts.Y)), nil
	}
	return nil, elastic.ErrNoConversionAvailable
}

type ConversionTest struct {
	source         interface{}
	expectedResult interface{}
	expectedError  error
}

var testData = []ConversionTest{
	{true, true, nil},
	{"hello", "hello", nil}, // check no conversion when same types
	{int8(1), int8(1), nil},
	{int16(2), int16(2), nil},
	{int32(3), int32(3), nil},
	{int64(4), int64(4), nil},
	{int8(-1), int8(-1), nil},
	{int16(-2), int16(-2), nil},
	{int32(-3), int32(-3), nil},
	{int64(-4), int64(-4), nil},
	{uint8(5), uint8(5), nil},
	{uint16(6), uint16(6), nil},
	{uint32(7), uint32(7), nil},
	{uint64(8), uint64(8), nil},
	{float32(9.2), float32(9.2), nil},
	{float64(194.2), float64(194.2), nil},
	{"1", int8(1), nil},
	{"2", int16(2), nil},
	{"3", int32(3), nil},
	{"4", int64(4), nil},
	{"-1", int8(-1), nil},
	{"-2", int16(-2), nil},
	{"-3", int32(-3), nil},
	{"-4", int64(-4), nil},
	{"5", uint8(5), nil},
	{"6", uint16(6), nil},
	{"7", uint32(7), nil},
	{"8", uint64(8), nil},
	{"9.2", float32(9.2), nil},
	{"19.3", float64(19.3), nil},
	{"-9.2", float32(-9.2), nil},
	{"-19.3", float64(-19.3), nil},
	{int8(1), "1", nil},
	{int16(2), "2", nil},
	{int32(3), "3", nil},
	{int64(4), "4", nil},
	{int8(-1), "-1", nil},
	{int16(-2), "-2", nil},
	{int32(-3), "-3", nil},
	{int64(-4), "-4", nil},
	{uint8(5), "5", nil},
	{uint16(6), "6", nil},
	{uint32(7), "7", nil},
	{uint64(8), "8", nil},
	{float32(9.2), "9.2", nil},
	{float64(19.3), "19.3", nil},
	{float32(-9.2), "-9.2", nil},
	{float64(-19.3), "-19.3", nil},
	{int(-1), uint(0xffffffffffffffff), nil},
	{"true", true, nil},
	{"false", false, nil},
	{true, "true", nil},
	{false, "false", nil},
	{5, float32(5), nil},
	{[]interface{}{1, 2, 3, 4}, []int{1, 2, 3, 4}, nil},
	{[]interface{}{"1", "2", "3", "-4"}, []int{1, 2, 3, -4}, nil},
	{[]interface{}{"1.1", "2.2", "3.3", "-4.4"}, []float32{1.1, 2.2, 3.3, -4.4}, nil},
	{[]interface{}{"1.1", "2.2", "3.3", "-4.4"}, []float64{1.1, 2.2, 3.3, -4.4}, nil},
	{map[string]interface{}{
		"uno":  1,
		"dos":  2,
		"tres": 3,
		"nil": nil,
	}, map[string]int{
		"uno":  1,
		"dos":  2,
		"tres": 3,
		"nil": 0,
	}, nil},
	{map[string]interface{}{
		"1": "uno",
		"2": "dos",
		"3": "tres",
		"4": nil,
	}, map[int]string{
		1: "uno",
		2: "dos",
		3: "tres",
		4: "",
	}, nil},
	{map[string]interface{}{
		"uno":  1,
		"dos":  2,
		"tres": 3,
		"nil": nil,
	}, map[string]int{
		"uno":  1,
		"dos":  2,
		"tres": 3,
		"nil": 0,
	}, nil},
	{[]byte{65, 66, 67, 0}, "ABC\x00", nil},
	{"ABC\x00", []byte{65, 66, 67, 0}, nil},
	{"hola", StringAlias("hola"), nil}, // test conversion between alias
	{StringAlias("hola"), "hola", nil}, // test conversion between alias
	{"XYZ", int8(1), ErrAny},           // test an unparseable string returns an error
	{"XYZ", int16(2), ErrAny},
	{"XYZ", int32(3), ErrAny},
	{"XYZ", int64(4), ErrAny},
	{"XYZ", int8(-1), ErrAny},
	{"XYZ", int16(-2), ErrAny},
	{"XYZ", int32(-3), ErrAny},
	{"XYZ", int64(-4), ErrAny},
	{"XYZ", uint8(5), ErrAny},
	{"XYZ", uint16(6), ErrAny},
	{"XYZ", uint32(7), ErrAny},
	{"XYZ", uint64(8), ErrAny},
	{"XYZ", float32(9.2), ErrAny},
	{"XYZ", float64(19.3), ErrAny},
	{"XYZ", float32(-9.2), ErrAny},
	{"XYZ", float64(-19.3), ErrAny},
	{true, 7, ErrAny},
	{ConversionTest{}, 4, elastic.ErrIncompatibleType},
	{&TestStruct{X: 5, Y: 7}, "(5, 7)", nil},                      // test fmt.Stringer
	{&TestStruct{X: 5, Y: 7}, float64(8.602325267042627), nil},    // Test Converter implementation
	{&TestStruct{X: 5, Y: 7}, FloatAlias(8.602325267042627), nil}, // Test Converter implementation
	{&TestStruct{X: 5, Y: 7}, StringAlias("(5, 7)"), nil},         // test fmt.Stringer implementation to an alias type
	{[]byte{0, 0, 0, 1, 0, 0, 0, 2}, TestStruct{X: 1, Y: 2}, nil}, // Test Target converter
	{&TestStruct{X: 5, Y: 7}, 99, elastic.ErrIncompatibleType},
	{FloatAlias(2.2), int(2), nil},      // test Source converter
	{FloatAlias(2.7), int(3), nil},      // test Source converter
	{FloatAlias(2.7), IntAlias(3), nil}, // test Source converter
	{float32(5.5), float64(5.5), nil},   // test upgrade/downgrade
	{float64(5.5), float32(5.5), nil},   // test upgrade/downgrade
}

func TestConvert(tx *testing.T) {
	t := ut.BeginTest(tx, false) // set to true to generate test results
	defer t.FinishTest()

	// The following adds a custom converter that rounds floats when converting them to integers
	elastic.Default.AddSourceConverter(reflect.TypeOf(FloatAlias(0)), func(source interface{}, targetType reflect.Type) (interface{}, error) {
		if targetType.Kind() == reflect.Int {
			f := source.(FloatAlias)
			d := f - FloatAlias(int(f))
			if d > 0.5 {
				return int(f) + 1, nil
			}
			return int(f), nil
		}
		return nil, elastic.ErrNoConversionAvailable
	})

	// The following adds a custom target converter that unpacks a TestStruct out of a byte array
	elastic.Default.AddTargetConverter(reflect.TypeOf(TestStruct{}), func(source interface{}, targetType reflect.Type) (interface{}, error) {
		switch source.(type) {
		case []byte:
			b := source.([]byte)
			if len(b) == 8 {
				return TestStruct{
					X: int(binary.BigEndian.Uint32(b)),
					Y: int(binary.BigEndian.Uint32(b[4:])),
				}, nil
			}
		}
		return nil, elastic.ErrNoConversionAvailable
	})

	// run all tests
	// each conversion test has a source value that will be converted to the type
	// of the expected result. Then the result is compared to the expected result.
	for _, ct := range testData {
		t.StartSubTest("Conversion of '%v' (%s) to %s", ct.source, reflect.TypeOf(ct.source), reflect.TypeOf(ct.expectedResult).String())
		r, err := elastic.Convert(ct.source, reflect.TypeOf(ct.expectedResult))
		if ct.expectedError == nil {
			t.Ok(err)                      // verify no error
			t.Equals(ct.expectedResult, r) // compare values to see if conversion was correct
		} else {
			if ct.expectedError == ErrAny {
				t.MustFail(err, "Conversion should have failed")
			} else {
				t.MustFailWith(err, ct.expectedError)
			}
		}

		// run the same test but using `elastic.Set` instead
		target := reflect.New(reflect.TypeOf(ct.expectedResult))
		err = elastic.Set(target.Interface(), ct.source)
		if ct.expectedError == nil {
			t.Ok(err)
			t.Equals(ct.expectedResult, target.Elem().Interface())
		} else {
			if ct.expectedError == ErrAny {
				t.MustFail(err, "Conversion should have failed")
			} else {
				t.MustFailWith(err, ct.expectedError)
			}
		}
	}

	// Test `Set` fails when the first parameter is not a pointer
	var x int
	err := elastic.Set(x, 4)
	t.MustFailWith(err, elastic.ErrExpectedPointer)

}
