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

func dumpSchema(schema any, yes bool) {
	if !yes {
		return
	}

	data, err := json.Marshal(schema)
	if err != nil {
		fmt.Println("Error unmarshalling schema")
		return
	}

	fmt.Println(string(data))
}

func TestInfoGen(t *testing.T) {
	schema := generateSchema(reflect.TypeOf(SimpleMessage{}))
	dumpSchema(schema, false)

	object := schema.(SchemaObject)

	messageProp, ok := object.Properties["message"]

	if !ok {
		t.Error("message prop missing")
		return
	}

	message := messageProp.(SchemaField)

	if message.Type != "string" {
		t.Errorf("Message has wrong type, got %s, should have %s", message.Type, "string")
	}

	_ = message
}

func TestInfoGenComplex(t *testing.T) {
	schema := generateSchema(reflect.TypeOf(ComplexType{}))
	dumpSchema(schema, true)
}
