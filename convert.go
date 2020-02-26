package elastic

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// ConverterFunc is called to override default conversions
type ConverterFunc func(source interface{}, targetType reflect.Type) (interface{}, error)

// ConverterTo interface allows you to define how your type should convert to others
type ConverterTo interface {
	ConvertTo(targetType reflect.Type) (interface{}, error)
}

// ConverterEngine keeps conversion configurations
type ConverterEngine struct {
	sourceConverters    map[reflect.Type][]ConverterFunc
	targetConverters    map[reflect.Type][]ConverterFunc
	interfaceConverters map[reflect.Type][]ConverterFunc
}

// Default is a default conversion engine
var Default = New()

// ErrExpectedPointer is returned when the function expects a pointer parameter
var ErrExpectedPointer = errors.New("Expected pointer")

// ErrIncompatibleType is returned when it is impossible to convert a type to another
var ErrIncompatibleType = errors.New("Incompatible types")

// ErrNoConversionAvailable is returned by any ConverterFunc when it does not know how to convert the passed values
var ErrNoConversionAvailable = errors.New("No conversion available")

// New instantiates a new Converter Engine
func New() *ConverterEngine {
	return &ConverterEngine{
		sourceConverters:    make(map[reflect.Type][]ConverterFunc),
		targetConverters:    make(map[reflect.Type][]ConverterFunc),
		interfaceConverters: make(map[reflect.Type][]ConverterFunc),
	}
}

// AddSourceConverter adds a source conversion function to the engine that knows how to convert the source type to some targets
func (ce *ConverterEngine) AddSourceConverter(sourceType reflect.Type, f ConverterFunc) {
	cf := ce.sourceConverters[sourceType]
	cf = append(cf, f)
	ce.sourceConverters[sourceType] = cf
}

// AddTargetConverter adds a target conversion function to the engine that knows how to convert the target type from some sources
func (ce *ConverterEngine) AddTargetConverter(targetType reflect.Type, f ConverterFunc) {
	cf := ce.targetConverters[targetType]
	cf = append(cf, f)
	ce.targetConverters[targetType] = cf
}

// AddInterfaceConverter adds a converion function for types that match the given interface (experimental)
func (ce *ConverterEngine) AddInterfaceConverter(interfaceType reflect.Type, f ConverterFunc) {
	if interfaceType.Kind() != reflect.Interface {
		panic("type must be an interface")
	}
	cf := ce.interfaceConverters[interfaceType]
	cf = append(cf, f)
	ce.interfaceConverters[interfaceType] = cf
}

// convertMap attempts to convert the source map to another type of map
func (ce *ConverterEngine) convertMap(source interface{}, targetType reflect.Type) (interface{}, error) {
	S := reflect.ValueOf(source)
	T := reflect.MakeMap(targetType)

	targetElementType := targetType.Elem()
	keyType := targetType.Key()

	for i := S.MapRange(); i.Next(); {
		value, err := ce.Convert(i.Value().Interface(), targetElementType)
		if err != nil {
			return nil, err
		}
		key, err := ce.Convert(i.Key().Interface(), keyType)
		if err != nil {
			return nil, err
		}
		T.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	}
	return T.Interface(), nil
}

// convertSlice attempts to convert a slice to another type of slice
func (ce *ConverterEngine) convertSlice(source interface{}, targetType reflect.Type) (interface{}, error) {
	S := reflect.ValueOf(source)
	T := reflect.MakeSlice(targetType, 0, S.Len())
	targetElementType := targetType.Elem()

	for i := 0; i < S.Len(); i++ {
		item, err := ce.Convert(S.Index(i).Interface(), targetElementType)
		if err != nil {
			return nil, err
		}
		T = reflect.Append(T, reflect.ValueOf(item))
	}
	return T.Interface(), nil
}

// kind2Exact converts a type of the same kind
func kind2Exact(source interface{}, targetType reflect.Type) interface{} {
	return reflect.ValueOf(source).Convert(targetType).Interface()
}

// Convert attempts to convert the source value to the given target type
// if it does not fail, the returned value is guaranteed to be of the target type
func (ce *ConverterEngine) Convert(source interface{}, targetType reflect.Type) (interface{}, error) {
	sourceType := reflect.TypeOf(source)
	if sourceType == targetType {
		return source, nil // no conversion necessary
	}

	// check if there are any custom source converters
	converters := ce.sourceConverters[reflect.TypeOf(source)]
	for _, converter := range converters {
		result, err := converter(source, targetType)
		if err == nil {
			return ce.Convert(result, targetType)
		}
		if err != ErrNoConversionAvailable {
			return nil, err
		}
	}

	// check if the source type implements ConverterTo
	converter, ok := source.(ConverterTo)
	if ok {
		result, err := converter.ConvertTo(targetType)
		if err == nil {
			return ce.Convert(result, targetType)
		}
		if err != ErrNoConversionAvailable {
			return nil, err
		}
	}

	// check if there are any custom target converters
	converters = ce.targetConverters[targetType]
	for _, converter := range converters {
		result, err := converter(source, targetType)
		if err == nil {
			return ce.Convert(result, targetType)
		}
		if err != ErrNoConversionAvailable {
			return nil, err
		}
	}

	// check for interface-based converter (experimental)
	for itype, converters := range ce.interfaceConverters {
		for _, converter := range converters {
			if sourceType.Implements(itype) {
				result, err := converter(source, targetType)
				if err == nil {
					return ce.Convert(result, targetType)
				}
				if err != ErrNoConversionAvailable {
					return nil, err
				}
			}
		}
	}

	S := reflect.ValueOf(source)

	// Conversion to string
	if targetType.Kind() == reflect.String {
		stringer, ok := source.(fmt.Stringer) // if target implements Stringer, use it.
		if ok {
			return kind2Exact(stringer.String(), targetType), nil
		}
		// Convert to string typical value types
		switch sourceType.Kind() {
		case reflect.Bool:
			return kind2Exact(strconv.FormatBool(S.Bool()), targetType), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return kind2Exact(strconv.FormatInt(S.Int(), 10), targetType), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return kind2Exact(strconv.FormatUint(S.Uint(), 10), targetType), nil
		case reflect.Float32, reflect.Float64:
			return kind2Exact(strconv.FormatFloat(S.Float(), 'g', 6, int(sourceType.Size())*8), targetType), nil
		}

	}

	if sourceType.Kind() == reflect.String {
		// Attempt to parse typical value types from the string
		switch targetType.Kind() {
		case reflect.Bool:
			b, err := strconv.ParseBool(S.String())
			if err != nil {
				return nil, err
			}
			return kind2Exact(b, targetType), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := strconv.ParseInt(S.String(), 10, int(targetType.Size())*8)
			if err != nil {
				return nil, err
			}
			return kind2Exact(i, targetType), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			i, err := strconv.ParseUint(S.String(), 10, int(targetType.Size())*8)
			if err != nil {
				return nil, err
			}
			return kind2Exact(i, targetType), nil
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(S.String(), int(targetType.Size())*8)
			if err != nil {
				return nil, err
			}
			return kind2Exact(f, targetType), nil
		}
	}

	// slice conversion
	if sourceType.Kind() == reflect.Slice && targetType.Kind() == reflect.Slice {
		return ce.convertSlice(source, targetType)
	}

	// map conversion
	if sourceType.Kind() == reflect.Map && targetType.Kind() == reflect.Map {
		return ce.convertMap(source, targetType)
	}

	// reflection-based conversion
	if reflect.TypeOf(source).ConvertibleTo(targetType) {
		return S.Convert(targetType).Interface(), nil
	}

	// no luck
	return nil, ErrIncompatibleType
}

// Set sets the given target pointer to sourcevalue, performing
// any type conversion necessary
func (ce *ConverterEngine) Set(target, source interface{}) error {
	T := reflect.ValueOf(target)
	if T.Kind() != reflect.Ptr {
		return ErrExpectedPointer
	}
	T = T.Elem()

	converted, err := ce.Convert(source, T.Type())
	if err != nil {
		return err
	}
	T.Set(reflect.ValueOf(converted))
	return nil
}

// Convert attempts to convert the source value to the given target type using the default engine
// if it does not fail, the returned value is guaranteed to be of the target type
func Convert(source interface{}, targetType reflect.Type) (interface{}, error) {
	return Default.Convert(source, targetType)
}

// Set sets the given target pointer to source value using the default engine
// performing any type conversion necessary
func Set(target, source interface{}) error {
	return Default.Set(target, source)
}
