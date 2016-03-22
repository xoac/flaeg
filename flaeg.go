package main

import (
	"errors"
	"flag"
	"reflect"
	"strings"
	"time"
)

// GetTypesRecursive links in namesmap a flag with there flildstruct Type
// You can whether provide objValue on a structure or a pointer to structure as first argument
// Flags are genereted from field name or from structags
func getTypesRecursive(objValue reflect.Value, namesmap map[string]reflect.Type, key string) error {
	name := key
	switch objValue.Kind() {
	case reflect.Struct:
		name += objValue.Type().Name()
		for i := 0; i < objValue.NumField(); i++ {
			if tag := objValue.Type().Field(i).Tag.Get("description"); len(tag) > 0 {
				fieldName := objValue.Type().Field(i).Name
				if tag := objValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
					fieldName = tag
				}
				if tag := objValue.Type().Field(i).Tag.Get("short"); len(tag) > 0 {
					if _, ok := namesmap[strings.ToLower(tag)]; ok {
						return errors.New("Tag already exists: " + tag)
					}
					namesmap[strings.ToLower(tag)] = objValue.Field(i).Type()
				}
				if len(key) == 0 {
					name = fieldName
				} else {
					name = key + "." + fieldName
				}
				if _, ok := namesmap[strings.ToLower(name)]; ok {
					return errors.New("Tag already exists: " + name)
				}
				namesmap[strings.ToLower(name)] = objValue.Field(i).Type()
				if err := getTypesRecursive(objValue.Field(i), namesmap, name); err != nil {
					return err
				}
			}
		}
	case reflect.Array, reflect.Map, reflect.Slice, reflect.Ptr:
		typ := objValue.Type().Elem()
		inst := reflect.New(typ).Elem()
		if err := getTypesRecursive(inst, namesmap, name); err != nil {
			return err
		}
	default:
		return nil
	}
	return nil
}

//ParseArgs : parses args into a map[tag]value, using map[type]parser
//args must be formated as like as flag documentation. See https://golang.org/pkg/flag
func parseArgs(args []string, tagsmap map[string]reflect.Type, parsers map[reflect.Type]flag.Value) (map[string]interface{}, error) {
	newParsers := map[string]flag.Value{}
	flagSet := flag.NewFlagSet("flaeg.ParseArgs", flag.ExitOnError)
	valmap := make(map[string]interface{})
	for tag, rType := range tagsmap {
		if parser, ok := parsers[rType]; ok {
			newparser := reflect.New(reflect.TypeOf(parser).Elem()).Interface().(flag.Value)
			flagSet.Var(newparser, tag, "help")
			newParsers[tag] = newparser
		}
	}
	if err := flagSet.Parse(args); err != nil {
		return nil, err
	}
	for tag, newParser := range newParsers {
		valmap[tag] = newParser
	}
	return valmap, nil
}

//FillStructRecursive initialize a value of any taged Struct given by reference
func fillStructRecursive(objValue reflect.Value, valmap map[string]interface{}, key string) error {
	name := key
	// fmt.Printf("objValue begin : %+v\n", objValue)
	switch objValue.Kind() {
	case reflect.Struct:
		name += objValue.Type().Name()
		inst := reflect.New(objValue.Type()).Elem()
		for i := 0; i < inst.NumField(); i++ {
			if tag := objValue.Type().Field(i).Tag.Get("description"); len(tag) > 0 {
				fieldName := objValue.Type().Field(i).Name
				if tag := objValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
					fieldName = tag
				}
				if tag := objValue.Type().Field(i).Tag.Get("short"); len(tag) > 0 {
					if err := setFields(objValue.Field(i), valmap, strings.ToLower(tag)); err != nil {
						return err
					}

				}
				if len(key) == 0 {
					name = fieldName
				} else {
					name = key + "." + fieldName
				}
				// fmt.Printf("tag : %s\n", name)
				if err := setFields(objValue.Field(i), valmap, strings.ToLower(name)); err != nil {
					return err
				}
				if err := fillStructRecursive(objValue.Field(i), valmap, name); err != nil {
					return err
				}
			}
		}
	case reflect.Ptr:
		if objValue.IsNil() {
			inst := reflect.New(objValue.Type().Elem())
			if err := fillStructRecursive(inst.Elem(), valmap, name); err != nil {
				return err
			}
			objValue.Set(inst)
		} else {
			if err := fillStructRecursive(objValue.Elem(), valmap, name); err != nil {
				return err
			}
		}
	default:
		return nil
	}
	// fmt.Printf("objValue end : %+v\n", objValue)
	return nil
}

// SetFields sets value to fieldValue using tag as key in valmap
func setFields(fieldValue reflect.Value, valmap map[string]interface{}, tag string) error {
	if reflect.DeepEqual(fieldValue.Interface(), reflect.New(fieldValue.Type()).Elem().Interface()) {
		if fieldValue.CanSet() {
			if val, ok := valmap[tag]; ok {
				// fmt.Printf("tag %s : set %s in a %s\n", tag, val, fieldValue.Kind())
				fieldValue.Set(reflect.ValueOf(val).Elem().Convert(fieldValue.Type()))
			}
		} else {
			return errors.New(fieldValue.Type().String() + " is not settable.")
		}

	}
	return nil
}

//loadParsers loads default parsers and custom parsers given as parameter. Return a map [reflect.Type]parsers
func loadParsers(customParsers map[reflect.Type]flag.Value) (map[reflect.Type]flag.Value, error) {
	parsers := map[reflect.Type]flag.Value{}
	var stringParser stringValue
	var boolParser boolValue
	var intParser intValue
	var timeParser timeValue
	parsers[reflect.TypeOf("")] = &stringParser
	parsers[reflect.TypeOf(true)] = &boolParser
	parsers[reflect.TypeOf(1)] = &intParser
	parsers[reflect.TypeOf(time.Now())] = &timeParser
	for rType, parser := range customParsers {
		if _, ok := parsers[rType]; ok {
			return parsers, errors.New("Parser already exists on type " + rType.String())
		}
		parsers[rType] = parser
	}
	return parsers, nil
}

// func Load(config interface{}, args []string, customParsers map[reflect.Type]flag.Value) error {

// }
