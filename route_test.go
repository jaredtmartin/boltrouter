package boltrouter_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaredtmartin/bolt-go"
	. "github.com/jaredtmartin/boltrouter"
)

func handleGetDog(w http.ResponseWriter, r *http.Request) ResponseType {
	return Content(bolt.String("Hello, World! Let's Get a Dog!"), bolt.String("Here's some more content!"))
}
func handlePostDog(w http.ResponseWriter, r *http.Request) ResponseType {
	return Content(bolt.String("Hello, World! Let's Post a Dog!"))
}
func handleDeleteDog(w http.ResponseWriter, r *http.Request) ResponseType {
	return Content(bolt.String("Hello, World! Let's Delete a Dog!"))
}
func handlePutDog(w http.ResponseWriter, r *http.Request) ResponseType {
	return Content(bolt.String("Hello, World! Let's Put a Dog!"))
}
func handlePatchDog(w http.ResponseWriter, r *http.Request) ResponseType {
	return Content(bolt.String("Hello, World! Let's Patch a Dog!"))
}
func handleSimpleError(w http.ResponseWriter, r *http.Request) ResponseType {
	return Error(fmt.Errorf("Something went wrong!"))
}
func handleDetailedError(w http.ResponseWriter, r *http.Request) ResponseType {
	err := fmt.Errorf("Details about the error.")
	return Error(fmt.Errorf("Something went wrong!: %w", err))
}

// func handleErrorWithContent(w http.ResponseWriter, r *http.Request) Response {
// 	return Response().Content(bolt.Button("Log In").Href("/login")).Error("You're not logged in.")
// }

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
func errorPage(err ResponseType) bolt.Element {
	// get everything before the : in the error message
	return bolt.NewElement("div").Children(
		bolt.NewElement("msg").Text(err.ErrPublic()),
		bolt.NewElement("detail").Text(err.ErrDetail()),
	)
}

func TestGet(t *testing.T) {
	mux := http.NewServeMux()
	router := NewRouter(mux, layout, errorPage)
	router.Path("/dog").Get(handleGetDog)
	server := httptest.NewServer(mux)
	defer server.Close()
	testRoute(server, "GET", "/dog", "<layout>Hello, World! Let's Get a Dog!Here's some more content!</layout>", t)
	testRoute(server, "POST", "/dog", "Method Not Allowed\n", t)
	testRoute(server, "GET", "/cat", "404 page not found\n", t)
}
func TestMultiMethod(t *testing.T) {
	mux := http.NewServeMux()
	router := NewRouter(mux, layout, errorPage)
	router.Path("/dog").
		Get(handleGetDog).
		Post(handlePostDog).
		Delete(handleDeleteDog).
		Put(handlePutDog).
		Patch(handlePatchDog)

	server := httptest.NewServer(mux)
	defer server.Close()
	testRoute(server, "GET", "/dog", "<layout>Hello, World! Let's Get a Dog!Here's some more content!</layout>", t)
	testRoute(server, "POST", "/dog", "<layout>Hello, World! Let's Post a Dog!</layout>", t)
	testRoute(server, "DELETE", "/dog", "<layout>Hello, World! Let's Delete a Dog!</layout>", t)
	testRoute(server, "PUT", "/dog", "<layout>Hello, World! Let's Put a Dog!</layout>", t)
	testRoute(server, "PATCH", "/dog", "<layout>Hello, World! Let's Patch a Dog!</layout>", t)
	testRoute(server, "PATCH", "/cat", "404 page not found\n", t)
}
func TestPost(t *testing.T) {
	mux := http.NewServeMux()
	router := NewRouter(mux, layout, errorPage)
	router.Path("/dog").Post(handlePostDog)
	server := httptest.NewServer(mux)
	defer server.Close()
	testRoute(server, "POST", "/dog", "<layout>Hello, World! Let's Post a Dog!</layout>", t)
	testRoute(server, "GET", "/dog", "Method Not Allowed\n", t)
	testRoute(server, "POST", "/cat", "404 page not found\n", t)
}
func TestErrors(t *testing.T) {
	mux := http.NewServeMux()
	router := NewRouter(mux, layout, errorPage)
	router.Path("/err").Get(handleSimpleError)
	server := httptest.NewServer(mux)
	defer server.Close()
	testRoute(server, "GET", "/err", "<layout><div><msg>Something went wrong!</msg><detail></detail></div></layout>", t)
	router.Path("/err3").Get(handleDetailedError)
	testRoute(server, "GET", "/err3", "<layout><div><msg>Something went wrong!</msg><detail>Details about the error.</detail></div></layout>", t)

}
