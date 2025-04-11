// rpc can be used to simply manage RPC functions.
//
// At some point, there'll be an endpoint that will return a json document with instructions on how to
// call the endpoints
package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

// functions
type (
	RpcMiddleware      func(*RpcRequest) error
	RpcFunctionHandler func(*RpcRequest) (RpcResponse, error)
)

type RpcRequest struct {
	Ctx     context.Context
	Headers map[string][]string
	// The raw bytes from the body. Will be null if no body is sent.
	RawBody []byte
	// Body converted to appropriate type, will only be done if body type is provided.
	// Requires casting.
	Body any
}

type RpcResponse interface {
	Write(w http.ResponseWriter)
}

type Error struct {
	Err  error
	Code int
}

func (e Error) Write(w http.ResponseWriter) {
	if e.Code == 0 {
		e.Code = http.StatusInternalServerError
	}

	w.WriteHeader(e.Code)
	fmt.Fprint(w, e.Err)
}

type Json struct {
	Body any
	Code int
}

func (e Json) Write(w http.ResponseWriter) {
	if e.Code == 0 {
		e.Code = http.StatusOK
	}

	if e.Body == nil {
		e.Body = map[string]string{}
	}

	data, err := json.Marshal(e.Body)
	if err != nil {
		newError := Error{
			Err:  err,
			Code: http.StatusBadRequest,
		}
		newError.Write(w)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(e.Code)
	w.Write(data)
}

// RpcFunction is the configuration object for functions.
// It uses a fluent api. Most methods take a pointer, mutate, and return the same pointer.
type RpcFunction struct {
	Handler  RpcFunctionHandler
	bodyType reflect.Type
}

// Set the body type. If set, calls without a body, or that do not deserialize properly will return a bad request.
func (function *RpcFunction) SetBodyType(t reflect.Type) *RpcFunction {
	function.bodyType = t

	return function
}

type RpcContainer struct {
	functions  map[string]*RpcFunction
	docs       map[string]FunctionInfo
	middlewars []RpcMiddleware
}

func NewRpcContainer() RpcContainer {
	return RpcContainer{
		functions:  map[string]*RpcFunction{},
		docs:       map[string]FunctionInfo{},
		middlewars: []RpcMiddleware{},
	}
}

func ErrHandler(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(w, err.Error())
}

func (container *RpcContainer) SetupMux(mux *http.ServeMux, prefix string) {
	container.BuildMetadata()
	mux.Handle(fmt.Sprintf("%s/", prefix), http.StripPrefix(prefix, container))
}

func (container *RpcContainer) AddFunction(key string, handler RpcFunctionHandler) *RpcFunction {
	function := RpcFunction{
		Handler: handler,
	}
	container.functions[key] = &function

	return &function
}

func (container *RpcContainer) AddMiddleware(middleware RpcMiddleware) {
	container.middlewars = append(container.middlewars, middleware)
}

// Primarily for unit tests
type SimpleMessage struct {
	Message string `json:"message"`
}
