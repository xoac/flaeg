package main

import (
	// "flag"
	"fmt"
	"log"
	"reflect"
)

//Recursive function which browse inside any kind of reflect.Value
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

//Recursive function which browse inside a struct and read some struct tags
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
