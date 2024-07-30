package types

import (
	"encoding/json"
	"fmt"
	"testing"
)

type testJson struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	EValue string `json:"evalue,omitempty"`
}

type testPerson struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (p testPerson) Say() {
	fmt.Println("I am Person struct")
}

type testStudent struct {
	testPerson
	Grade int `json:"grade"`
}

func (p testStudent) Say() {
	fmt.Println("I am Student struct")
}

func TestJsonStruct(t *testing.T) {
	test := testJson{
		Key:   "Test",
		Value: "value",
	}
	jsonData, err := json.Marshal(test)
	if err != nil {
		t.Fatalf("Marshal json %v", err)
	}
	t.Logf("orignal json %s", string(jsonData))

	var value string = "{\"key\":\"Test\",\"Value\":\"value\"}"
	js := testJson{}
	if err := json.Unmarshal([]byte(value), &js); err != nil {
		t.Fatalf("Unmarshal json %v", err)
	}
	t.Logf("Unmarshal json sucess %v", js)

	test2 := testJson{
		Key: "Test",
	}
	jsonData, err = json.Marshal(test2)
	if err != nil {
		t.Fatalf("Marshal json %v", err)
	}
	t.Logf("orignal json %s from struct %v", string(jsonData), test2)
}

func TestComposition(t *testing.T) {
	person := testPerson{
		Name: "Test",
		Age:  18,
	}
	person.Say()
	jsonData, err := json.Marshal(person)
	if err != nil {
		t.Fatalf("Marshal json %v", err)
	}
	t.Logf("orignal json %s", string(jsonData))

	student := testStudent{
		testPerson: testPerson{
			Name: "Test1",
			Age:  15,
		},
		Grade: 12,
	}
	student.Say()
	student.testPerson.Say()
	jsonData, err = json.Marshal(student)
	if err != nil {
		t.Fatalf("Marshal json %v", err)
	}
	t.Logf("orignal json %s", string(jsonData))
}

func TestNodeType(t *testing.T) {
	var nt NodeType = 0x00
	t.Logf("%v ~ {%v, %v, %v}", nt, nt.IsPublicNode(), nt.IsClientNode(), nt.IsModelNode())

	nt |= PublicIpFlag
	t.Logf("%v ~ {%v, %v, %v}", nt, nt.IsPublicNode(), nt.IsClientNode(), nt.IsModelNode())

	nt |= PeersCollectFlag
	t.Logf("%v ~ {%v, %v, %v}", nt, nt.IsPublicNode(), nt.IsClientNode(), nt.IsModelNode())

	nt |= ModelFlag
	t.Logf("%v ~ {%v, %v, %v}", nt, nt.IsPublicNode(), nt.IsClientNode(), nt.IsModelNode())

	nt &= ^ModelFlag
	t.Logf("%v ~ {%v, %v, %v}", nt, nt.IsPublicNode(), nt.IsClientNode(), nt.IsModelNode())

	nt &= ^PeersCollectFlag
	t.Logf("%v ~ {%v, %v, %v}", nt, nt.IsPublicNode(), nt.IsClientNode(), nt.IsModelNode())

	nt &= ^PublicIpFlag
	t.Logf("%v ~ {%v, %v, %v}", nt, nt.IsPublicNode(), nt.IsClientNode(), nt.IsModelNode())
}
