package main

import (
	"flag"
	"fmt"
	"log"
	"reflect"
)

/* used as example
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
*/

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

//ParseArgs : parses args into value, stored in map[tag]object
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

//FillStructRecursive : uses ParseArgs to recursively initialize an instance of Struct
func FillStructRecursive(objValue reflect.Value, valmap map[string]interface{}) {
	fmt.Println("kind : " + objValue.Kind().String())
	switch objValue.Kind() {
	case reflect.Ptr:
		typ := objValue.Type().Elem()
		inst := reflect.New(typ).Elem()
		switch inst.Kind() {
		case reflect.Struct:
			for i := 0; i < inst.NumField(); i++ {
				//TODO if short
				if tag := inst.Type().Field(i).Tag.Get("short"); len(tag) > 0 {
					println("TODO : tag : short")
				}
				if tag := inst.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
					for tag2, val := range valmap {
						if tag == tag2 {
							fmt.Printf("field %d in the struct is kind of : %s\n", i, objValue.Elem().Field(i).Kind().String())
							if objValue.Elem().Field(i).CanSet() {
								fmt.Printf("will put %s (a kind of : %s) into this field\n", val, reflect.ValueOf(val).Elem().Kind().String())
								objValue.Elem().Field(i).Set(reflect.ValueOf(val).Elem().Convert(objValue.Elem().Field(i).Type()))
								fmt.Printf("dfyj %+v\n", objValue.Elem().Field(i))
							} else {
								log.Fatal("sorry but the type %s is not a settable ...\n", objValue.Elem().Type)
							}
						}
					}
				}
				//recursion TODO
				//FillStructRecursive(objValue.Elem().Field(i), valmap)
			}
		default:
			println("TODO ptr on no struct")
		}

	}

}
