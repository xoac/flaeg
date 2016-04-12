package main

import (
	"errors"
	"fmt"
	flag "github.com/ogier/pflag"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"time"
)

type StructField struct {
	reflect.StructField
	Short string
}

// GetTypesRecursive links in flagmap a flag with its StructField
// You can whether provide objValue on a structure or a pointer to structure as first argument
// Flags are genereted from field name or from StructTag
func getTypesRecursive(objValue reflect.Value, flagmap map[string]StructField, key string) error {
	name := key
	switch objValue.Kind() {
	case reflect.Struct:

		for i := 0; i < objValue.NumField(); i++ {
			if objValue.Type().Field(i).Anonymous {
				if err := getTypesRecursive(objValue.Field(i), flagmap, name); err != nil {
					return err
				}
			} else if len(objValue.Type().Field(i).Tag.Get("description")) > 0 {

				name += objValue.Type().Name()
				fieldName := objValue.Type().Field(i).Name
				if tag := objValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
					fieldName = tag
				}
				if len(key) == 0 {
					//Lower Camel Case
					//name = strings.ToLower(string(fieldName[0])) + fieldName[1:]
					name = strings.ToLower(fieldName)
				} else {
					name = key + "." + strings.ToLower(fieldName)
				}
				if _, ok := flagmap[name]; ok {
					return errors.New("Tag already exists: " + name)
				}
				structField := StructField{objValue.Type().Field(i), objValue.Type().Field(i).Tag.Get("short")}
				flagmap[name] = structField

				if err := getTypesRecursive(objValue.Field(i), flagmap, name); err != nil {
					return err
				}
			}

		}
	case reflect.Ptr:
		if len(key) > 0 {
			field := flagmap[name]
			field.Type = reflect.TypeOf(false)
			flagmap[name] = field
		}
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
func parseArgs(args []string, flagmap map[string]StructField, parsers map[reflect.Type]Parser) (map[string]Parser, error) {
	//Return var
	valmap := make(map[string]Parser)
	//Visitor in flag.Parse
	flagList := []*flag.Flag{}
	visitor := func(fl *flag.Flag) {
		flagList = append(flagList, fl)
	}
	newParsers := map[string]Parser{}
	flagSet := flag.NewFlagSet("flaeg.Load", flag.ContinueOnError)
	//Disable output
	flagSet.SetOutput(ioutil.Discard)

	for flag, structField := range flagmap {
		//for _, flag := range flags {
		//structField := flagmap[flag]
		if parser, ok := parsers[structField.Type]; ok {
			newparser := reflect.New(reflect.TypeOf(parser).Elem()).Interface().(Parser)
			if len(structField.Short) == 1 {
				// fmt.Printf("short : %s long : %s\n", structField.Short, flag)
				flagSet.VarP(newparser, flag, structField.Short, structField.Tag.Get("description"))
			} else {
				flagSet.Var(newparser, flag, structField.Tag.Get("description"))
			}
			newParsers[flag] = newparser
		}
	}

	// Call custom helper
	if err := flagSet.Parse(args); err != nil {
		return nil, err
	}

	//Fill flagList with parsed flags
	flagSet.Visit(visitor)
	//Return parsers on parsed flag
	for _, flag := range flagList {
		valmap[flag.Name] = newParsers[flag.Name]
	}

	return valmap, nil
}

func getDefaultValue(defaultValue reflect.Value, defaultValmap map[string]reflect.Value, key string) error {
	name := key
	switch defaultValue.Kind() {
	case reflect.Struct:

		for i := 0; i < defaultValue.NumField(); i++ {
			if defaultValue.Type().Field(i).Anonymous {
				if err := getDefaultValue(defaultValue.Field(i), defaultValmap, name); err != nil {
					return err
				}
			} else if len(defaultValue.Type().Field(i).Tag.Get("description")) > 0 {
				name += defaultValue.Type().Name()
				fieldName := defaultValue.Type().Field(i).Name
				if tag := defaultValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
					fieldName = tag
				}
				if len(key) == 0 {
					name = strings.ToLower(fieldName)
				} else {
					name = key + "." + strings.ToLower(fieldName)
				}
				if _, ok := defaultValmap[name]; ok {
					return errors.New("Tag already exists: " + name)
				}
				defaultValmap[name] = defaultValue.Field(i)
				if err := getDefaultValue(defaultValue.Field(i), defaultValmap, name); err != nil {
					return err
				}
			}
		}
	case reflect.Ptr:
		if !defaultValue.IsNil() {
			if err := getDefaultValue(defaultValue.Elem(), defaultValmap, name); err != nil {
				return err
			}
		}
	default:
		return nil
	}
	return nil
}

//FillStructRecursive initialize a value of any taged Struct given by reference
func fillStructRecursive(objValue reflect.Value, defaultValmap map[string]reflect.Value, valmap map[string]Parser, key string) error {
	name := key
	switch objValue.Kind() {
	case reflect.Struct:

		for i := 0; i < objValue.Type().NumField(); i++ {
			if objValue.Type().Field(i).Anonymous {
				if err := fillStructRecursive(objValue.Field(i), defaultValmap, valmap, name); err != nil {
					return err
				}
			} else if len(objValue.Type().Field(i).Tag.Get("description")) > 0 {
				name += objValue.Type().Name()
				fieldName := objValue.Type().Field(i).Name
				if tag := objValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
					fieldName = tag
				}
				if len(key) == 0 {
					name = strings.ToLower(fieldName)
				} else {
					name = key + "." + strings.ToLower(fieldName)
				}

				if objValue.Field(i).Kind() == reflect.Ptr {
					if err := fillStructRecursive(objValue.Field(i), defaultValmap, valmap, name); err != nil {
						return err
					}
					return nil
				}

				if val, ok := valmap[name]; ok {
					if err := setFields(objValue.Field(i), val); err != nil {
						return err
					}
				} else if objValue.Field(i).CanSet() {
					objValue.Field(i).Set(defaultValmap[name])
				} else {
					return errors.New(objValue.Field(i).Type().String() + " is not settable.")
				}
				if err := fillStructRecursive(objValue.Field(i), defaultValmap, valmap, name); err != nil {
					return err
				}
			}
		}

	case reflect.Ptr:
		if objValue.IsNil() {
			contains := false

			if _, ok := valmap[name]; !ok {
				for flag := range valmap {
					// TODO replace by regexp
					if strings.Contains(flag, name+".") {
						contains = true
						break
					}
				}
			} else {
				contains = true
			}

			if contains {
				objValue.Set(defaultValmap[name])
				if err := fillStructRecursive(objValue.Elem(), defaultValmap, valmap, name); err != nil {
					return err
				}
			}

		} else {
			if err := fillStructRecursive(objValue.Elem(), defaultValmap, valmap, name); err != nil {
				return err
			}
		}

	default:
		return nil
	}
	return nil
}

// SetFields sets value to fieldValue using tag as key in valmap
func setFields(fieldValue reflect.Value, val Parser) error {
	if fieldValue.CanSet() {
		fieldValue.Set(reflect.ValueOf(val).Elem().Convert(fieldValue.Type()))
	} else {
		return errors.New(fieldValue.Type().String() + " is not settable.")
	}
	return nil
}

//loadParsers loads default parsers and custom parsers given as parameter. Return a map [reflect.Type]parsers
func loadParsers(customParsers map[reflect.Type]Parser) (map[reflect.Type]Parser, error) {
	parsers := map[reflect.Type]Parser{}
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

//PrintHelp generates and prints command line help
func PrintHelp(flagmap map[string]StructField, defaultValmap map[string]reflect.Value, parsers map[reflect.Type]Parser) error {
	// Define a templates
	// Using POSXE STD : http://pubs.opengroup.org/onlinepubs/9699919799/
	//TO DO : program description, bugs report, home page, full doc
	const helper = `
Usage: {{.ProgName}} [--flag=flag_argument] [-f[flag_argument]] ...     set flag_argument to flag(s)
   or: {{.ProgName}} [--flag[=true|false| ]] [-f[true|false| ]] ...     set true/false to boolean flag(s)   

Flags:{{range $j, $flag := .Flags}}{{$description:= index $.Descriptions $j}}{{$defaultValues := index $.DefaultValues $j}}
{{printf "\t%-50s %s (default \"%s\")" $flag $description $defaultValues}}{{end}}`

	// Preprocess data

	// Sort alphabetically & Delete unparsable flags in a slice
	flags := []string{}
	for flag, field := range flagmap {
		if _, ok := parsers[field.Type]; ok {
			flags = append(flags, flag)
		}
	}
	sort.Strings(flags)

	// Process data
	printDescriptions := []string{}
	printDefaultValues := []string{}
	printFlags := []string{}
	for _, flag := range flags {
		field := flagmap[flag]
		if len(field.Short) == 1 {
			printFlags = append(printFlags, "-"+field.Short+", --"+flag)
		} else {
			printFlags = append(printFlags, "--"+flag)
		}
		printDescriptions = append(printDescriptions, field.Tag.Get("description"))
		//flag on pointer ?
		if defaultValmap[flag].Kind() != reflect.Ptr {
			// Set defaultValue on parsers
			parsers[field.Type].SetValue(defaultValmap[flag].Interface())
		}
		printDefaultValues = append(printDefaultValues, parsers[field.Type].String())

	}

	// Get ProgramName
	_, progName := path.Split(os.Args[0])

	// Use a struct to give data to template
	tempStruct := struct {
		ProgName      string
		Flags         []string
		Descriptions  []string
		DefaultValues []string
	}{
		progName,
		printFlags,
		printDescriptions,
		printDefaultValues,
	}

	//Run Template
	tmplHelper, err := template.New("helper").Parse(helper)
	if err != nil {
		return err
	}
	err = tmplHelper.Execute(os.Stdout, tempStruct)
	if err != nil {
		return err
	}
	//And footer
	fmt.Fprintf(os.Stdout, "\n\t%-50s %s\n", "-h, -help", "Print Help (this message) and exit")
	return nil
}

//PrintError takes a not nil error and prints command line help
func PrintError(err error, flagmap map[string]StructField, defaultValmap map[string]reflect.Value, parsers map[reflect.Type]Parser) error {
	if err != flag.ErrHelp {
		fmt.Printf("Error : %s\n", err)
	}
	PrintHelp(flagmap, defaultValmap, parsers)
	return err
}

//LoadWithParsers initializes config : struct fields given by reference, with args : arguments.
//Some custom parsers may be given.
func LoadWithParsers(config interface{}, defaultValue interface{}, args []string, customParsers map[reflect.Type]Parser) error {
	parsers, err := loadParsers(customParsers)
	if err != nil {
		return err
	}
	tagsmap := make(map[string]StructField)
	if err := getTypesRecursive(reflect.ValueOf(config), tagsmap, ""); err != nil {
		return err
	}
	defaultValmap := make(map[string]reflect.Value)
	if err := getDefaultValue(reflect.ValueOf(defaultValue), defaultValmap, ""); err != nil {
		return err
	}
	valmap, err := parseArgs(args, tagsmap, parsers)
	if err != nil {
		return PrintError(err, tagsmap, defaultValmap, parsers)
	}
	if err := fillStructRecursive(reflect.ValueOf(config), defaultValmap, valmap, ""); err != nil {
		return err
	}

	return nil
}

//Load initializes config : struct fields given by reference, with args : arguments.
//Some custom parsers may be given.
func Load(config interface{}, defaultValue interface{}, args []string) error {
	return LoadWithParsers(config, defaultValue, args, nil)
}
