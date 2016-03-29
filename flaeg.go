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
func parseArgs(args []string, tagsmap map[string]reflect.Type, defaultValmap map[string]flag.Value, parsers map[reflect.Type]flag.Value) (map[string]flag.Value, error) {
	newParsers := map[string]flag.Value{}
	flagSet := flag.NewFlagSet("flaeg.ParseArgs", flag.ExitOnError)
	valmap := make(map[string]flag.Value)
	for tag, rType := range tagsmap {

		if parser, ok := parsers[rType]; ok {
			newparser := reflect.New(reflect.TypeOf(parser).Elem()).Interface().(flag.Value)
			flagSet.Var(newparser, tag, "help")
			// if _, ok := defaultValmap[tag]; !ok {
			// 	return nil, errors.New("default value doesn't exit for flag " + tag)
			// }
			// if err := flagSet.Set(tag, defaultValmap[tag].String()); err != nil {
			// 	return nil, err
			// }
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
func fillStructRecursive(objValue reflect.Value, valmap map[string]flag.Value, key string) error {
	name := key
	// fmt.Printf("objValue begin : %+v\n", objValue)
	switch objValue.Kind() {
	case reflect.Struct:
		name += objValue.Type().Name()
		// inst := reflect.New(objValue.Type()).Elem()
		// for i := 0; i < inst.NumField(); i++ {
		for i := 0; i < objValue.Type().NumField(); i++ {
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
func setFields(fieldValue reflect.Value, valmap map[string]flag.Value, tag string) error {
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
		parsers[rType] = parser
	}
	return parsers, nil
}

//Load initializes config : struct fields given by reference, with args : arguments.
//Some custom parsers may be given.
func Load(config interface{}, args []string, customParsers map[reflect.Type]flag.Value) error {
	parsers, err := loadParsers(customParsers)
	if err != nil {
		return err
	}
	tagsmap := make(map[string]reflect.Type)
	if err := getTypesRecursive(reflect.ValueOf(config), tagsmap, ""); err != nil {
		return err
	}
	valmap, err := parseArgs(args, tagsmap, nil, parsers)
	if err != nil {
		return err
	}
	if err := fillStructRecursive(reflect.ValueOf(config), valmap, ""); err != nil {
		return err
	}

	return nil
}

//getStructRecursive initialize a value of any taged Struct given by reference
func getStructRecursive(objValue reflect.Value, parsers map[reflect.Type]flag.Value, key string) (map[string]interface{}, error) {
	name := key
	valmap := make(map[string]interface{})
	// fmt.Printf("objValue begin : %+v\n", objValue)
	switch objValue.Kind() {
	case reflect.Struct:
		name += objValue.Type().Name()
		for i := 0; i < objValue.Type().NumField(); i++ {
			if desc := objValue.Type().Field(i).Tag.Get("description"); len(desc) > 0 {
				if parser, isOk := parsers[objValue.Field(i).Type()]; isOk {
					rvParser := reflect.New(reflect.TypeOf(parser).Elem())

					rvParser.Set(objValue.Field(i))

					fieldName := objValue.Type().Field(i).Name
					if tag := objValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
						fieldName = tag
					}
					if tag := objValue.Type().Field(i).Tag.Get("short"); len(tag) > 0 {
						valmap[strings.ToLower(tag)] = rvParser.Interface().(flag.Value)
					}
					if len(key) == 0 {
						name = fieldName
					} else {
						name = key + "." + fieldName
					}
					valmap[strings.ToLower(name)] = rvParser.Interface().(flag.Value)
				}
				rtValmap, err := getStructRecursive(objValue.Field(i), parsers, name)
				if err != nil {
					return nil, err
				}
				for k, v := range rtValmap {
					valmap[k] = v
				}

			}
		}

	case reflect.Ptr:
		if !objValue.IsNil() {
			rtValmap, err := getStructRecursive(objValue.Elem(), parsers, name)
			if err != nil {
				return nil, err
			}
			for k, v := range rtValmap {
				valmap[k] = v
			}
		}
	default: //SLICE, map ?
		return valmap, nil
	}
	// fmt.Printf("objValue end : %+v\n", objValue)
	return valmap, nil
}
