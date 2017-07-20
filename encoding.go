package env

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

func Marshal(s interface{}) ([]byte, error) {
	buff := &bytes.Buffer{}
	err := NewEncoder(buff).Encode(s)
	return buff.Bytes(), err
}

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

func (e *Encoder) Encode(obj interface{}) error {
	return e.EncodeWithPrefix("", obj)
}

func (e *Encoder) EncodeWithPrefix(pre string, obj interface{}) error {
	return e.doEncode(pre, obj)
}

// Deprecated
func (e *Encoder) encodeWithPrefix(pre []byte, v reflect.Value) error {

	switch v.Kind() {

	// boolean
	case reflect.Bool:
		env := strconv.AppendBool(append(pre, '='), v.Bool())
		_, err := e.w.Write(append(env, '\n'))
		if err != nil {
			return err
		}

	// int
	case reflect.Int,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		env := strconv.AppendInt(append(pre, '='), v.Int(), 10)
		_, err := e.w.Write(append(env, '\n'))
		if err != nil {
			return err
		}

	// uint
	case reflect.Uint,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		env := strconv.AppendUint(append(pre, '='), v.Uint(), 10)
		_, err := e.w.Write(append(env, '\n'))
		if err != nil {
			return err
		}

	// float
	case reflect.Float32, reflect.Float64:
		env := strconv.AppendFloat(append(pre, '='), v.Float(), 'f', -1, 64)
		_, err := e.w.Write(append(env, '\n'))
		if err != nil {
			return err
		}

	// array
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			err := e.encodeWithPrefix(
				strconv.AppendInt(append(pre, '_'), int64(i), 10),
				v.Index(i),
			)
			if err != nil {
				return err
			}
		}

	// map
	case reflect.Map:
		for _, k := range v.MapKeys() {
			err := e.encodeWithPrefix(
				append(append(pre, '_'), []byte(k.String())...),
				v.MapIndex(k),
			)
			if err != nil {
				return err
			}
		}

	// pointer
	case reflect.Ptr:
		if v.IsNil() {
			env := append(pre, '=')
			_, err := e.w.Write(append(env, '\n'))
			if err != nil {
				return err
			}
		}
		err := e.encodeWithPrefix(pre, v.Elem())
		if err != nil {
			return err
		}

	// string
	case reflect.String:
		env := append(append(pre, '='), []byte(v.String())...)
		_, err := e.w.Write(append(env, '\n'))
		if err != nil {
			return err
		}

	// struct
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {

			f := v.Type().Field(i)

			// name and options
			name := f.Name
			omitempty := false

			opts := strings.Split(f.Tag.Get("env"), ",")
			// tag name?
			if len(opts) > 0 {
				name = opts[0]
			}
			// other opts?
			opts = opts[1:]
			for _, opt := range opts {
				// omitempty
				if opt == "omitempty" {
					omitempty = true
				}
			}

			fv := v.Field(i)
			if !omitempty || !reflect.DeepEqual(reflect.Zero(f.Type), fv) {
				err := e.encodeWithPrefix(
					append(append(pre, '_'), []byte(name)...),
					fv,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (e *Encoder) doEncode(pre string, obj interface{}) error {

	envm, err := MapWithPrefix(pre, obj)
	if err != nil {
		return err
	}

	for k, v := range envm {

		// omitempty
		if v.Omitempty &&
			reflect.DeepEqual(
				reflect.Zero(v.val.Type()).Interface(),
				v.val.Interface()) {
			continue
		}

		switch v.val.Kind() {

		// builtin
		case reflect.Bool,
			reflect.Int,
			reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint,
			reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.String:

			err := e.writeEnv(k, v.val.Interface())
			if err != nil {
				return err
			}

		// pointer
		case reflect.Ptr:
			err := e.writeEnv(k, "")
			if err != nil {
				return err
			}

		// ignore
		default:
		}
	}

	return nil
}

func (e *Encoder) writeEnv(k string, v interface{}) error {
	env := fmt.Sprintf("%v=%v\n", k, v)
	_, err := e.w.Write([]byte(env))
	return err
}
