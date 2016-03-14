package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
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
	Data  [][]interface{} `long:"clients.data" description:"overwrite clients data"` //
	Hosts []serverInfo    `group:"clients.hosts" description:"overwrite clients host names"`
}
type example struct {
	Title    string                `short:"t" description:"overwrite title"` //
	Owner    ownerInfo             `group:"Owner info"`
	Database databaseInfo          `group:"Database info"`
	Servers  map[string]serverInfo `group:"Servers" description:"overwrite servers info --servers.[ip|dc] [srv name]: value"`
	Clients  *clientInfo           `group:"Clients"`
	Pouet    interface{}           `short:"z" description:"pouet"`
}

//Test function ReflectRecursive
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

func TestGetTagsRecursive(t *testing.T) {
	//Test struct, slice, string
	// fmt.Println("--------------Test struct, slice, string--------------------")
	//var cl1 clientInfo
	// var sinf1 serverInfo
	// var sinf2 serverInfo
	// sinf1.Dc = "dc1"
	// sinf1.IP = "ip1"
	// sinf2.Dc = "dc2"
	// sinf2.IP = "ip2"
	// cl1.Hosts = []serverInfo{sinf1, sinf2}
	//fmt.Println(GetTagsRecursive(reflect.ValueOf(cl1)))
	// if tags := GetTagsRecursive(reflect.ValueOf(cl1)); !reflect.DeepEqual(tags["--clients.data"].Interface(), reflect.ValueOf(cl1.Data).Interface()) || !reflect.DeepEqual(tags["--clients.hosts"].Interface(), []string{"un", "deux", "trois"}) {
	// 	fmt.Printf("%+v\n%+v\n%+v\n%+v\n", tags["--clients.data"], reflect.ValueOf(cl1.Data), tags["--clients.hosts"], reflect.ValueOf(cl1.Hosts))
	// 	fmt.Printf("%+v\n", tags)
	// 	t.Fail()
	// }

	//Test all
	// 	fmt.Println("------------------Test all------------------------------------")
	var ex1 example
	// 	ex1.init()
	tagsmap := make(map[string]reflect.Type)
	GetTagsRecursive(reflect.ValueOf(ex1), tagsmap)
	fmt.Printf("%+v\n", tagsmap)
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

func TestParseArgs(t *testing.T) {
	os.Args = append(os.Args, "-servers.dc=toto", "-servers.ip=tztz")
	fmt.Printf("ARGS : %+v\n", os.Args)
	var srv1 serverInfo
	parsers := map[reflect.Type]flag.Value{}
	var myStringParser parserString
	parsers[reflect.TypeOf("reflect.String")] = &myStringParser

	tagsmap := make(map[string]reflect.Type)
	GetTagsRecursive(reflect.ValueOf(srv1), tagsmap)
	pargs := ParseArgs(tagsmap, parsers)

	fmt.Printf("parsers : %+v\n", pargs)
}
