package main

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

//example of complex Struct
type ownerInfo struct {
	Name         string    `long:"owner.name" description:"overwrite owner name"`         //
	Organization string    `long:"owner.org" description:"overwrite owner organisation"`  //
	Bio          string    `long:"owner.bio" description:"overwrite owner biography"`     //
	Dob          time.Time `long:"owner.dob" description:"overwrite owner date of birth"` //
}
type databaseInfo struct {
	Server        string `long:"database.srv" description:"overwrite database server ip address"`                     //
	ConnectionMax int    `long:"database.comax" description:"overwrite maximum number of connection on the database"` //
	Enable        bool   `long:"database.ena" description:"overwrite database enable"`                                //
}
type serverInfo struct {
	IP string `long:"servers.ip" description:"overwrite server ip address"`        //
	Dc string `long:"servers.dc" description:"overwrite server domain controller"` //
}
type clientInfo struct {
	Data  []int        `long:"clients.data" description:"overwrite clients data"` //
	Hosts []serverInfo `group:"clients.hosts" description:"overwrite clients host names"`
}
type example struct {
	Title    string                `short:"t" description:"overwrite title"` //
	Owner    ownerInfo             `group:"Owner info"`
	Database databaseInfo          `group:"Database info"`
	Servers  map[string]serverInfo `group:"Servers" description:"overwrite servers info --servers.[ip|dc] [srv name]: value"`
	Clients  *clientInfo           `group:"Clients"`
}

/*//Test function ReflectRecursive
func TestReflectRecursive(t *testing.T) {
	//Test slice, string
	// tabStr := []string{"un", "deux", "trois"}
	// ReflectRecursive(reflect.ValueOf(tabStr))

	// //Test struct, slice, string
	// var cl1 clientInfo
	// cl1.Hosts = []string{"un", "deux", "trois"}
	// ReflectRecursive(reflect.ValueOf(cl1))

	// //Test map, struct , string
	// var srv1 map[string]serverInfo
	// srv1 = make(map[string]serverInfo)
	// srv1["first"] = serverInfo{"192.168.2.2", "smth"}
	// ReflectRecursive(reflect.ValueOf(srv1))

	//Test all
	// var ex1 example
	// ex1.init()
	// fmt.Printf("%+v\n", ex1)
	// ReflectRecursive(reflect.ValueOf(ex1))

}

func TestReadTagsRecursive(t *testing.T) {
	//Test struct, slice, string
	fmt.Println("--------------Test struct, slice, string--------------------")
	var cl1 clientInfo
	cl1.Hosts = []serverInfo{{"ip1", "dc1"}, {"ip2", "dc2"}}
	ReadTagsRecursive(reflect.TypeOf(cl1))

	//Test all
	fmt.Println("------------------Test all------------------------------------")
	var ex1 example
	ex1.init()
	ReadTagsRecursive(reflect.TypeOf(ex1))
}

//Init structs
func (ex *example) init() {
	ex.Title = "myTitle"
	ex.Owner.Name = "myName"
	ex.Owner.Organization = "myOrg"
	ex.Owner.Dob = time.Now()
	ex.Database.Server = "192.168.1.2"
	ex.Database.ConnectionMax = 5000
	ex.Database.Enable = true
	ex.Servers = make(map[string]serverInfo)
	ex.Servers["first"] = serverInfo{"192.168.2.2", "smth"}
	//ex.Clients->Hosts[0] = "one"
}
*/

func TestGetTagsRecursive(t *testing.T) {
	//Test all
	var ex1 example
	tagsmap := make(map[string]reflect.Type)
	GetTagsRecursive(reflect.ValueOf(ex1), tagsmap)

	checkType := map[string]reflect.Type{
		"owner.org":      reflect.TypeOf(""),
		"database.ena":   reflect.TypeOf(true),
		"owner.bio":      reflect.TypeOf(""),
		"database.comax": reflect.TypeOf(1),
		"database.srv":   reflect.TypeOf(""),
		"servers.ip":     reflect.TypeOf(""),
		"owner.name":     reflect.TypeOf(""),
		"servers.dc":     reflect.TypeOf(""),
		"clients.data":   reflect.TypeOf([]int{}),
		"t":              reflect.TypeOf(""),
		"owner.dob":      reflect.TypeOf(time.Now()),
	}
	for tag, tagType := range tagsmap {
		if checkType[tag] != tagType {
			t.Fatalf("Type %s (of tag : %s) doesn't match with %s\n", tagType, tag, checkType[tag])
		}
	}

}

// -- custom Value
type customValue []int

func bracket(r rune) bool {
	return r == '{' || r == '}' || r == ',' || r == ';'
}
func (c *customValue) Set(s string) error {
	tabStr := strings.FieldsFunc(s, bracket)
	for _, str := range tabStr {
		v, err := strconv.Atoi(str)
		if err != nil {
			return err
		}
		*c = append(*c, v)
	}
	return nil
}

func (c *customValue) String() string { return fmt.Sprintf("%v", *c) }

func TestParseArgs(t *testing.T) {
	parsers := map[reflect.Type]flag.Value{}
	var myStringParser stringValue
	var myBoolParser boolValue
	var myIntParser intValue
	var myCustomParser customValue
	var myTimeParser timeValue
	parsers[reflect.TypeOf("")] = &myStringParser
	parsers[reflect.TypeOf(true)] = &myBoolParser
	parsers[reflect.TypeOf(1)] = &myIntParser
	parsers[reflect.TypeOf([]int{})] = &myCustomParser
	parsers[reflect.TypeOf(time.Now())] = &myTimeParser

	//Test all
	//fail !
	var ex1 example
	tagsmap := make(map[string]reflect.Type)

	GetTagsRecursive(reflect.ValueOf(ex1), tagsmap)
	fmt.Println(tagsmap)

	pargs := ParseArgs([]string{"servers.dc=toto", ""}, tagsmap, parsers)

	fmt.Printf("parsers : %+v\n", pargs)
}

func TestFillStructRecursive(t *testing.T) {
	var srv1 serverInfo
	parsers := map[reflect.Type]flag.Value{}
	var myStringParser stringValue
	parsers[reflect.TypeOf("")] = &myStringParser

	tagsmap := make(map[string]reflect.Type)
	GetTagsRecursive(reflect.ValueOf(srv1), tagsmap)
	pargs := ParseArgs([]string{"servers.dc=toto", "servers.ip=tztz"}, tagsmap, parsers)
	FillStructRecursive(reflect.ValueOf(&srv1), pargs)
	fmt.Printf("%+v\n", srv1)
	//rValue := FillStructRecursive(reflect.ValueOf(&srv1), pargs)
	// srv2 := reflect.New(reflect.TypeOf(rValue).Elem()).Interface().(serverInfo)
	// fmt.Println(srv2)

}
