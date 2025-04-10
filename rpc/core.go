// rpc can be used to simply manage RPC functions.
//
// At some point, there'll be an endpoint that will return a json document with instructions on how to
// call the endpoints
package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
)

// functions
type (
	RpcMiddleware      func(*RpcRequest) error
	RpcFunctionHandler func(*RpcRequest) (RpcResponse, error)
)

type RpcRequest struct {
	Ctx     context.Context
	Headers map[string][]string
	RawBody []byte
	Body    any
}

type RpcResponse struct {
	Code int
	Body []byte
}

func NewRpcResponse(data []byte) RpcResponse {
	return RpcResponse{
		Code: http.StatusOK,
		Body: data,
	}
}

func ErrorResponseStatus(err error, code int) RpcResponse {
	return RpcResponse{
		Code: code,
		Body: []byte(err.Error()),
	}
}

func JsonResponse(data any) (RpcResponse, error) {
	rawData, err := json.Marshal(data)
	if err != nil {
		return RpcResponse{}, err
	}

	return RpcResponse{
		Code: http.StatusOK,
		Body: rawData,
	}, nil
}

func (response *RpcResponse) Write(w http.ResponseWriter) {
	w.WriteHeader(response.Code)
	w.Write(response.Body)
}

type RpcFunction struct {
	Handler  RpcFunctionHandler
	bodyType reflect.Type
}

func NewRpcFunction(f RpcFunctionHandler) RpcFunction {
	return RpcFunction{
		Handler: f,
	}
}

func (function *RpcFunction) SetBodyType(t reflect.Type) *RpcFunction {
	function.bodyType = t

	return function
}

type RpcContainer struct {
	functions  map[string]*RpcFunction
	docs       map[string]any
	middlewars []RpcMiddleware
}

func NewRpcContainer() RpcContainer {
	return RpcContainer{
		functions:  map[string]*RpcFunction{},
		docs:       map[string]any{},
		middlewars: []RpcMiddleware{},
	}
}

func ErrHandler(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(w, err.Error())
}

func (conainer *RpcContainer) BuildDocs() {
	for key := range conainer.functions {
		conainer.docs[key] = "hi"
	}
}

func (container *RpcContainer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if r.URL.Path == "/_info" {
			container.ServeInfo(w)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "Not Found")
		}
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad Request")
		return
	}

	functionKey := strings.TrimLeft(r.URL.Path, "/")

	f, ok := container.functions[functionKey]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Function '%s' Not Found", functionKey)
		return
	}

	var data []byte = nil
	var body any = nil

	if r.Body != nil {
		_data, err := io.ReadAll(r.Body)
		if err != nil {
			ErrHandler(w, err)
			return
		}
		data = _data // This is weird because go is weird

		// INFO: This is kind of weird because of how the unmarshaller works.
		// newObject is a container for both the value and the interface.
		// Pulling the pointer, and changing the data underneath persists to the container
		if f.bodyType != nil {
			newObject := reflect.New(f.bodyType)
			bodyInterface := newObject.Interface()

			err := json.Unmarshal(data, bodyInterface)
			if err != nil {
				errorResponse := ErrorResponseStatus(err, http.StatusBadRequest)
				errorResponse.Write(w)
				return
			}

			// Elem extracts the value from the pointer.
			body = newObject.Elem().Interface()
		}
	}

	request := RpcRequest{
		Ctx:     context.Background(),
		Headers: r.Header,
		RawBody: data,
		Body:    body,
	}

	for _, middleware := range container.middlewars {
		err := middleware(&request)
		if err != nil {
			errorResponse := ErrorResponseStatus(err, http.StatusInternalServerError)
			errorResponse.Write(w)
			return
		}
	}

	response, err := f.Handler(&request)
	if err != nil {
		// Global error handler
		// TODO: Make this customizable
		response = ErrorResponseStatus(err, http.StatusInternalServerError)
	}

	response.Write(w)
}

func (container *RpcContainer) ServeInfo(w http.ResponseWriter) {
	fmt.Fprint(w, "ok")
}

func (container *RpcContainer) SetupMux(mux *http.ServeMux, prefix string) {
	container.BuildDocs()
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
