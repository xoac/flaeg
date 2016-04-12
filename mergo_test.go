package main

import (
	"fmt"
	"github.com/imdario/mergo"
	"testing"
	"time"
)

func TestMergo(t *testing.T) {
	fmt.Println("Hi there!")
	var struct1 example
	struct1.Title = "myTitle1"
	struct1.Owner.Name = "myOwnName1"
	struct1.Owner.Organization = "myOwnOrg1"
	struct1.Owner.Dob, _ = time.Parse(time.RFC3339, "1111-05-27T07:32:00Z")
	struct1.Database.Server = "mySrv1"
	struct1.Database.ConnectionMax = 1111
	struct1.Database.Enable = false
	struct1.Servers.Dc = "myServersDc"
	struct1.Clients = &clientInfo{Data: []int{1, 2, 3, 4}, Hosts: []ServerInfo{{"myIp1", "myDc1"}, {"myIp2", "myDc2"}}}

	var struct2 example

	struct2.Owner.Name = "myOwnName2"
	struct2.Owner.Bio = "myOwnBio2"
	struct2.Owner.Dob, _ = time.Parse(time.RFC3339, "1979-05-27T07:32:00Z")
	struct2.Database.Enable = true
	struct2.Servers.IP = "myServersIp2"
	struct2.Servers.Dc = "myServersDc"
	struct2.Clients = &clientInfo{Data: []int{1, 2, 3}, Hosts: []ServerInfo{}}

	//Fusion
	//struct2 overwrite struc1
	if err := mergo.Merge(&struct2, struct1); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	fmt.Printf("struct1 :%+v\nstruct2 :%+v\n", struct1, struct2)
	fmt.Printf("struct1 :%+v\nstruct2 :%+v\n", struct1.Clients, struct2.Clients)

}
