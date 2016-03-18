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
	Name         string    `long:"name" description:"overwrite owner name"`
	Organization string    `long:"org" description:"overwrite owner organisation"`
	Bio          string    `long:"bio" description:"overwrite owner biography"`
	Dob          time.Time `long:"dob" description:"overwrite owner date of birth"`
}
type databaseInfo struct {
	Server        string `long:"srv" description:"overwrite database server ip address"`
	ConnectionMax int    `long:"comax" description:"overwrite maximum number of connection on the database"`
	Enable        bool   `long:"ena" description:"overwrite database enable"`
}
type serverInfo struct {
	IP string `long:"ip" description:"overwrite server ip address"`
	Dc string `long:"dc" description:"overwrite server domain controller"`
}
type clientInfo struct {
	Data  []int        `long:"data" description:"overwrite clients data"`
	Hosts []serverInfo `description:"overwrite clients host names"`
}
type example struct {
	Title    string       `short:"t" description:"overwrite title"` //
	Owner    ownerInfo    `long:"own"  description:"overwrite server ip address"`
	Database databaseInfo ` description:"overwrite server ip address"`
	Servers  serverInfo   `description:"overwrite servers info --servers.[ip|dc] [srv name]: value"`
	Clients  *clientInfo  `long:"cli"  description:"overwrite server ip address"`
}

func TestGetTypesRecursive(t *testing.T) {
	//Test all
	var ex1 example
	namesmap := make(map[string]reflect.Type)
	if err := GetTypesRecursive(reflect.ValueOf(&ex1), namesmap, ""); err != nil {
		t.Errorf("Error %s", err.Error())
	}

	checkType := map[string]reflect.Type{
		"Title":          reflect.TypeOf(""),
		"own":            reflect.TypeOf(ownerInfo{}),
		"cli":            reflect.TypeOf(&clientInfo{}),
		"cli.Hosts.ip":   reflect.TypeOf(""),
		"t":              reflect.TypeOf(""),
		"Database":       reflect.TypeOf(databaseInfo{}),
		"cli.data":       reflect.TypeOf([]int{}),
		"cli.Hosts":      reflect.TypeOf([]serverInfo{}),
		"cli.Hosts.dc":   reflect.TypeOf(""),
		"own.name":       reflect.TypeOf(""),
		"own.bio":        reflect.TypeOf(""),
		"own.dob":        reflect.TypeOf(time.Time{}),
		"Database.srv":   reflect.TypeOf(""),
		"Database.comax": reflect.TypeOf(0),
		"Servers":        reflect.TypeOf(serverInfo{}),
		"own.org":        reflect.TypeOf(""),
		"Database.ena":   reflect.TypeOf(true),
		"Servers.ip":     reflect.TypeOf(""),
		"Servers.dc":     reflect.TypeOf(""),
	}
	for name, nameType := range namesmap {
		if checkType[name] != nameType {
			t.Fatalf("Tag : %s, got %s expected %s\n", name, nameType, checkType[name])
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
	//creating parsers
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
	var ex1 example
	tagsmap := make(map[string]reflect.Type)
	GetTypesRecursive(reflect.ValueOf(ex1), tagsmap, "")
	// fmt.Println(tagsmap)
	args := []string{
		"-own.org", "org",
		"-database.ena", //or +"=true"
		"-own.bio", "bio",
		"-database.comax", "123",
		"-database.srv", "srv",
		"-servers.ip", "ip",
		"-own.name", "name",
		"-servers.dc", "dc",
		"-cli.data", "{1,2,3,4}",
		"-t", "title",
		"-owner.dob", "1979-05-27T07:32:00Z",
	}
	pargs := ParseArgs(args, tagsmap, parsers)

	//CHECK
	myTime, _ := time.Parse(time.RFC3339, "1979-05-27T07:32:00Z")
	checkParse := map[string]interface{}{
		"own.org":        stringValue("org"),
		"database.ena":   boolValue(true),
		"own.bio":        stringValue("bio"),
		"database.comax": intValue(123),
		"database.srv":   stringValue("srv"),
		"servers.ip":     stringValue("ip"),
		"own.name":       stringValue("name"),
		"servers.dc":     stringValue("dc"),
		"cli.data":       customValue([]int{1, 2, 3, 4}),
		"t":              stringValue("title"),
		"own.dob":        timeValue(myTime),
	}
	for tag, inter := range pargs {
		v1 := reflect.ValueOf(checkParse[tag]).Interface()
		v2 := reflect.ValueOf(inter).Elem().Interface()
		if !reflect.DeepEqual(v1, v2) {
			t.Fatalf("Error tag %s : expected %+v got %+v", tag, v1, v2)
		}
	}
}

func TestFillStructRecursive(t *testing.T) {

	//creating parsers
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

	//Test
	var ex example
	tagsmap := make(map[string]reflect.Type)
	GetTypesRecursive(reflect.ValueOf(ex), tagsmap, "")
	args := []string{
		"-owner.org", "org",
		"-database.ena", //or +"=true"
		"-owner.bio", "bio",
		"-database.comax", "123",
		"-database.srv", "srv",
		"-servers.ip", "ip",
		"-owner.name", "name",
		"-servers.dc", "dc",
		"-clients.data", "{1,2,3,4}",
		"-t", "title",
		"-owner.dob", "1979-05-27T07:32:00Z",
	}
	pargs := ParseArgs(args, tagsmap, parsers)
	FillStructRecursive(reflect.ValueOf(&ex), pargs)

	//CHECK
	var check example
	check.Title = "title"
	check.Owner.Name = "name"
	check.Owner.Organization = "org"
	check.Owner.Bio = "bio"
	check.Owner.Dob, _ = time.Parse(time.RFC3339, "1979-05-27T07:32:00Z")
	check.Database.Server = "srv"
	check.Database.ConnectionMax = 123
	check.Database.Enable = true
	check.Servers.IP = "ip"
	check.Servers.Dc = "dc"
	check.Clients = &clientInfo{Data: []int{1, 2, 3, 4}}
	if !reflect.DeepEqual(ex, check) {
		t.Fatalf("expected\t: %+v\ngot\t\t: %+v", check, ex)
	}
}
