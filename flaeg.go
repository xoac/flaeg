package main

import (
	"flag"
	"fmt"
	"log"
	"reflect"
)

//GetTagsRecursive : Recursive function which links in a maps 'short' and 'long' tags with there value
func GetTagsRecursive(objValue reflect.Value, tagsmap map[string]reflect.Type) {
	switch objValue.Kind() {
	case reflect.Struct:
		for i := 0; i < objValue.NumField(); i++ {
			if tag := objValue.Type().Field(i).Tag.Get("short"); len(tag) > 0 {
				tagsmap[tag] = objValue.Field(i).Type()
			}
			if tag := objValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
				tagsmap[tag] = objValue.Field(i).Type()
			}
			GetTagsRecursive(objValue.Field(i), tagsmap)
		}
	case reflect.Array, reflect.Map, reflect.Slice, reflect.Ptr:
		typ := objValue.Type().Elem()
		inst := reflect.New(typ).Elem()
		GetTagsRecursive(inst, tagsmap)
	default:
		return
	}
}

//ParseArgs : parses args into a map[tag]value, using map[type]parser
//args must be formated as like as flag documentation. See https://golang.org/pkg/flag
func ParseArgs(args []string, tagsmap map[string]reflect.Type, parsers map[reflect.Type]flag.Value) map[string]interface{} {
	//Check if all reflect.Type from tagsmap are in parsers
	for tag, rType := range tagsmap {
		if _, ok := parsers[rType]; !ok {
			log.Fatalf("Error tag %s : Parser(flag.Value) not found for reflect.Type %s in the map parsers\n", tag, rType)
		}
	}

	newParsers := map[string]flag.Value{}
	flagSet := flag.NewFlagSet("flaeg.ParseArgs", flag.ExitOnError)
	valmap := make(map[string]interface{})
	for tag, rType := range tagsmap {
		newparser := reflect.New(reflect.TypeOf(parsers[rType]).Elem()).Interface().(flag.Value)
		flagSet.Var(newparser, tag, "help")
		newParsers[tag] = newparser
	}
	flagSet.Parse(args)
	for tag, newParser := range newParsers {
		valmap[tag] = newParser
	}
	return valmap
}

//FillStructRecursive initialize a value of any taged Struct given by reference
func FillStructRecursive(objValue reflect.Value, valmap map[string]interface{}) {
	switch objValue.Kind() {
	case reflect.Struct:
		inst := reflect.New(objValue.Type()).Elem()
		for i := 0; i < inst.NumField(); i++ {
			taged := false
			if tag := inst.Type().Field(i).Tag.Get("short"); len(tag) > 0 {
				taged = true
				SetFields(objValue.Field(i), valmap, tag)
			}
			if tag := inst.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
				taged = true
				SetFields(objValue.Field(i), valmap, tag)
			}
			if !taged {
				FillStructRecursive(objValue.Field(i), valmap)
			}
		}
	case reflect.Ptr:
		if objValue.IsNil() {
			inst := reflect.New(objValue.Type().Elem())
			FillStructRecursive(inst.Elem(), valmap)
			objValue.Set(inst)
		} else {
			FillStructRecursive(objValue.Elem(), valmap)
		}

	case reflect.Slice, reflect.Array, reflect.Map:
		fmt.Println("NEVER HERE ?")
	}
}

// SetFields sets value to fieldValue using tag as key in valmap
func SetFields(fieldValue reflect.Value, valmap map[string]interface{}, tag string) {
	if fieldValue.CanSet() {
		if val, ok := valmap[tag]; ok {
			fieldValue.Set(reflect.ValueOf(val).Elem().Convert(fieldValue.Type()))
		}
	} else {
		log.Fatalf("Error : type %s is not a settable ...\n", fieldValue.Kind())
	}

}
