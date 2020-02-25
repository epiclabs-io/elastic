# elastic
Converts go types no matter what

`elastic` is a simple library that converts any type to another the best way possible. This is useful when the type is only known at run-time, which usually happens when serializing data. `elastic` allows your code to be flexible regarding type conversion if that is what you're looking for.

It is also capable of seeing through alias types and converting slices and maps to and from other types of slices and maps, providing there is some logical way to convert them.

Default conversion can be overridden by providing custom conversion functions for specific types.
Struct types can also implement the `ConverterTo` interface to help with conversion to and from specific types.


## Quick examples:

### convert value types:

```go
	// note that using elastic wouldn't make sense if you are certain
    // f is a float64 at compile time.
    var f interface{} = float64(5.5)
	var i int
    
    err := elastic.Set(&i, f)
	if err != nil {
		log.Fatal(f)
	}

	fmt.Println(i) // prints 5
```

### convert slices:

```go
	var ints []int
	err = elastic.Set(&ints, []interface{}{1, 2, 3, "4", float64(5), 6})
	if err != nil {
		log.Fatal(f)
	}

	fmt.Println(ints) // prints [1 2 3 4 5 6]
```

### convert maps:

```go
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
	fmt.Println(intmap) // prints map[1:uno 2:dos 3:tres]
```

# Simple API:

## `elastic.Convert()`
Converts the passed value to the target type
#### Syntax:
`elastic.Convert(source interface{}, targetType reflect.Type) (interface{}, error)`

* `source`: value to convert
* `targetType` the type you want to convert `source` to

#### Returns
The converted value or an error if it fails.

## `elastic.Set()`
Sets the given variable to the passed value
#### Syntax:
`elastic.Set(target, source interface{}) error`
* `target`: value to set. Must be a pointer.
* `source` the value to convert

#### Returns
Only an error if it fails.

# Advanced API:

You can create different instances of the elastic conversion engine so that you can customize conversions independently

## `elastic.New()`
Returns a new conversion engine. It has a `.Set()` and `.Convert()` as above that will work according to the rules set for this engine

## `AddSourceConverter() and AddTargetConverter()`
Registers a conversion function for the given type, either when the type is found on the source side or the target side.

These are useful when you do not control the type you whish to make convertible

#### Syntax:
`engine.AddSourceConverter(sourceType reflect.Type, f ConverterFunc)`
* `sourceType`: type you want to set a custom conversion function for
* `f`: Conversion function to invoke when this type is found as a source

`engine.AddTargetConverter(targetType reflect.Type, f ConverterFunc)`
* `targetType`: type you want to set a custom conversion function for
* `f`: Conversion function to invoke when this type is found as a target

The value returned by your function does not have to be *exactly* of type `targetType`. For example if a `float64` is requested and you return an integer, `elastic` will deal with it.

#### Example:
```go
package main

type Vector struct {
	X float64
	Y float64
}

func main() {

	// Add a custom converter to convert Vector to float64 or int
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

	v := Vector{
		X: 3.0,
		Y: 4.0,
	}

	f, err = elastic.Convert(v, reflect.TypeOf(float64(0)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(f) // prints 5

	var s string
	elastic.Set(&s, v)
	fmt.Println(s) // prints (3, 4)

}

```

## `ConverterTo` interface

```go
type ConverterTo interface {
	ConvertTo(targetType reflect.Type) (interface{}, error)
}
```

Implement this interface in your type to provide conversion to another types. This function will be invoked every time your type is on the right-hand side of a conversion.
The value returned by your function does not have to be *exactly* of type `targetType`. For example if a `float64` is requested and you return an integer, `elastic` will deal with it.

#### Example:

```go
type Vector struct {
	X float64
	Y float64
}

func (v *Vector) ConvertTo(targetType reflect.Type) (interface{}, error) {
		switch targetType.Kind() {
		case reflect.Float64, reflect.Int:
			return math.Sqrt(float64(v.X*v.X) + float64(v.Y*v.Y)), nil
		case reflect.String:
			return fmt.Sprintf("(%g, %g)", v.X, v.Y), nil
		}
		return nil, elastic.ErrNoConversionAvailable
}

func main() {
	v := Vector{
		X: 3.0,
		Y: 4.0,
	}
	
	var i int
	elastic.Set(&i, v)
	fmt.Println(i) // prints 5

	var s string
	elastic.Set(&s, v)
	fmt.Println(s) // prints (3, 4)
}


```