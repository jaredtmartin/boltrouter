package route_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaredtmartin/route"
)

func handleGetDog(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World! Let's Get a Dog!"))
}
func handlePostDog(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World! Let's Post a Dog!"))
}
func handleDeleteDog(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World! Let's Delete a Dog!"))
}
func handlePutDog(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World! Let's Put a Dog!"))
}
func handlePatchDog(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World! Let's Patch a Dog!"))
}

func testRoute(server *httptest.Server, method, path, expectedBody string, t *testing.T) {
	req, err := http.NewRequest(method, server.URL+path, nil)
	if err != nil {
		t.Fatalf("Failed to create %s request: %v", method, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform %s request: %v", method, err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.String()

	if body != expectedBody {
		t.Errorf("Expected response body %q for %s request, got %q", expectedBody, method, body)
	}
}
func TestGet(t *testing.T) {
	mux := http.NewServeMux()
	route.Path(mux, "/dog").Get(handleGetDog)
	server := httptest.NewServer(mux)
	defer server.Close()
	testRoute(server, "GET", "/dog", "Hello, World! Let's Get a Dog!", t)
	testRoute(server, "POST", "/dog", "Method Not Allowed\n", t)
	testRoute(server, "GET", "/cat", "404 page not found\n", t)
}
func TestMultiMethod(t *testing.T) {
	mux := http.NewServeMux()
	route.Path(mux, "/dog").
		Get(handleGetDog).
		Post(handlePostDog).
		Delete(handleDeleteDog).
		Put(handlePutDog).
		Patch(handlePatchDog)

	server := httptest.NewServer(mux)
	defer server.Close()
	testRoute(server, "GET", "/dog", "Hello, World! Let's Get a Dog!", t)
	testRoute(server, "POST", "/dog", "Hello, World! Let's Post a Dog!", t)
	testRoute(server, "DELETE", "/dog", "Hello, World! Let's Delete a Dog!", t)
	testRoute(server, "PUT", "/dog", "Hello, World! Let's Put a Dog!", t)
	testRoute(server, "PATCH", "/dog", "Hello, World! Let's Patch a Dog!", t)
	testRoute(server, "PATCH", "/cat", "404 page not found\n", t)
}
