package env

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

func Map(obj interface{}) (map[string]Options, error) {
	m := NewMapping()
	return m, m.Map(obj)
}

func MapWithPrefix(pre string, obj interface{}) (map[string]Options, error) {
	m := NewMapping()
	return m, m.MapWithPrefix(pre, obj)
}

type Mapping map[string]Options

type Options struct {
	val       reflect.Value
	Omitempty bool
}

func DefaultOptions() Options {
	return Options{
		Omitempty: false,
	}
}

func NewMapping() Mapping { return Mapping{} }

func (m Mapping) Map(obj interface{}) error {
	return m.MapWithPrefix("", obj)
}

func (m Mapping) MapWithPrefix(pre string, obj interface{}) error {
	return m.doMap([]byte(pre), reflect.ValueOf(obj), DefaultOptions())
}

func (m Mapping) store(key []byte, val reflect.Value, opts Options) error {

	if _, exists := m[string(key)]; exists {
		return errors.New("duplicated env variable name " + string(key))
	}
	opts.val = val
	m[string(key)] = opts

	return nil
}

func (m Mapping) doMapBuiltin(
	pre []byte, val reflect.Value, opts Options) error {

	return m.store(pre, val, opts)
}

func (m Mapping) doMapArray(pre []byte, val reflect.Value, opts Options) error {

	err := m.store(pre, val, opts)
	if err != nil {
		return err
	}

	if len(pre) > 0 {
		pre = append(pre, '_')
	}

	for i := 0; i < val.Len(); i++ {
		err = m.doMap(
			strconv.AppendInt(pre, int64(i), 10),
			val.Index(i),
			DefaultOptions(),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m Mapping) doMapMap(pre []byte, val reflect.Value, opts Options) error {

	err := m.store(pre, val, opts)
	if err != nil {
		return err
	}

	if len(pre) > 0 {
		pre = append(pre, '_')
	}

	for _, k := range val.MapKeys() {
		err := m.doMap(
			append(pre, []byte(k.String())...),
			val.MapIndex(k),
			DefaultOptions(),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m Mapping) doMapPtr(pre []byte, val reflect.Value, opts Options) error {

	if val.IsNil() {
		return m.store(pre, val, opts)
	}

	return m.doMap(pre, val.Elem(), opts)
}

func (m Mapping) doMapStruct(
	pre []byte, val reflect.Value, opts Options) error {

	if len(pre) > 0 {
		pre = append(pre, '_')
	}

	for i := 0; i < val.NumField(); i++ {

		f := val.Type().Field(i)

		// name and options
		name := f.Name
		omitempty := false

		opts := strings.Split(f.Tag.Get("env"), ",")
		// tag name?
		if len(opts[0]) > 0 {
			name = opts[0]
		}
		opts = opts[1:]
		// other opts?
		for _, opt := range opts {
			// omitempty
			if opt == "omitempty" {
				omitempty = true
			}
		}

		err := m.doMap(
			append(pre, []byte(name)...),
			val.Field(i),
			Options{
				Omitempty: omitempty,
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m Mapping) doMap(pre []byte, val reflect.Value, opts Options) error {

	switch val.Kind() {

	// builtin
	case reflect.Bool,
		reflect.Int,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return m.doMapBuiltin(pre, val, opts)

	// array
	case reflect.Array, reflect.Slice:
		return m.doMapArray(pre, val, opts)

	// map
	case reflect.Map:
		return m.doMapMap(pre, val, opts)

	// pointer
	case reflect.Ptr:
		return m.doMapPtr(pre, val, opts)

	// struct
	case reflect.Struct:
		return m.doMapStruct(pre, val, opts)

	// unknown type
	default:
		return errors.New("unsupported type " + val.Kind().String())
	}
}
