package rpc

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

// This here is for testing the non exposed methods of info generation
// In particular, the schema generation

type ComplexType struct {
	Messages []SimpleMessage
	Map      map[string]any
	Name     string
	Page     *int
}

func TestInfoGen(t *testing.T) {
	schema := generateSchema(reflect.TypeOf(SimpleMessage{}))
	data, err := json.Marshal(schema)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(data))

	object := schema.(SchemaObject)
	if _, ok := object.Properties["message"]; !ok {
		t.Error("message prop missing")
		return
	}
}

func TestInfoGenComplex(t *testing.T) {
	schema := generateSchema(reflect.TypeOf(ComplexType{}))
	data, err := json.Marshal(schema)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(data))
	_ = data
}
