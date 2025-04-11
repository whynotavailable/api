package rpc

import (
	"fmt"
	"net/http"
	"reflect"
)

// This will be used to generate and deal with function metadata.

type AnyMap = map[string]any

type FunctionInfo struct {
	Body any
}

func (conainer *RpcContainer) BuildMetadata() {
	for key, function := range conainer.functions {
		conainer.docs[key] = generateInfo(function)
	}
}

func generateInfo(function *RpcFunction) FunctionInfo {
	if function.bodyType == nil {
		return FunctionInfo{}
	}

	info := FunctionInfo{}

	info.Body = generateSchema(function.bodyType)

	return info
}

type SchemaObject struct {
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties"`
}

type SchemaMap struct {
	Type                 string `json:"type"`
	AdditionalProperties any    `json:"additionalProperties"`
}

type SchemaArray struct {
	Type  string `json:"type"`
	Items any    `json:"items"`
}

type SchemaField struct {
	Type string `json:"type"`
}

func generateSchema(elemType reflect.Type) any {
	if elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
	}

	if elemType.Kind() == reflect.Struct {
		schema := SchemaObject{
			Type:       "object",
			Properties: map[string]any{},
		}

		for i := range elemType.NumField() {
			propType := elemType.Field(i)
			setName, ok := propType.Tag.Lookup("json")
			if !ok {
				setName = propType.Name
			}
			schema.Properties[setName] = generateSchema(propType.Type)
		}

		return schema
	} else if elemType.Kind() == reflect.Map {
		return SchemaMap{
			Type:                 "object",
			AdditionalProperties: generateSchema(elemType.Elem()),
		}
	} else if elemType.Kind() == reflect.Slice {
		return SchemaArray{
			Type:  "array",
			Items: generateSchema(elemType.Elem()),
		}
	} else {
		return SchemaField{
			Type: translateKind(elemType.Kind().String()),
		}
	}
}

var kindMapping map[string]string = map[string]string{
	"interface": "object",
}

func translateKind(kind string) string {
	if mapping, ok := kindMapping[kind]; ok {
		return mapping
	}

	return kind
}

func (container *RpcContainer) ServeInfo(w http.ResponseWriter) {
	fmt.Fprint(w, "ok")
}
