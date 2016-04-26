package flaeg

import (
	"fmt"
	"testing"
)

//StructPtr : Struct with pointers
type StructPtr struct {
	PtrStruct1 *Struct1 `description:"Enable Struct1"`
	PtrStruct2 *Struct2 `description:"Enable Struct1"`
}

//Struct1 : trivial Struct
type Struct1 struct {
	S1Int    int    `description:"Struct 1 Int"`
	S1String string `description:"Struct 1 String"`
	S1Bool   bool   `description:"Struct 1 Int"`
}

//Struct2 : trivial Struct
type Struct2 struct {
	S2Int64  int64  `description:"Struct 2 Int64"`
	S2String string `description:"Struct 2 String"`
	S2Bool   bool   `description:"Struct 2 Int"`
}

func TestPtrLoad(t *testing.T) {
	defaultStructPtr := StructPtr{
		PtrStruct1: &Struct1{
			11,
			"oui",
			true,
		},
		PtrStruct2: &Struct2{
			42,
			"okay",
			false,
		},
	}
	config := StructPtr{
		PtrStruct1: &Struct1{
			44,
			"non",
			false,
		},
		PtrStruct2: nil,
	}
	args := []string{
		// "-h",
		"--ptrstruct1.s1string=quarantequatre",
		// "--ptrstruct1",
		"--ptrstruct2.s2string=toto",
	}
	if err := Load(&config, &defaultStructPtr, args); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	fmt.Printf("result PtrStruct1: %+v\n", config.PtrStruct1)
	fmt.Printf("result PtrStruct2: %+v\n", config.PtrStruct2)
}
