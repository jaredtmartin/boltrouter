package boltrouter_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaredtmartin/bolt-go"
	"github.com/jaredtmartin/boltrouter"
)

func handleGetDog(w http.ResponseWriter, r *http.Request) (bolt.Element, error) {
	return bolt.String("Hello, World! Let's Get a Dog!"), nil
}
func handlePostDog(w http.ResponseWriter, r *http.Request) (bolt.Element, error) {
	return bolt.String("Hello, World! Let's Post a Dog!"), nil
}
func handleDeleteDog(w http.ResponseWriter, r *http.Request) (bolt.Element, error) {
	return bolt.String("Hello, World! Let's Delete a Dog!"), nil
}
func handlePutDog(w http.ResponseWriter, r *http.Request) (bolt.Element, error) {
	return bolt.String("Hello, World! Let's Put a Dog!"), nil
}
func handlePatchDog(w http.ResponseWriter, r *http.Request) (bolt.Element, error) {
	return bolt.String("Hello, World! Let's Patch a Dog!"), nil
}
func handleSimpleError(w http.ResponseWriter, r *http.Request) (bolt.Element, error) {
	return nil, fmt.Errorf("Something went wrong!")
}
func handleErrorWithContent(w http.ResponseWriter, r *http.Request) (bolt.Element, error) {
	return bolt.Button("Log In").Href("/login"), fmt.Errorf("You're not logged in.")
}

func testRoute(server *httptest.Server, method, path, expectedBody string, t *testing.T) {
	t.Helper()
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
func layout(w http.ResponseWriter, r *http.Request, elements ...bolt.Element) bolt.Element {
	return bolt.NewElement("layout").Children(elements...)
}
func errorPage(err error, children ...bolt.Element) bolt.Element {
	return bolt.NewElement("error").Text(err.Error()).Children(children...)
}

func TestGet(t *testing.T) {
	mux := http.NewServeMux()
	router := boltrouter.NewRouter(mux, layout, errorPage)
	router.Path("/dog").Get(handleGetDog)
	server := httptest.NewServer(mux)
	defer server.Close()
	testRoute(server, "GET", "/dog", "<layout>Hello, World! Let's Get a Dog!</layout>", t)
	testRoute(server, "POST", "/dog", "Method Not Allowed\n", t)
	testRoute(server, "GET", "/cat", "404 page not found\n", t)
}
func TestMultiMethod(t *testing.T) {
	mux := http.NewServeMux()
	router := boltrouter.NewRouter(mux, layout, errorPage)
	router.Path("/dog").
		Get(handleGetDog).
		Post(handlePostDog).
		Delete(handleDeleteDog).
		Put(handlePutDog).
		Patch(handlePatchDog)

	server := httptest.NewServer(mux)
	defer server.Close()
	testRoute(server, "GET", "/dog", "<layout>Hello, World! Let's Get a Dog!</layout>", t)
	testRoute(server, "POST", "/dog", "<layout>Hello, World! Let's Post a Dog!</layout>", t)
	testRoute(server, "DELETE", "/dog", "<layout>Hello, World! Let's Delete a Dog!</layout>", t)
	testRoute(server, "PUT", "/dog", "<layout>Hello, World! Let's Put a Dog!</layout>", t)
	testRoute(server, "PATCH", "/dog", "<layout>Hello, World! Let's Patch a Dog!</layout>", t)
	testRoute(server, "PATCH", "/cat", "404 page not found\n", t)
}
func TestPost(t *testing.T) {
	mux := http.NewServeMux()
	router := boltrouter.NewRouter(mux, layout, errorPage)
	router.Path("/dog").Post(handlePostDog)
	server := httptest.NewServer(mux)
	defer server.Close()
	testRoute(server, "POST", "/dog", "<layout>Hello, World! Let's Post a Dog!</layout>", t)
	testRoute(server, "GET", "/dog", "Method Not Allowed\n", t)
	testRoute(server, "POST", "/cat", "404 page not found\n", t)
}
func TestErrors(t *testing.T) {
	mux := http.NewServeMux()
	router := boltrouter.NewRouter(mux, layout, errorPage)
	router.Path("/err").Get(handleSimpleError)
	router.Path("/auth").Get(handleErrorWithContent)
	server := httptest.NewServer(mux)
	defer server.Close()
	testRoute(server, "GET", "/err", "<error>Something went wrong!</error>", t)
	testRoute(server, "GET", "/auth", "<error><a href=\"/login\" type=\"button\">Log In</a>You're not logged in.</error>", t)
}
