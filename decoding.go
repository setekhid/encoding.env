package env

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"
)

func Unmarshal(e []byte, s interface{}) error {
	return NewDecoder(bytes.NewReader(e)).Decode(s)
}

type Decoder struct {
	lr *bufio.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		lr: bufio.NewReader(r),
	}
}

func (d *Decoder) Decode(obj interface{}) error {
	return d.DecodeWithPrefix("", obj)
}

func (d *Decoder) DecodeWithPrefix(pre string, obj interface{}) error {
	return d.doDecode(pre, obj)
}

func (d *Decoder) doDecode(pre string, obj interface{}) error {

	// map object
	envm, err := MapWithPrefix(pre, obj)
	if err != nil {
		return err
	}

	// iterate env lines
	line, err := d.lr.ReadString('\n')
	for err == nil {

		// string key value
		line = line[:len(line)-1]
		sep_i := strings.Index(line, "=")
		if sep_i < 0 {
			return errors.New("invalid env defination " + line)
		}
		k, v := line[:sep_i], line[sep_i+1:]

		// env reflect value
		env, exists := envm[k]
		if exists {
			err := d.setEnv(env, k, v)
			if err != nil {
				return err
			}
		}
		// TODO support to set array and map element

		line, err = d.lr.ReadString('\n')
	}

	return nil
}

func (d *Decoder) setEnv(env Options, key, val string) error {

	switch env.val.Kind() {

	// boolean
	case reflect.Bool:
		v, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		env.val.SetBool(v)

	// integer
	case reflect.Int,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		env.val.SetInt(v)

	// unsigned integer
	case reflect.Uint,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		env.val.SetUint(v)

	// float
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		env.val.SetFloat(v)

	// string
	case reflect.String:
		env.val.SetString(val)

	// pointer
	case reflect.Ptr:
		if env.val.IsNil() {
			env.val.Set(reflect.New(env.val.Type().Elem()))
		}
		opts := env
		opts.val = env.val.Elem()
		err := d.setEnv(opts, key, val)
		if err != nil {
			return err
		}

	// unassignable
	default:
		return errors.New(
			"unassignable value with type " + env.val.Kind().String() +
				" and key " + key)
	}

	return nil
}
