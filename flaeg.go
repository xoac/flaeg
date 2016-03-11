package main

import (
	// "flag"
	"fmt"
	"log"
	"reflect"
)

// ReflectRecursive : Recursive function which browse inside any kind of reflect.Value
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

//ReadTagsRecursive : Recursive function which browse inside a struct and read some struct tags
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

//GetTagsRecursive : Recursive function which link in a maps 'short' and 'long' tags with there value
func GetTagsRecursive(objType reflect.Value) (tagsmap map[string]reflect.Value) {
	tagsmap = make(map[string]reflect.Value)
	if objType.Kind() == reflect.Struct {
		for i := 0; i < objType.NumField(); i++ {
			fmt.Printf("Kind %s\n", objType.Field(i).Kind().String())
			if tag := objType.Type().Field(i).Tag.Get("short"); len(tag) > 0 {
				tagsmap["-"+tag] = objType.Field(i)
			}
			if tag := objType.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
				tagsmap["--"+tag] = objType.Field(i)
			}

			switch objType.Field(i).Kind() {
			case reflect.Struct:
				for k, v := range GetTagsRecursive(objType.Field(i)) {
					tagsmap[k] = v
				}
			case reflect.Map:
				for _, key := range objType.Field(i).MapKeys() {
					for k, v := range GetTagsRecursive(objType.Field(i).MapIndex(key)) {
						tagsmap[k] = v
					}
				}
			case reflect.Slice:
				for j := 0; j < objType.Field(i).Len(); j++ {
					fmt.Printf("Slice elem : %+v", objType.Field(i).Index(j))
					for k, v := range GetTagsRecursive(objType.Field(i).Index(j)) {
						tagsmap[k] = v
					}
				}
			case reflect.Interface:
				for k, v := range GetTagsRecursive(objType.Field(i).Elem()) {
					tagsmap[k] = v
				}
			case reflect.Ptr:
				val := objType.Field(i).Elem()
				if !val.IsValid() {
					fmt.Printf("%+v : IS NOT VALID\n", objType.Field(i))
					typ := objType.Field(i).Type().Elem()
					inst := reflect.New(typ)
					fmt.Printf("%+v\n", inst.Elem())
					fmt.Printf("%s\n", inst.Elem().Kind())
					// for k, v := range GetTagsRecursive(reflect.ValueOf(inst.Elem().Interface())) {
					for k, v := range GetTagsRecursive(inst.Elem()) {
						fmt.Printf("%s -> %+v", k, v)
						tagsmap[k] = v
					}
				} else {
					for k, v := range GetTagsRecursive(val) {
						tagsmap[k] = v
					}
				}
			}
		}

	} else {
		log.Printf("sorry but %s is not a %s : ", objType.Kind().String(), reflect.Struct.String())
		return

	}
	return
}
