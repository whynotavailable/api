package rpc_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/whynotavailable/api/rpc"
	"github.com/whynotavailable/api/utils"
)

func TestSimpleHandler(t *testing.T) {
	messageText := "hi"
	rpcContainer := rpc.NewRpcContainer()

	rpcContainer.AddFunction("get-hello", func(r *rpc.RpcRequest) rpc.RpcResponse {
		body := r.Body.(rpc.SimpleMessage)

		return rpc.Json{
			Body: body,
		}
	}).SetBodyType(reflect.TypeOf(rpc.SimpleMessage{}))

	bodyBytes, _ := json.Marshal(rpc.SimpleMessage{
		Message: messageText,
	})

	r, err := http.NewRequest(http.MethodPost, "/get-hello", bytes.NewBuffer(bodyBytes))
	if err != nil {
		t.Error(err)
		return
	}

	rr := httptest.NewRecorder()

	rpcContainer.ServeHTTP(rr, r)

	if rr.Code != http.StatusOK {
		t.Errorf("Incorrect status code, expected 200 got %d", rr.Code)
		return
	}

	result, err := utils.FancyJson[rpc.SimpleMessage](rr.Body.Bytes())
	if err != nil {
		t.Error(err)
		return
	}

	if result.Message != messageText {
		t.Errorf("Incorrect response, got %s", result.Message)
		return
	}
}

func ExampleRpcContainer() {
	rpcContainer := rpc.NewRpcContainer()

	// Add your stuff

	rpcContainer.SetupMux(http.DefaultServeMux, "/rpc")
	http.ListenAndServe("0.0.0.0:3456", nil)
}

func ExampleRpcContainer_AddFunction() {
	rpcContainer := rpc.NewRpcContainer()

	rpcContainer.AddFunction("hello", func(*rpc.RpcRequest) rpc.RpcResponse {
		return rpc.Json{
			Body: map[string]string{
				"hi": "dave",
			},
		}
	})

	rpcContainer.AddFunction("hello-named", func(r *rpc.RpcRequest) rpc.RpcResponse {
		body := r.Body.(rpc.SimpleMessage)

		return rpc.Json{
			Body: map[string]string{
				"hi": body.Message,
			},
		}
	}).SetBodyType(reflect.TypeOf(rpc.SimpleMessage{}))

	rpcContainer.SetupMux(http.DefaultServeMux, "/rpc")
	http.ListenAndServe("0.0.0.0:3456", nil)
}
