package main

import (
	"fmt"
	"log"
	"math"
	"reflect"

	"github.com/epiclabs-io/elastic"
)

// Vector is a sample structure that represents a vector
type Vector struct {
	X float64
	Y float64
}

func main() {

	var f interface{} = float64(5.5)
	var i int

	// convert value types
	// note that using convert wouldn't make sense if you are certain
	// f is a float64 at compile time.
	err := elastic.Set(&i, f)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(i)

	var ints []int
	err = elastic.Set(&ints, []interface{}{1, 2, 3, "4", float64(5), 6})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ints)

	someMap := map[string]interface{}{
		"1": "uno",
		"2": "dos",
		"3": "tres",
	}

	intmap := make(map[int]string)
	err = elastic.Set(&intmap, someMap)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(intmap)

	// Add a custom converter to convert Vector to float64
	// (Calculates the modulus of the vector)
	elastic.Default.AddSourceConverter(reflect.TypeOf(Vector{}), func(source interface{}, targetType reflect.Type) (interface{}, error) {
		vector := source.(Vector)
		switch targetType.Kind() {
		case reflect.Float64, reflect.Int:
			return math.Sqrt(float64(vector.X*vector.X) + float64(vector.Y*vector.Y)), nil
		case reflect.String:
			return fmt.Sprintf("(%g, %g)", vector.X, vector.Y), nil
		}
		return nil, elastic.ErrNoConversionAvailable

	})

	// Add a custom converter to convert a string to vector
	elastic.Default.AddSourceConverter(reflect.TypeOf(string("")), func(source interface{}, targetType reflect.Type) (interface{}, error) {
		switch targetType {
		case reflect.TypeOf(Vector{}):
			var v Vector
			_, err := fmt.Sscanf(source.(string), "(%g, %g)", &v.X, &v.Y)
			if err != nil {
				return nil, err
			}
			return v, nil
		}
		return nil, elastic.ErrNoConversionAvailable

	})

	v := Vector{
		X: 3.0,
		Y: 4.0,
	}

	f, err = elastic.Convert(v, reflect.TypeOf(float64(0)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(f) // prints 5

	var n int
	elastic.Set(&n, v)
	fmt.Println(n) // prints 5

	var s string
	elastic.Set(&s, v)
	fmt.Println(s) // prints (3, 4)

	elastic.Set(&v, "(2, 8)")
	fmt.Println(v) // prints {2, 8}

}
