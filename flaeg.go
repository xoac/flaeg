package main

import (
	"errors"
	"flag"
	"reflect"
	"strings"
	"time"
)

// GetTypesRecursive links in flagmap a flag with its StructField
// You can whether provide objValue on a structure or a pointer to structure as first argument
// Flags are genereted from field name or from StructTag
func getTypesRecursive(objValue reflect.Value, flagmap map[string]reflect.StructField, key string) error {
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
					if _, ok := flagmap[strings.ToLower(tag)]; ok {
						return errors.New("Tag already exists: " + tag)
					}
					flagmap[strings.ToLower(tag)] = objValue.Type().Field(i)
				}
				if len(key) == 0 {
					name = fieldName
				} else {
					name = key + "." + fieldName
				}
				if _, ok := flagmap[strings.ToLower(name)]; ok {
					return errors.New("Tag already exists: " + name)
				}
				flagmap[strings.ToLower(name)] = objValue.Type().Field(i)
				if err := getTypesRecursive(objValue.Field(i), flagmap, name); err != nil {
					return err
				}
			}
		}
	case reflect.Ptr:
		if len(key) > 0 {
			flagmap[strings.ToLower(name)] = reflect.StructField{
				flagmap[strings.ToLower(name)].Name,
				flagmap[strings.ToLower(name)].PkgPath,
				reflect.TypeOf(false),
				flagmap[strings.ToLower(name)].Tag,
				flagmap[strings.ToLower(name)].Offset,
				flagmap[strings.ToLower(name)].Index,
				flagmap[strings.ToLower(name)].Anonymous,
			}
		}
		typ := objValue.Type().Elem()
		inst := reflect.New(typ).Elem()
		if err := getTypesRecursive(inst, flagmap, name); err != nil {
			return err
		}

	case reflect.Array, reflect.Map, reflect.Slice:
		typ := objValue.Type().Elem()
		inst := reflect.New(typ).Elem()
		if err := getTypesRecursive(inst, flagmap, name); err != nil {
			return err
		}
	default:
		return nil
	}
	return nil
}

//ParseArgs : parses args return valmap map[flag]Getter, using parsers map[type]Getter
//args must be formated as like as flag documentation. See https://golang.org/pkg/flag
func parseArgs(args []string, flagmap map[string]reflect.StructField, parsers map[reflect.Type]flag.Getter) (map[string]flag.Getter, error) {
	valmap := make(map[string]flag.Getter)
	flagList := []*flag.Flag{}
	visitor := func(fl *flag.Flag) {
		// fmt.Printf("inside : %s\n", fl.Name)
		flagList = append(flagList, fl)
	}
	newParsers := map[string]flag.Getter{}
	flagSet := flag.NewFlagSet("flaeg.ParseArgs", flag.ExitOnError)
	for tag, structField := range flagmap {

		if parser, ok := parsers[structField.Type]; ok {
			newparser := reflect.New(reflect.TypeOf(parser).Elem()).Interface().(flag.Getter)
			// fmt.Printf("help to print : %s\n", structField.Tag.Get("description"))
			flagSet.Var(newparser, tag, structField.Tag.Get("description"))
			newParsers[tag] = newparser
		}
	}

	if err := flagSet.Parse(args); err != nil {
		return nil, err
	}
	flagSet.Visit(visitor)
	for _, flag := range flagList {
		valmap[flag.Name] = newParsers[flag.Name]
	}

	return valmap, nil
}

//FillStructRecursive initialize a value of any taged Struct given by reference
func fillStructRecursive(objValue reflect.Value, defaultValue reflect.Value, valmap map[string]flag.Getter, key string) error {
	name := key
	switch objValue.Kind() {

	case reflect.Struct:
		name += objValue.Type().Name()
		for i := 0; i < objValue.Type().NumField(); i++ {
			if tag := objValue.Type().Field(i).Tag.Get("description"); len(tag) > 0 {

				fieldName := objValue.Type().Field(i).Name
				if tag := objValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
					fieldName = tag
				}
				if len(key) == 0 {
					name = fieldName
				} else {
					name = key + "." + fieldName
				}

				if objValue.Field(i).Kind() == reflect.Ptr {
					if err := fillStructRecursive(objValue.Field(i), defaultValue.Field(i), valmap, name); err != nil {
						return err
					}
					return nil
				}

				if val, ok := valmap[strings.ToLower(name)]; ok {
					if err := setFields(objValue.Field(i), val); err != nil {
						return err
					}
				} else {
					if tag := objValue.Type().Field(i).Tag.Get("short"); len(tag) > 0 && valmap[strings.ToLower(tag)] != nil {
						if err := setFields(objValue.Field(i), valmap[strings.ToLower(tag)]); err != nil {
							return err
						}
					} else {
						if objValue.Field(i).CanSet() {
							objValue.Field(i).Set(defaultValue.Field(i))
						} else {
							return errors.New(objValue.Field(i).Type().String() + " is not settable.")
						}
						if err := fillStructRecursive(objValue.Field(i), defaultValue.Field(i), valmap, name); err != nil {
							return err
						}
					}
				}
			}
		}

	case reflect.Ptr:
		if objValue.IsNil() {
			contains := false

			if _, ok := valmap[strings.ToLower(name)]; !ok {
				for flag := range valmap {
					// TODO replace by regexp
					if strings.Contains(flag, strings.ToLower(name)+".") {
						contains = true
						break
					}
				}
			} else {
				contains = true
			}

			if contains {
				inst := reflect.New(objValue.Type().Elem())
				if inst.Elem().CanSet() {
					inst.Elem().Set(defaultValue.Elem())
				} else {
					return errors.New(inst.Elem().Type().String() + " is not settable.")
				}
				if err := fillStructRecursive(inst.Elem(), defaultValue.Elem(), valmap, name); err != nil {
					return err
				}
				objValue.Set(inst)
			}

		} else {
			if err := fillStructRecursive(objValue.Elem(), defaultValue.Elem(), valmap, name); err != nil {
				return err
			}
		}

	default:
		return nil
	}
	return nil
}

// SetFields sets value to fieldValue using tag as key in valmap
func setFields(fieldValue reflect.Value, val flag.Getter) error {
	// if reflect.DeepEqual(fieldValue.Interface(), reflect.New(fieldValue.Type()).Elem().Interface()) {
	if fieldValue.CanSet() {
		fieldValue.Set(reflect.ValueOf(val).Elem().Convert(fieldValue.Type()))
	} else {
		return errors.New(fieldValue.Type().String() + " is not settable.")
	}

	// }
	return nil
}

//loadParsers loads default parsers and custom parsers given as parameter. Return a map [reflect.Type]parsers
func loadParsers(customParsers map[reflect.Type]flag.Getter) (map[reflect.Type]flag.Getter, error) {
	parsers := map[reflect.Type]flag.Getter{}
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
func Load(config interface{}, defaultValue interface{}, args []string, customParsers map[reflect.Type]flag.Getter) error {
	parsers, err := loadParsers(customParsers)
	if err != nil {
		return err
	}
	tagsmap := make(map[string]reflect.StructField)
	if err := getTypesRecursive(reflect.ValueOf(config), tagsmap, ""); err != nil {
		return err
	}
	valmap, err := parseArgs(args, tagsmap, parsers)
	if err != nil {
		return err
	}
	if err := fillStructRecursive(reflect.ValueOf(config), reflect.ValueOf(defaultValue), valmap, ""); err != nil {
		return err
	}

	return nil
}
