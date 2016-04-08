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

// GetTypesRecursive links in flagmap a flag with its StructField
// You can whether provide objValue on a structure or a pointer to structure as first argument
// Flags are genereted from field name or from StructTag
func getTypesRecursive(objValue reflect.Value, flagmap map[string]reflect.StructField, key string) error {
	name := key
	switch objValue.Kind() {
	case reflect.Struct:
		name += objValue.Type().Name()
		for i := 0; i < objValue.NumField(); i++ {
			if len(objValue.Type().Field(i).Tag.Get("description")) > 0 {
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
			field := flagmap[strings.ToLower(name)]
			field.Type = reflect.TypeOf(false)
			if tag := field.Tag.Get("short"); len(tag) > 0 {
				flagmap[strings.ToLower(tag)] = field
			}
			flagmap[strings.ToLower(name)] = field
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
func parseArgs(args []string, flagmap map[string]reflect.StructField, parsers map[reflect.Type]Parser) (map[string]Parser, error) {
	//Return var
	valmap := make(map[string]Parser)
	//Preprocess
	// Reverse flagmap map[string]reflect.StructField in map[reflect.StructField][]string
	// & Delete unparsable flags in a slice
	fieldmap := map[string][]string{}
	flags := []string{}
	i := 0
	for flag, field := range flagmap {
		if _, ok := parsers[field.Type]; ok {
			if len(fieldmap[field.PkgPath+field.Name]) == 0 {
				flags = append(flags, flag)
				i++
			}
			fieldmap[field.PkgPath+field.Name] = append(fieldmap[field.PkgPath+field.Name], flag)

		}
	}

	//Visitor in flag.Parse
	flagList := []*flag.Flag{}
	visitor := func(fl *flag.Flag) {
		flagList = append(flagList, fl)
	}
	newParsers := map[string]Parser{}
	flagSet := flag.NewFlagSet("flaeg.Load", flag.ContinueOnError)
	//Disable output
	flagSet.SetOutput(ioutil.Discard)

	// for tag, structField := range flagmap {
	for _, flag := range flags {
		structField := flagmap[flag]
		if parser, ok := parsers[structField.Type]; ok {
			newparser := reflect.New(reflect.TypeOf(parser).Elem()).Interface().(Parser)
			if shortLongFlags, ok := fieldmap[structField.PkgPath+structField.Name]; len(shortLongFlags) == 2 && ok {
				if len(shortLongFlags[0]) > len(shortLongFlags[1]) {
					flagSet.VarP(newparser, shortLongFlags[0], shortLongFlags[1], structField.Tag.Get("description"))
				} else {
					flagSet.VarP(newparser, shortLongFlags[1], shortLongFlags[0], structField.Tag.Get("description"))
				}
				newParsers[shortLongFlags[0]] = newparser
				newParsers[shortLongFlags[1]] = newparser
			} else {
				flagSet.Var(newparser, flag, structField.Tag.Get("description"))
				newParsers[flag] = newparser
			}
		}
	}

	// Call custom helper
	if err := flagSet.Parse(args); err != nil {
		fmt.Printf("error:%+v\n", err)
		if err == flag.ErrHelp {
			fmt.Printf("HELP\n")
		}
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

//FillStructRecursive initialize a value of any taged Struct given by reference
func fillStructRecursive(objValue reflect.Value, defaultValmap map[string]reflect.Value, valmap map[string]Parser, key string) error {
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
					if err := fillStructRecursive(objValue.Field(i), defaultValmap, valmap, name); err != nil {
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
							objValue.Field(i).Set(defaultValmap[strings.ToLower(name)])
						} else {
							return errors.New(objValue.Field(i).Type().String() + " is not settable.")
						}
						if err := fillStructRecursive(objValue.Field(i), defaultValmap, valmap, name); err != nil {
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
				objValue.Set(defaultValmap[strings.ToLower(name)])
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

//Load initializes config : struct fields given by reference, with args : arguments.
//Some custom parsers may be given.
func Load(config interface{}, defaultValue interface{}, args []string, customParsers map[reflect.Type]Parser) error {
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
	defaultValmap := make(map[string]reflect.Value)
	if err := getDefaultValue(reflect.ValueOf(defaultValue), defaultValmap, ""); err != nil {
		return err
	}
	if err := fillStructRecursive(reflect.ValueOf(config), defaultValmap, valmap, ""); err != nil {
		return err
	}

	return nil
}

//PrintError takes a not nil error and prints command line help
func PrintError(err error, flagmap map[string]reflect.StructField, defaultValmap map[string]reflect.Value, parsers map[reflect.Type]Parser) {
	if err != flag.ErrHelp {
		fmt.Printf("Error : %s\n", err)
	}
	PrintHelp(flagmap, defaultValmap, parsers)
}

func getDefaultValue(defaultValue reflect.Value, defaultValmap map[string]reflect.Value, key string) error {

	name := key
	switch defaultValue.Kind() {
	case reflect.Struct:
		name += defaultValue.Type().Name()
		for i := 0; i < defaultValue.NumField(); i++ {
			if len(defaultValue.Type().Field(i).Tag.Get("description")) > 0 {
				fieldName := defaultValue.Type().Field(i).Name
				if tag := defaultValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
					fieldName = tag
				}
				if tag := defaultValue.Type().Field(i).Tag.Get("short"); len(tag) > 0 {
					if _, ok := defaultValmap[strings.ToLower(tag)]; ok {
						return errors.New("Tag already exists: " + tag)
					}
					defaultValmap[strings.ToLower(tag)] = defaultValue.Field(i)
				}
				if len(key) == 0 {
					name = fieldName
				} else {
					name = key + "." + fieldName
				}
				if _, ok := defaultValmap[strings.ToLower(name)]; ok {
					return errors.New("Tag already exists: " + name)
				}
				defaultValmap[strings.ToLower(name)] = defaultValue.Field(i)
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

//PrintHelp generates and prints command line help
func PrintHelp(flagmap map[string]reflect.StructField, defaultValmap map[string]reflect.Value, parsers map[reflect.Type]Parser) error {
	// Define a templates
	const helper = `
Usage: {{.ProgName}}                                  run {{.ProgName}} with default values
   or: {{.ProgName}} -flag args | -flag=args ...      use args as value on flags
   or: {{.ProgName}} -flag | -flag=true ...           set true if flags are boolean      

Flags:{{range $j, $flag := .Flags}}{{$description:= index $.Descriptions $j}}{{$defaultValues := index $.DefaultValues $j}}
{{printf "\t%-50s %s (default \"%s\")" $flag $description $defaultValues}}{{end}}`

	// Preprocess data

	// Reverse flagmap map[string]reflect.StructField in map[reflect.StructField][]string
	// & Sort alphabetically & Delete unparsable flags in a slice
	fieldmap := map[string][]string{}
	flags := []string{}
	i := 0
	for flag, field := range flagmap {
		if _, ok := parsers[field.Type]; ok {
			if len(fieldmap[field.PkgPath+field.Name]) == 0 {
				flags = append(flags, flag)
				i++
			}
			fieldmap[field.PkgPath+field.Name] = append(fieldmap[field.PkgPath+field.Name], flag)

		}
	}
	sort.Strings(flags)

	// Process data
	printDescriptions := []string{}
	printDefaultValues := []string{}
	printFlags := []string{}
	for _, flag := range flags {
		field := flagmap[flag]
		if len(fieldmap[field.PkgPath+field.Name]) == 2 {
			printFlags = append(printFlags, "-"+fieldmap[field.PkgPath+field.Name][0]+", -"+fieldmap[field.PkgPath+field.Name][1])
		} else {
			printFlags = append(printFlags, "-"+fieldmap[field.PkgPath+field.Name][0])
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
