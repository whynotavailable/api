package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
)

// This file is for the serve function, and that's pretty much it.

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

	var body any = nil

	if r.Body != nil {
		if f.setBody {
			data, err := io.ReadAll(r.Body)
			if err != nil {
				ErrHandler(w, err)
				return
			}

			newObject := reflect.New(f.bodyType)
			bodyInterface := newObject.Interface()

			err = json.Unmarshal(data, bodyInterface)
			if err != nil {
				Error{
					Err:  err,
					Code: http.StatusBadRequest,
				}.Write(w)
				return
			}

			// Elem extracts the value from the pointer.
			body = newObject.Elem().Interface()

			r.Body.Close()
			r.Body = nil
		}
	} else {
		if f.bodyType != nil {
			Error{
				Err:  errors.New("body required"),
				Code: http.StatusBadRequest,
			}.Write(w)
			return
		}
	}

	request := RpcRequest{
		Ctx:     context.Background(),
		Headers: r.Header,
		RawBody: r.Body,
		Body:    body,
		Writer:  w,
	}

	for _, middleware := range container.middlewars {
		err := middleware(&request)
		if err != nil {
			Error{
				Err: err,
			}.Write(w)
			return
		}
	}

	response := f.Handler(&request)

	if response != nil {
		response.Write(w)
	}
}
