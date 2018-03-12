package parse

import (
	"reflect"
	"testing"
	"time"
	"encoding/json"
	"strings"
)

func TestSliceStringsSet(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected SliceStrings
	}{
		{
			desc:     "one value",
			value:    "str",
			expected: SliceStrings{"str"},
		},
		{
			desc:     "two values comma",
			value:    "str1,str2",
			expected: SliceStrings{"str1", "str2"},
		},
		{
			desc:     "two values semicolon",
			value:    "str1;str2",
			expected: SliceStrings{"str1", "str2"},
		},
		{
			desc:     "three values semicolon",
			value:    "str1,str2;str3",
			expected: SliceStrings{"str1", "str2", "str3"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var slice SliceStrings
			if err := slice.Set(test.value); err != nil {
				t.Fatalf("Error :%s", err)
			}

			if !reflect.DeepEqual(slice, test.expected) {
				t.Errorf("Got: %v\nexpected: %s", slice, test.expected)
			}
		})
	}
}

func TestSliceStringsSetAdd(t *testing.T) {
	slice := SliceStrings{"str1"}

	// test
	if err := slice.Set("str2,str3"); err != nil {
		t.Fatalf("Error :%s", err)
	}

	// check
	expected := SliceStrings{"str1", "str2", "str3"}
	if !reflect.DeepEqual(slice, expected) {
		t.Errorf("Expected: %s\ngot: %s", expected, slice)
	}
}

func TestSliceStringsGet(t *testing.T) {
	testCases := []struct {
		desc     string
		values   SliceStrings
		expected []string
	}{
		{
			desc:     "one value",
			values:   SliceStrings{"str1"},
			expected: []string{"str1"},
		},
		{
			desc:     "two values",
			values:   SliceStrings{"str1", "str2"},
			expected: []string{"str1", "str2"},
		},
		{
			desc:     "three values",
			values:   SliceStrings{"str1", "str2", "str3"},
			expected: []string{"str1", "str2", "str3"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if !reflect.DeepEqual(test.values.Get(), test.expected) {
				t.Errorf("Got: %v\nexpected: %s", test.values.Get(), test.expected)
			}
		})
	}
}

func TestSliceStringsString(t *testing.T) {
	testCases := []struct {
		desc     string
		values   SliceStrings
		expected string
	}{
		{
			desc:     "one value",
			values:   SliceStrings{"str"},
			expected: "[str]",
		},
		{
			desc:     "two values",
			values:   SliceStrings{"str1", "str2"},
			expected: "[str1 str2]",
		},
		{
			desc:     "three values",
			values:   SliceStrings{"str1", "str2", "str3"},
			expected: "[str1 str2 str3]",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if !reflect.DeepEqual(test.values.String(), test.expected) {
				t.Errorf("Got: %s\nexpected: %s", test.values.String(), test.expected)
			}
		})
	}
}

func TestSliceStringsSetValue(t *testing.T) {
	testCases := []struct {
		desc     string
		values   []string
		expected SliceStrings
	}{
		{
			desc:     "one value",
			values:   []string{"str"},
			expected: SliceStrings{"str"},
		},
		{
			desc:     "two values",
			values:   []string{"str1", "str2"},
			expected: SliceStrings{"str1", "str2"},
		},
		{
			desc:     "three values",
			values:   []string{"str1", "str2", "str3"},
			expected: SliceStrings{"str1", "str2", "str3"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var slice SliceStrings
			slice.SetValue(test.values)

			if !reflect.DeepEqual(slice, test.expected) {
				t.Errorf("Got: %s\nexpected: %s", slice, test.expected)
			}
		})
	}
}

func TestSetDuration(t *testing.T) {
	tests := []struct {
		in  string
		out time.Duration
	}{
		{
			in:  "42",
			out: 42 * time.Second,
		},
		{
			in:  "42s",
			out: 42 * time.Second,
		},
		{
			in:  "5m",
			out: 5 * time.Minute,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.in, func(t *testing.T) {
			t.Parallel()

			var dur Duration
			if err := dur.Set(test.in); err != nil {
				t.Error(err)
			}

			if time.Duration(dur) != test.out {
				t.Errorf("got %#v, want %#v", time.Duration(dur), test.out)
			}
		})
	}
}

func TestUnmarshalTextDuration(t *testing.T) {
	testCases := []struct {
		desc     string
		value    []byte
		expected time.Duration
	}{
		{
			desc:     "no unit",
			value:    []byte("42"),
			expected: 42 * time.Second,
		},
		{
			desc:     "second",
			value:    []byte("42s"),
			expected: 42 * time.Second,
		},
		{
			desc:     "minute second",
			value:    []byte("4m2s"),
			expected: 242 * time.Second,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var dur Duration
			if err := dur.UnmarshalText(test.value); err != nil {
				t.Error(err)
			}

			if time.Duration(dur) != test.expected {
				t.Errorf("got %#v, want %#v", time.Duration(dur), test.expected)
			}
		})
	}
}

func TestMarshalTextDuration(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected string
	}{
		{
			desc:     "second",
			value:    "42",
			expected: "42s",
		},
		{
			desc:     "hour minute second",
			value:    "3670",
			expected: "1h1m10s",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var dur Duration
			err := dur.Set(test.value)
			if err != nil {
				t.Error(err)
			}

			result, _ := dur.MarshalText()

			if string(result) != test.expected {
				t.Errorf("got %v, want %s", dur, test.expected)
			}
		})
	}
}

func TestUnmarshalJsonDuration(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected time.Duration
	}{
		{
			desc:     "1 second",
			value:    "1000000000",
			expected: time.Duration(1000000000),
		},
		{
			desc:     "with units",
			value:    "\"1m10s\"",
			expected: time.Duration(70000000000),
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var dur Duration
			if err := dur.UnmarshalJSON([]byte(test.value)); err != nil {
				t.Error(err)
			}

			if time.Duration(dur) != test.expected {
				t.Errorf("got %#v, want %#v", time.Duration(dur), test.expected)
			}
		})
	}
}

func TestUnmarshalJsonDurationError(t *testing.T) {
	testCases := []struct {
		desc  string
		value string
	}{
		{
			desc:  "empty",
			value: "",
		},
		{
			desc:  "invalid units",
			value: "1k10s",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var dur Duration
			if err := dur.UnmarshalJSON([]byte(test.value)); err == nil {
				t.Errorf("want error got nil")
			}
		})
	}
}

type Object struct {
	Timeout Duration
}

func TestJsonMarshal(t *testing.T) {
	pointer := &Object{
		Timeout: Duration(666 * time.Second),
	}

	bytes, err := json.Marshal(pointer)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bytes), "666000000000") {
		t.Fatalf("Marshal fail: %s", bytes)
	}

	object := Object{
		Timeout: Duration(666 * time.Second),
	}

	bytes, err = json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bytes), "666000000000") {
		t.Fatalf("Marshal fail: %s", bytes)
	}
}

func TestJsonUnmarshal(t *testing.T) {
	pointer := Object{
	}

	err := json.Unmarshal([]byte(`{"Timeout": "10s"}`), &pointer)
	if err != nil {
		t.Fatal(err)
	}
	if pointer.Timeout != 10000000000 {
		t.Fatalf("Wrong value: %d instead of 10000000000", pointer.Timeout)
	}
}
