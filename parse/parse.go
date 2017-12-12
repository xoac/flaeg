package parse

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
)

//TODO : add parsers on all types in https://golang.org/pkg/builtin/

// Parser is an interface that allows the contents of a flag.Getter to be set.
type Parser interface {
	flag.Getter
	SetValue(interface{})
}

// -- bool Value
type BoolValue bool

func (b *BoolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*b = BoolValue(v)
	return err
}

func (b *BoolValue) Get() interface{} { return bool(*b) }

func (b *BoolValue) String() string { return fmt.Sprintf("%v", *b) }

func (b *BoolValue) IsBoolFlag() bool { return true }

func (b *BoolValue) SetValue(val interface{}) {
	*b = BoolValue(val.(bool))
}

// optional interface to indicate boolean flags that can be
// supplied without "=value" text
type BoolFlag interface {
	flag.Value
	IsBoolFlag() bool
}

// -- int Value
type IntValue int

func (i *IntValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = IntValue(v)
	return err
}

func (i *IntValue) Get() interface{} { return int(*i) }

func (i *IntValue) String() string { return fmt.Sprintf("%v", *i) }

func (i *IntValue) SetValue(val interface{}) {
	*i = IntValue(val.(int))
}

// -- int64 Value
type Int64Value int64

func (i *Int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = Int64Value(v)
	return err
}

func (i *Int64Value) Get() interface{} { return int64(*i) }

func (i *Int64Value) String() string { return fmt.Sprintf("%v", *i) }

func (i *Int64Value) SetValue(val interface{}) {
	*i = Int64Value(val.(int64))
}

// -- uint Value
type UintValue uint

func (i *UintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*i = UintValue(v)
	return err
}

func (i *UintValue) Get() interface{} { return uint(*i) }

func (i *UintValue) String() string { return fmt.Sprintf("%v", *i) }

func (i *UintValue) SetValue(val interface{}) {
	*i = UintValue(val.(uint))
}

// -- uint64 Value
type Uint64Value uint64

func (i *Uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*i = Uint64Value(v)
	return err
}

func (i *Uint64Value) Get() interface{} { return uint64(*i) }

func (i *Uint64Value) String() string { return fmt.Sprintf("%v", *i) }

func (i *Uint64Value) SetValue(val interface{}) {
	*i = Uint64Value(val.(uint64))
}

// -- string Value
type StringValue string

func (s *StringValue) Set(val string) error {
	*s = StringValue(val)
	return nil
}

func (s *StringValue) Get() interface{} { return string(*s) }

func (s *StringValue) String() string { return fmt.Sprintf("%s", *s) }

func (s *StringValue) SetValue(val interface{}) {
	*s = StringValue(val.(string))
}

// -- float64 Value
type Float64Value float64

func (f *Float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	*f = Float64Value(v)
	return err
}

func (f *Float64Value) Get() interface{} { return float64(*f) }

func (f *Float64Value) String() string { return fmt.Sprintf("%v", *f) }

func (f *Float64Value) SetValue(val interface{}) {
	*f = Float64Value(val.(float64))
}

// Duration is a custom type suitable for parsing duration values.
// It supports `time.ParseDuration`-compatible values and suffix-less digits; in
// the latter case, seconds are assumed.
type Duration time.Duration

// Set sets the duration from the given string value.
func (d *Duration) Set(s string) error {
	if v, err := strconv.Atoi(s); err == nil {
		*d = Duration(time.Duration(v) * time.Second)
		return nil
	}

	v, err := time.ParseDuration(s)
	*d = Duration(v)
	return err
}

// Get returns the duration value.
func (d *Duration) Get() interface{} { return time.Duration(*d) }

// String returns a string representation of the duration value.
func (d *Duration) String() string { return (*time.Duration)(d).String() }

// SetValue sets the duration from the given Duration-asserted value.
func (d *Duration) SetValue(val interface{}) {
	*d = Duration(val.(Duration))
}

// UnmarshalText deserializes the given text into a duration value.
// It is meant to support TOML decoding of durations.
func (d *Duration) UnmarshalText(text []byte) error {
	return d.Set(string(text))
}

// -- time.Time Value
type TimeValue time.Time

func (t *TimeValue) Set(s string) error {
	v, err := time.Parse(time.RFC3339, s)
	*t = TimeValue(v)
	return err
}

func (t *TimeValue) Get() interface{} { return time.Time(*t) }

func (t *TimeValue) String() string { return (*time.Time)(t).String() }

func (t *TimeValue) SetValue(val interface{}) {
	*t = TimeValue(val.(time.Time))
}

//SliceStrings parse slice of strings
type SliceStrings []string

//Set adds strings elem into the the parser
//it splits str on , and ;
func (s *SliceStrings) Set(str string) error {
	fargs := func(c rune) bool {
		return c == ',' || c == ';'
	}
	// get function
	slice := strings.FieldsFunc(str, fargs)
	*s = append(*s, slice...)
	return nil
}

//Get []string
func (s *SliceStrings) Get() interface{} { return []string(*s) }

//String return slice in a string
func (s *SliceStrings) String() string { return fmt.Sprintf("%v", *s) }

//SetValue sets []string into the parser
func (s *SliceStrings) SetValue(val interface{}) {
	*s = SliceStrings(val.([]string))
}
