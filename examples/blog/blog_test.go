package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pashagolub/pgxmock/v2"
)

func (a *api) assertJSON(actual []byte, data interface{}, t *testing.T) {
	expected, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when marshaling expected json data", err)
	}

	if !bytes.Equal(expected, actual) {
		t.Errorf("the expected json: %s is different from actual %s", expected, actual)
	}
}

func TestShouldGetPosts(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	// create app with mocked db, request and response to test
	app := &api{mock}
	req, err := http.NewRequest("GET", "http://localhost/posts", nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected while creating request", err)
	}
	w := httptest.NewRecorder()

	// before we actually execute our api function, we need to expect required DB actions
	rows := mock.NewRows([]string{"id", "title", "body"}).
		AddRow(1, "post 1", "hello").
		AddRow(2, "post 2", "world")

	mock.ExpectQuery("^SELECT (.+) FROM posts$").WillReturnRows(rows)

	// now we execute our request
	app.posts(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status code to be 200, but got: %d\nBody: %v", w.Code, w.Body)
	}

	data := struct {
		Posts []*post
	}{Posts: []*post{
		{ID: 1, Title: "post 1", Body: "hello"},
		{ID: 2, Title: "post 2", Body: "world"},
	}}
	app.assertJSON(w.Body.Bytes(), data, t)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestShouldRespondWithErrorOnFailure(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	// create app with mocked db, request and response to test
	app := &api{mock}
	req, err := http.NewRequest("GET", "http://localhost/posts", nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected while creating request", err)
	}
	w := httptest.NewRecorder()

	// before we actually execute our api function, we need to expect required DB actions
	mock.ExpectQuery("^SELECT (.+) FROM posts$").WillReturnError(fmt.Errorf("some error"))

	// now we execute our request
	app.posts(w, req)

	if w.Code != 500 {
		t.Fatalf("expected status code to be 500, but got: %d", w.Code)
	}

	data := struct {
		Error string
	}{"failed to fetch posts: some error"}
	app.assertJSON(w.Body.Bytes(), data, t)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestNoPostsReturned(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	// create app with mocked db, request and response to test
	app := &api{mock}
	req, err := http.NewRequest("GET", "http://localhost/posts", nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected while creating request", err)
	}
	w := httptest.NewRecorder()

	mock.ExpectQuery("^SELECT (.+) FROM posts$").WillReturnRows(mock.NewRows([]string{"id", "title", "body"}))
	// now we execute our request
	app.posts(w, req)
	if w.Code != 200 {
		t.Fatalf("expected status code to be 200, but got: %d\nBody: %v", w.Code, w.Body)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
