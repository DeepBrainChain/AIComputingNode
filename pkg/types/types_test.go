package types

import (
	"encoding/json"
	"fmt"
	"testing"
)

type JsonTest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (p Person) Say() {
	fmt.Println("I am Person struct")
}

type Student struct {
	Person
	Grade int `json:"grade"`
}

func (p Student) Say() {
	fmt.Println("I am Student struct")
}

func TestJsonStruct(t *testing.T) {
	test := JsonTest{
		Key:   "Test",
		Value: "value",
	}
	jsonData, err := json.Marshal(test)
	if err != nil {
		t.Fatalf("Marshal json %v", err)
	}
	t.Logf("orignal json %s", string(jsonData))

	var value string = "{\"key\":\"Test\",\"Value\":\"value\"}"
	js := JsonTest{}
	if err := json.Unmarshal([]byte(value), &js); err != nil {
		t.Fatalf("Unmarshal json %v", err)
	}
	t.Logf("Unmarshal json sucess %v", js)

	test2 := JsonTest{
		Key: "Test",
	}
	jsonData, err = json.Marshal(test2)
	if err != nil {
		t.Fatalf("Marshal json %v", err)
	}
	t.Logf("orignal json %s from struct %v", string(jsonData), test2)
}

func TestComposition(t *testing.T) {
	person := Person{
		Name: "Test",
		Age:  18,
	}
	person.Say()
	jsonData, err := json.Marshal(person)
	if err != nil {
		t.Fatalf("Marshal json %v", err)
	}
	t.Logf("orignal json %s", string(jsonData))

	student := Student{
		Person: Person{
			Name: "Test1",
			Age:  15,
		},
		Grade: 12,
	}
	student.Say()
	student.Person.Say()
	jsonData, err = json.Marshal(student)
	if err != nil {
		t.Fatalf("Marshal json %v", err)
	}
	t.Logf("orignal json %s", string(jsonData))
}
