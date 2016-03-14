package main

import (
	"flag"
	"fmt"
	"log"
	"reflect"
)

// ReflectRecursive : Recursive function which browses inside any kind of reflect.Value
func ReflectRecursive(original reflect.Value) {
	fmt.Println("kind : " + original.Kind().String())
	switch original.Kind() {
	case reflect.Struct:
		for i := 0; i < original.NumField(); i++ {
			ReflectRecursive(original.Field(i))
		}
	case reflect.Map:
		for _, key := range original.MapKeys() {
			ReflectRecursive(original.MapIndex(key))
		}
	case reflect.Slice:
		for i := 0; i < original.Len(); i++ {
			ReflectRecursive(original.Index(i))
		}
	case reflect.Interface:
		ReflectRecursive(original.Elem())
	case reflect.String:
		f := original.Interface()
		fmt.Println("String content : " + reflect.ValueOf(f).String())
	}
}

//ReadTagsRecursive : Recursive function which browses inside a struct and read some struct tags
func ReadTagsRecursive(objType reflect.Type) {
	if objType.Kind() == reflect.Struct {
		for i := 0; i < objType.NumField(); i++ {
			fmt.Printf("\nVAR =%+v \n", objType.Field(i))
			fmt.Println("group : " + objType.Field(i).Tag.Get("group"))
			fmt.Println("short : " + objType.Field(i).Tag.Get("short"))
			fmt.Println("long : " + objType.Field(i).Tag.Get("long"))
			fmt.Println("description : " + objType.Field(i).Tag.Get("description"))
			if objType.Field(i).Type.Kind() == reflect.Struct {
				ReadTagsRecursive(objType.Field(i).Type)
			}
		}
	} else {
		log.Fatal("sorry but %s is not a %s : ", objType.Kind().String(), reflect.Struct.String())
	}
}

//GetTagsRecursive : Recursive function which links in a maps 'short' and 'long' tags with there value
func GetTagsRecursive(objValue reflect.Value, tagsmap map[string]reflect.Type) {
	switch objValue.Kind() {
	case reflect.Struct:
		for i := 0; i < objValue.NumField(); i++ {
			if tag := objValue.Type().Field(i).Tag.Get("short"); len(tag) > 0 {
				tagsmap[""+tag] = objValue.Field(i).Type()
			}
			if tag := objValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
				tagsmap[""+tag] = objValue.Field(i).Type()
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

//ParseArgs : parses args into value, stored in map[tag]object
func ParseArgs(tagsmap map[string]reflect.Type, parsers map[reflect.Type]flag.Value) map[string]interface{} {
	newParsers := map[string]flag.Value{}
	valmap := make(map[string]interface{})
	for tag, rType := range tagsmap {
		newparser := reflect.New(reflect.TypeOf(parsers[rType]).Elem()).Interface().(flag.Value)
		fmt.Println(newparser)
		flag.Var(newparser, tag, "help")
		newParsers[tag] = newparser
	}
	flag.Parse()
	for tag, newParser := range newParsers {
		valmap[tag] = newParser
	}
	return valmap
}

// func FillStructRecursive(strct *interface{}, valmap[string]interface{})
//  {

//  }

type parserString string

func (p *parserString) Set(str string) error {
	*p = parserString(str)
	return nil
}
func (p *parserString) String() string {
	return string(*p)
}
