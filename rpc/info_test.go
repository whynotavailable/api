package rpc_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/whynotavailable/api/rpc"
)

func TestInfo(t *testing.T) {
	rpcContainer := rpc.RpcContainer{}
	stripContainer := http.StripPrefix("/rpc", &rpcContainer) // Simulate

	r, err := http.NewRequest(http.MethodGet, "/rpc/_info", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()

	stripContainer.ServeHTTP(rr, r)

	data, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Error(err)
		return
	}

	info := map[string]any{}
	err = json.Unmarshal(data, &info)
	if err != nil {
		t.Error(err)
		return
	}
}
