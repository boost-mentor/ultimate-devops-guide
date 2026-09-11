package service

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/boost-mentor/kubernetes-production-labs/apps/devops-may-cry/internal/store"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	return New(store.NewMemory(), slog.New(slog.NewTextHandler(io.Discard, nil)), "pod-test", true)
}

func TestQuoteIncludesPodAndStore(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/quote", nil)
	response := httptest.NewRecorder()
	testServer(t).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["pod"] != "pod-test" {
		t.Fatalf("unexpected body: %#v", body)
	}
}

func TestOrderContractAndSequence(t *testing.T) {
	server := testServer(t)
	for index, want := range []string{"NS-0001", "NS-0002"} {
		request := httptest.NewRequest(http.MethodGet, "/order?on=dns-ghost", nil)
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("request %d status = %d", index, response.Code)
		}
		var body store.Order
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Number != want || body.Target != "dns-ghost" || body.HandledBy != "pod-test" {
			t.Fatalf("request %d body = %#v", index, body)
		}
	}
}

func TestPostOrderRejectsUnknownFields(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"target":"disk-full","secret":true}`))
	response := httptest.NewRecorder()
	testServer(t).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestReadyAndLive(t *testing.T) {
	server := testServer(t)
	for _, path := range []string{"/livez", "/readyz"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s status = %d", path, response.Code)
		}
	}
}

func TestUnknownPathIsNotAHealthyFallback(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/nope", nil)
	response := httptest.NewRecorder()
	testServer(t).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestDangerousLabEndpointsAreDisabledByDefault(t *testing.T) {
	server := New(store.NewMemory(), slog.New(slog.NewTextHandler(io.Discard, nil)), "pod-test", false)
	for _, path := range []string{"/overload?sec=1", "/wound?mb=1"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, request)
		if response.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, want %d", path, response.Code, http.StatusNotFound)
		}
	}
}
