package flaeg

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

//Configuration is a struct which contains all differents type to field
//using parsers on string, time.Duration, pointer, bool, int, int64, time.Time, float64
type Configuration struct {
	Name     string        //no description struct tag, it will not be flaged
	LogLevel string        `short:"l" description:"Log level"`      //string type field, short flag "-l"
	Timeout  time.Duration `description:"Timeout duration"`         //time.Duration type field
	Db       *DatabaseInfo `description:"Enable database"`          //pointer type field (on DatabaseInfo)
	Owner    *OwnerInfo    `description:"Enable Owner description"` //another pointer type field (on OwnerInfo)
}

type ServerInfo struct {
	Watch  bool   `description:"Watch device"`      //bool type
	IP     string `description:"Server ip address"` //string type field
	Load   int    `description:"Server load"`       //int type field
	Load64 int64  `description:"Server load"`       //int64 type field, same description just to be sure it works
}
type DatabaseInfo struct {
	ServerInfo             //Go throught annonymous field
	ConnectionMax   uint   `long:"comax" description:"Number max of connections on database"` //uint type field, long flag "--comax"
	ConnectionMax64 uint64 `description:"Number max of connections on database"`              //uint64 type field, same description just to be sure it works
}
type OwnerInfo struct {
	Name        *string      `description:"Owner name"`                     //pointer type field on string
	DateOfBirth time.Time    `long:"dob" description:"Owner date of birth"` //time.Time type field, long flag "--dob"
	Rate        float64      `description:"Owner rate"`                     //float64 type field
	Servers     []ServerInfo `description:"Owner Server"`                   //slice of ServerInfo type field, need a custom parser
}

//newDefaultConfiguration returns a pointer on Configuration with default values
func newDefaultConfiguration() *Configuration {
	var db DatabaseInfo
	db.Watch = true
	db.IP = "192.168.1.2"
	db.Load = 32
	db.Load64 = 64
	db.ConnectionMax = 3200000000            //max 4294967295
	db.ConnectionMax64 = 6400000000000000000 //max 18446744073709551615

	var own OwnerInfo
	str := "DefaultOwnerNamePointer"
	own.Name = &str
	own.DateOfBirth, _ = time.Parse(time.RFC3339, "1979-05-27T07:32:00Z")
	own.Rate = 0.111
	own.Servers = []ServerInfo{
		ServerInfo{IP: "192.168.1.2"},
		ServerInfo{IP: "192.168.1.3"},
		ServerInfo{IP: "192.168.1.4"},
	}
	return &Configuration{
		Name:     "defaultName",
		LogLevel: "ERROR",
		Timeout:  time.Millisecond,
		Db:       &db,
		Owner:    &own,
	}
}

//newConfiguration returns a pointer on Configuration initialized
func newConfiguration() *Configuration {
	var own OwnerInfo
	str := "InitOwnerNamePointer"
	own.Name = &str
	own.DateOfBirth, _ = time.Parse(time.RFC3339, "1993-09-12T07:32:00Z")
	own.Rate = 0.999
	return &Configuration{
		Name:     "initName",
		LogLevel: "DEBUG",
		Timeout:  time.Second,
		Owner:    &own,
	}
}

func TestGetTypesRecursive(t *testing.T) {
	config := newConfiguration()
	flagmap := make(map[string]StructField)
	if err := getTypesRecursive(reflect.ValueOf(config), flagmap, ""); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	// Check only type
	checkType := map[string]reflect.Type{
		"loglevel":           reflect.TypeOf(""),
		"timeout":            reflect.TypeOf(time.Second),
		"db":                 reflect.TypeOf(true),
		"db.watch":           reflect.TypeOf(true),
		"db.ip":              reflect.TypeOf(""),
		"db.load":            reflect.TypeOf(0),
		"db.load64":          reflect.TypeOf(int64(0)),
		"db.comax":           reflect.TypeOf(uint(0)),
		"db.connectionmax64": reflect.TypeOf(uint64(0)),
		"owner":              reflect.TypeOf(true),
		"owner.name":         reflect.TypeOf(true),
		"owner.dob":          reflect.TypeOf(time.Now()),
		"owner.rate":         reflect.TypeOf(float64(1.1)),
		"owner.servers":      reflect.TypeOf([]ServerInfo{}),
	}
	for name, field := range flagmap {
		// fmt.Printf("%s : %+v\n", name, field)
		if checkType[name] != field.Type {
			t.Fatalf("Tag : %s, got %s expected %s\n", name, field.Type, checkType[name])
		}
	}
}

//CUSTOM PARSER
// -- sliceServerValue format {IP,DC}
type sliceServerValue []ServerInfo

func (c *sliceServerValue) Set(s string) error {
	//could use RegExp
	srv := ServerInfo{IP: s}
	*c = append(*c, srv)
	return nil
}

func (c *sliceServerValue) Get() interface{} { return []ServerInfo(*c) }

func (c *sliceServerValue) String() string { return fmt.Sprintf("%v", *c) }

func (c *sliceServerValue) SetValue(val interface{}) {
	*c = sliceServerValue(val.([]ServerInfo))
}

func TestLoadParsers(t *testing.T) {
	//inti customParsers
	customParsers := map[reflect.Type]Parser{
		reflect.TypeOf([]ServerInfo{}): &sliceServerValue{},
	}
	//test
	parsers, err := loadParsers(customParsers)
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//Check
	check := map[reflect.Type]Parser{
		reflect.TypeOf([]ServerInfo{}): &sliceServerValue{},
	}
	var boolParser boolValue
	check[reflect.TypeOf(true)] = &boolParser
	var intParser intValue
	check[reflect.TypeOf(1)] = &intParser
	var int64Parser int64Value
	check[reflect.TypeOf(int64(1))] = &int64Parser
	var uintParser uintValue
	check[reflect.TypeOf(uint(1))] = &uintParser
	var uint64Parser uint64Value
	check[reflect.TypeOf(uint64(1))] = &uint64Parser
	var stringParser stringValue
	check[reflect.TypeOf("")] = &stringParser
	var float64Parser float64Value
	check[reflect.TypeOf(float64(1.5))] = &float64Parser
	var durationParser durationValue
	check[reflect.TypeOf(time.Second)] = &durationParser
	var timeParser timeValue
	check[reflect.TypeOf(time.Now())] = &timeParser

	for typ, parser := range parsers {
		if !reflect.DeepEqual(parser, check[typ]) {
			t.Fatalf("Got %s expected %s\n", parser, check[typ])
		}
	}
}

//Test ParseArgs with trivial flags (ie not short, not on custom parser, not on pointer)
func TestParseArgsTrivialFlags(t *testing.T) {
	//We assume that getTypesRecursive works well
	config := newConfiguration()
	flagmap := make(map[string]StructField)
	if err := getTypesRecursive(reflect.ValueOf(config), flagmap, ""); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//init parsers
	parsers := map[reflect.Type]Parser{
		reflect.TypeOf([]ServerInfo{}): &sliceServerValue{},
	}
	var boolParser boolValue
	parsers[reflect.TypeOf(true)] = &boolParser
	var intParser intValue
	parsers[reflect.TypeOf(1)] = &intParser
	var int64Parser int64Value
	parsers[reflect.TypeOf(int64(1))] = &int64Parser
	var uintParser uintValue
	parsers[reflect.TypeOf(uint(1))] = &uintParser
	var uint64Parser uint64Value
	parsers[reflect.TypeOf(uint64(1))] = &uint64Parser
	var stringParser stringValue
	parsers[reflect.TypeOf("")] = &stringParser
	var float64Parser float64Value
	parsers[reflect.TypeOf(float64(1.5))] = &float64Parser
	var durationParser durationValue
	parsers[reflect.TypeOf(time.Second)] = &durationParser
	var timeParser timeValue
	parsers[reflect.TypeOf(time.Now())] = &timeParser
	//init args
	args := []string{
		"--loglevel=OFF",
		"--timeout=9ms",
	}
	//test
	valmap, err := parseArgs(args, flagmap, parsers)
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//check
	check := map[string]Parser{}
	stringParser.SetValue("OFF")
	check["loglevel"] = &stringParser
	durationParser.SetValue(9 * time.Millisecond)
	check["timeout"] = &durationParser

	for flag, parser := range valmap {
		if !reflect.DeepEqual(parser, check[flag]) {
			t.Fatalf("Got %s expected %s\n", parser, check[flag])
		}
	}
}

//Test ParseArgs with short flags
func TestParseArgsShortFlags(t *testing.T) {
	//We assume that getTypesRecursive works well
	config := newConfiguration()
	flagmap := make(map[string]StructField)
	if err := getTypesRecursive(reflect.ValueOf(config), flagmap, ""); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//init parsers
	parsers := map[reflect.Type]Parser{
		reflect.TypeOf([]ServerInfo{}): &sliceServerValue{},
	}
	var boolParser boolValue
	parsers[reflect.TypeOf(true)] = &boolParser
	var intParser intValue
	parsers[reflect.TypeOf(1)] = &intParser
	var int64Parser int64Value
	parsers[reflect.TypeOf(int64(1))] = &int64Parser
	var uintParser uintValue
	parsers[reflect.TypeOf(uint(1))] = &uintParser
	var uint64Parser uint64Value
	parsers[reflect.TypeOf(uint64(1))] = &uint64Parser
	var stringParser stringValue
	parsers[reflect.TypeOf("")] = &stringParser
	var float64Parser float64Value
	parsers[reflect.TypeOf(float64(1.5))] = &float64Parser
	var durationParser durationValue
	parsers[reflect.TypeOf(time.Second)] = &durationParser
	var timeParser timeValue
	parsers[reflect.TypeOf(time.Now())] = &timeParser
	//init args
	args := []string{
		"-lWARN",
	}
	//test
	valmap, err := parseArgs(args, flagmap, parsers)
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//check
	check := map[string]Parser{}
	stringParser.Set("WARN")
	check["loglevel"] = &stringParser

	for flag, parser := range valmap {
		if !reflect.DeepEqual(parser, check[flag]) {
			t.Fatalf("Got %s expected %s\n", parser, check[flag])
		}
	}
}

//Test ParseArgs call Flag on pointers
func TestParseArgsPointerFlag(t *testing.T) {
	//We assume that getTypesRecursive works well
	config := newConfiguration()
	flagmap := make(map[string]StructField)
	if err := getTypesRecursive(reflect.ValueOf(config), flagmap, ""); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//init parsers
	parsers := map[reflect.Type]Parser{
		reflect.TypeOf([]ServerInfo{}): &sliceServerValue{},
	}
	var boolParser boolValue
	parsers[reflect.TypeOf(true)] = &boolParser
	var intParser intValue
	parsers[reflect.TypeOf(1)] = &intParser
	var int64Parser int64Value
	parsers[reflect.TypeOf(int64(1))] = &int64Parser
	var uintParser uintValue
	parsers[reflect.TypeOf(uint(1))] = &uintParser
	var uint64Parser uint64Value
	parsers[reflect.TypeOf(uint64(1))] = &uint64Parser
	var stringParser stringValue
	parsers[reflect.TypeOf("")] = &stringParser
	var float64Parser float64Value
	parsers[reflect.TypeOf(float64(1.5))] = &float64Parser
	var durationParser durationValue
	parsers[reflect.TypeOf(time.Second)] = &durationParser
	var timeParser timeValue
	parsers[reflect.TypeOf(time.Now())] = &timeParser
	//init args
	args := []string{
		"--db",
		"--owner",
	}
	//test
	valmap, err := parseArgs(args, flagmap, parsers)
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//check
	check := map[string]Parser{}
	checkDb := boolValue(true)
	check["db"] = &checkDb
	checkOwner := boolValue(true)
	check["owner"] = &checkOwner

	for flag, parser := range valmap {
		if !reflect.DeepEqual(parser, check[flag]) {
			t.Fatalf("Got %s expected %s\n", parser, check[flag])
		}
	}
}

//Test ParseArgs with flags under a pointer and a long flag
func TestParseArgsUnderPointerFlag(t *testing.T) {
	//We assume that getTypesRecursive works well
	config := newConfiguration()
	flagmap := make(map[string]StructField)
	if err := getTypesRecursive(reflect.ValueOf(config), flagmap, ""); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//init parsers
	parsers := map[reflect.Type]Parser{
		reflect.TypeOf([]ServerInfo{}): &sliceServerValue{},
	}
	var boolParser boolValue
	parsers[reflect.TypeOf(true)] = &boolParser
	var intParser intValue
	parsers[reflect.TypeOf(1)] = &intParser
	var int64Parser int64Value
	parsers[reflect.TypeOf(int64(1))] = &int64Parser
	var uintParser uintValue
	parsers[reflect.TypeOf(uint(1))] = &uintParser
	var uint64Parser uint64Value
	parsers[reflect.TypeOf(uint64(1))] = &uint64Parser
	var stringParser stringValue
	parsers[reflect.TypeOf("")] = &stringParser
	var float64Parser float64Value
	parsers[reflect.TypeOf(float64(1.5))] = &float64Parser
	var durationParser durationValue
	parsers[reflect.TypeOf(time.Second)] = &durationParser
	var timeParser timeValue
	parsers[reflect.TypeOf(time.Now())] = &timeParser
	//init args
	args := []string{
		"--owner.name",
		"--db.comax=5000000000",
	}
	//test
	valmap, err := parseArgs(args, flagmap, parsers)
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//check
	check := map[string]Parser{}
	boolParser.SetValue(true)
	check["owner.name"] = &boolParser
	uintParser.SetValue(uint(5000000000))
	check["db.comax"] = &uintParser
	for flag, parser := range valmap {
		if !reflect.DeepEqual(parser, check[flag]) {
			t.Fatalf("Got %s expected %s\n", parser, check[flag])
		}
	}
}

//Test ParseArgs with flag on pointer and flag under a pointer together
func TestParseArgsPointerFlagUnderPointerFlag(t *testing.T) {
	//We assume that getTypesRecursive works well
	config := newConfiguration()
	flagmap := make(map[string]StructField)
	if err := getTypesRecursive(reflect.ValueOf(config), flagmap, ""); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//init parsers
	parsers := map[reflect.Type]Parser{
		reflect.TypeOf([]ServerInfo{}): &sliceServerValue{},
	}
	var boolParser boolValue
	parsers[reflect.TypeOf(true)] = &boolParser
	var intParser intValue
	parsers[reflect.TypeOf(1)] = &intParser
	var int64Parser int64Value
	parsers[reflect.TypeOf(int64(1))] = &int64Parser
	var uintParser uintValue
	parsers[reflect.TypeOf(uint(1))] = &uintParser
	var uint64Parser uint64Value
	parsers[reflect.TypeOf(uint64(1))] = &uint64Parser
	var stringParser stringValue
	parsers[reflect.TypeOf("")] = &stringParser
	var float64Parser float64Value
	parsers[reflect.TypeOf(float64(1.5))] = &float64Parser
	var durationParser durationValue
	parsers[reflect.TypeOf(time.Second)] = &durationParser
	var timeParser timeValue
	parsers[reflect.TypeOf(time.Now())] = &timeParser
	//init args
	args := []string{
		"--db",
		"--db.watch",
		"--db.connectionmax64=900",
	}
	//test
	valmap, err := parseArgs(args, flagmap, parsers)
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//check
	check := map[string]Parser{}
	boolParser.SetValue(true)
	check["db"] = &boolParser
	uint64Parser.SetValue(uint64(900))
	check["db.connectionmax64"] = &uint64Parser
	check["db.watch"] = &boolParser
	for flag, parser := range valmap {
		if !reflect.DeepEqual(parser, check[flag]) {
			t.Fatalf("Got %s expected %s\n", parser, check[flag])
		}
	}
}

//Test ParseArgs call Flag with custom parsers
func TestParseArgsCustomFlag(t *testing.T) {
	//We assume that getTypesRecursive works well
	config := newConfiguration()
	flagmap := make(map[string]StructField)
	if err := getTypesRecursive(reflect.ValueOf(config), flagmap, ""); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//init parsers
	parsers := map[reflect.Type]Parser{
		reflect.TypeOf([]ServerInfo{}): &sliceServerValue{},
	}
	var boolParser boolValue
	parsers[reflect.TypeOf(true)] = &boolParser
	var intParser intValue
	parsers[reflect.TypeOf(1)] = &intParser
	var int64Parser int64Value
	parsers[reflect.TypeOf(int64(1))] = &int64Parser
	var uintParser uintValue
	parsers[reflect.TypeOf(uint(1))] = &uintParser
	var uint64Parser uint64Value
	parsers[reflect.TypeOf(uint64(1))] = &uint64Parser
	var stringParser stringValue
	parsers[reflect.TypeOf("")] = &stringParser
	var float64Parser float64Value
	parsers[reflect.TypeOf(float64(1.5))] = &float64Parser
	var durationParser durationValue
	parsers[reflect.TypeOf(time.Second)] = &durationParser
	var timeParser timeValue
	parsers[reflect.TypeOf(time.Now())] = &timeParser
	//init args
	args := []string{
		"--owner.servers=127.0.0.1",
		"--owner.servers=1.0.0.1",
	}
	//test
	valmap, err := parseArgs(args, flagmap, parsers)
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//check
	check := map[string]Parser{}
	checkOwnerServers := sliceServerValue{
		ServerInfo{IP: "127.0.0.1"},
		ServerInfo{IP: "1.0.0.1"},
	}
	check["owner.servers"] = &checkOwnerServers

	for flag, parser := range valmap {
		if !reflect.DeepEqual(parser, check[flag]) {
			t.Fatalf("Got %s expected %s\n", parser, check[flag])
		}
	}
}

//Test ParseArgs with all flags possible with custom parsers
func TestParseArgsAll(t *testing.T) {
	//We assume that getTypesRecursive works well
	config := newConfiguration()
	flagmap := make(map[string]StructField)
	if err := getTypesRecursive(reflect.ValueOf(config), flagmap, ""); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//init parsers
	parsers := map[reflect.Type]Parser{
		reflect.TypeOf([]ServerInfo{}): &sliceServerValue{},
	}
	var boolParser boolValue
	parsers[reflect.TypeOf(true)] = &boolParser
	var intParser intValue
	parsers[reflect.TypeOf(1)] = &intParser
	var int64Parser int64Value
	parsers[reflect.TypeOf(int64(1))] = &int64Parser
	var uintParser uintValue
	parsers[reflect.TypeOf(uint(1))] = &uintParser
	var uint64Parser uint64Value
	parsers[reflect.TypeOf(uint64(1))] = &uint64Parser
	var stringParser stringValue
	parsers[reflect.TypeOf("")] = &stringParser
	var float64Parser float64Value
	parsers[reflect.TypeOf(float64(1.5))] = &float64Parser
	var durationParser durationValue
	parsers[reflect.TypeOf(time.Second)] = &durationParser
	var timeParser timeValue
	parsers[reflect.TypeOf(time.Now())] = &timeParser
	//init args
	args := []string{
		"--loglevel=INFO",
		"--timeout=1s",
		"--db",
		"--db.watch",
		"--db.ip=192.168.0.1",
		"--db.load=-1",
		"--db.load64=-164",
		"--db.comax=2",
		"--db.connectionmax64=264",
		"--owner",
		"--owner.name",
		"--owner.dob=2016-04-20T17:39:00Z",
		"--owner.rate=0.222",
		"--owner.servers=1.0.0.1",
	}
	//test
	valmap, err := parseArgs(args, flagmap, parsers)
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//check
	check := map[string]Parser{}
	stringParser.SetValue("INFO")
	check["loglevel"] = &stringParser
	durationParser.SetValue(time.Second)
	check["timeout"] = &durationParser
	boolParser.SetValue(true)
	check["db"] = &boolParser
	check["db.watch"] = &boolParser
	checkDcIP := stringValue("192.168.0.1")
	check["db.ip"] = &checkDcIP
	intParser.SetValue(-1)
	check["db.load"] = &intParser
	int64Parser.SetValue(int64(-164))
	check["db.load64"] = &int64Parser
	uintParser.SetValue(uint(2))
	check["db.comax"] = &uintParser
	uint64Parser.SetValue(uint64(264))
	check["db.connectionmax64"] = &uint64Parser
	check["owner"] = &boolParser
	check["owner.name"] = &boolParser
	timeParser.Set("2016-04-20T17:39:00Z")
	check["owner.dob"] = &timeParser
	float64Parser.SetValue(0.222)
	check["owner.rate"] = &float64Parser
	checkOwnerServers := sliceServerValue{
		ServerInfo{IP: "1.0.0.1"},
	}
	check["owner.servers"] = &checkOwnerServers

	for flag, parser := range valmap {
		if !reflect.DeepEqual(parser, check[flag]) {
			t.Fatalf("Got %s expected %s\n", parser, check[flag])
		}
	}
}
