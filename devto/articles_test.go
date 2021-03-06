package devto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/VictorAvelar/devto-api-go/testdata"
)

var ctx = context.Background()

func TestArticlesResource_List(t *testing.T) {
	setup()
	defer teardown()
	cont, err := ioutil.ReadAll(strings.NewReader(testdata.ListResponse))
	if err != nil {
		t.Error(err)
	}

	testMux.HandleFunc("/api/articles", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(cont)
	})

	http.DefaultClient.Get(testServer.URL)

	var ctx = context.Background()
	list, err := testClientPub.Articles.List(ctx, ArticleListOptions{})
	if err != nil {
		t.Error(err)
	}

	if len(list) != 3 {
		t.Errorf("not all articles where parsed")
	}

	for _, a := range list {
		if a.Title == "" {
			t.Error("parsing failed / empty titles")
		}
	}
}

func TestArticlesResource_ListWithQueryParams(t *testing.T) {
	setup()
	defer teardown()
	testMux.HandleFunc("/api/articles", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.String(), "username=victoravelar") {
			t.Error("url mismatch")
		}

		var articles []ListedArticle
		if err := json.NewEncoder(w).Encode(&articles); err != nil {
			t.Errorf("error marshalling ListedArticles to JSON: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	q := ArticleListOptions{
		Tags:     "go",
		Username: "victoravelar",
		State:    "fresh",
		Top:      "1",
		Page:     1,
	}
	list, err := testClientPub.Articles.List(ctx, q)
	if err != nil {
		t.Error(err)
	}
	if len(list) != 0 {
		t.Error("response is unexpected")
	}
}

func TestArticlesResource_ListForTag(t *testing.T) {
	setup()
	defer teardown()
	testMux.HandleFunc("/api/articles", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("tag") != "go" {
			t.Errorf(`expected "tag" query param to be "go", got "%s"`, q.Get("tag"))
		}
		if q.Get("page") != "3" {
			t.Errorf(`expected "page" query param to be "3", got "%s"`, q.Get("page"))
		}

		var articles []ListedArticle
		if err := json.NewEncoder(w).Encode(&articles); err != nil {
			t.Errorf("error marshalling ListedArticles to JSON: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	list, err := testClientPub.Articles.ListForTag(ctx, "go", 3)
	if err != nil {
		t.Error(err)
	}
	if len(list) != 0 {
		t.Error("response is unexpected")
	}
}

func TestArticlesResource_ListForUser(t *testing.T) {
	setup()
	defer teardown()
	testMux.HandleFunc("/api/articles", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("username") != "victoravelar" {
			t.Errorf(`expected "username" query param to be "victoravelar", got "%s"`, q.Get("username"))
		}
		if q.Get("page") != "3" {
			t.Errorf(`expected "page" query param to be "3", got "%s"`, q.Get("page"))
		}

		var articles []ListedArticle
		if err := json.NewEncoder(w).Encode(&articles); err != nil {
			t.Errorf("error marshalling ListedArticles to JSON: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	list, err := testClientPub.Articles.ListForUser(ctx, "victoravelar", 3)
	if err != nil {
		t.Error(err)
	}
	if len(list) != 0 {
		t.Error("response is unexpected")
	}
}

var gopherDay = time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC)

var myArticles = []ListedArticle{{
	TypeOf:      "article",
	ID:          1123,
	Title:       "Joy of tunneling",
	Description: "How to dig the perfect gopher hole",
	User: User{
		Name:     "B. Neathyourlawn",
		Username: "bneathyourlawn",
	},
	TagList:   []string{"go"},
	Published: false,
}, {
	TypeOf:      "article",
	ID:          5813,
	Title:       "Carrot stew recipe",
	Description: "A traditional meal since gophers first came to a garden near you",
	User: User{
		Name:     "B. Neathyourlawn",
		Username: "bneathyourlawn",
	},
	TagList:     []string{"go", "cooking"},
	Published:   true,
	PublishedAt: &gopherDay,
}}

func TestArticlesResource_ListMyPublishedArticles(t *testing.T) {
	setup()
	defer teardown()
	testMux.HandleFunc("/api/articles/me/published", func(w http.ResponseWriter, r *http.Request) {
		articles := myArticles[:1]
		if err := json.NewEncoder(w).Encode(&articles); err != nil {
			t.Errorf("error marshalling ListedArticles to JSON: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	list, err := testClientPro.Articles.ListMyPublishedArticles(ctx, &MyArticlesOptions{Page: 1})
	if err != nil {
		t.Error(err)
	}
	if len(list) != 1 {
		t.Fatalf("should have gotten 1 article back, got %d", len(list))
	}
	if list[0].Title != "Joy of tunneling" {
		t.Errorf(`expected title to be "Joy of tunneling", got back "%s"`, list[0].Title)
	}

	// Test without authentication
	list, err = testClientPub.Articles.ListMyPublishedArticles(ctx, &MyArticlesOptions{Page: 1})
	if err != ErrProtectedEndpoint {
		t.Errorf("error should be ErrProtectedEndpoint, was %v", err)
	}
	if list != nil {
		t.Errorf("should get back nil slice of articles from unauthenticated request")
	}
}

func TestArticlesResource_ListMyUnpublishedArticles(t *testing.T) {
	setup()
	defer teardown()
	testMux.HandleFunc("/api/articles/me/unpublished", func(w http.ResponseWriter, r *http.Request) {
		articles := myArticles[1:]
		if err := json.NewEncoder(w).Encode(&articles); err != nil {
			t.Errorf("error marshalling ListedArticles to JSON: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	list, err := testClientPro.Articles.ListMyUnpublishedArticles(ctx, &MyArticlesOptions{Page: 1})
	if err != nil {
		t.Error(err)
	}
	if len(list) != 1 {
		t.Fatalf("should have gotten 1 article back, got %d", len(list))
	}
	if list[0].Title != "Carrot stew recipe" {
		t.Errorf(`expected title to be "Carrot stew recipe", got back "%s"`, list[0].Title)
	}

	// Test without authentication
	list, err = testClientPub.Articles.ListMyUnpublishedArticles(ctx, &MyArticlesOptions{Page: 1})
	if err != ErrProtectedEndpoint {
		t.Errorf("error should be ErrProtectedEndpoint, was %v", err)
	}
	if list != nil {
		t.Errorf("should get back nil slice of articles from unauthenticated request")
	}
}

func TestArticlesResource_ListAllMyArticles(t *testing.T) {
	setup()
	defer teardown()
	testMux.HandleFunc("/api/articles/me/all", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(&myArticles); err != nil {
			t.Errorf("error marshalling ListedArticles to JSON: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	list, err := testClientPro.Articles.ListAllMyArticles(ctx, &MyArticlesOptions{Page: 1})
	if err != nil {
		t.Error(err)
	}
	if len(list) != 2 {
		t.Fatalf("should have gotten 1 article back, got %d", len(list))
	}
	if list[0].Title != "Joy of tunneling" {
		t.Errorf(`expected title to be "Joy of tunneling", got back "%s"`, list[0].Title)
	}
	if list[1].Title != "Carrot stew recipe" {
		t.Errorf(`expected title to be "Carrot stew recipe", got back "%s"`, list[1].Title)
	}

	// Test without authentication
	list, err = testClientPub.Articles.ListAllMyArticles(ctx, &MyArticlesOptions{Page: 1})
	if err != ErrProtectedEndpoint {
		t.Errorf("error should be ErrProtectedEndpoint, was %v", err)
	}
	if list != nil {
		t.Errorf("should get back nil slice of articles from unauthenticated request")
	}
}

func TestArticlesResource_Find(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api/articles/", func(w http.ResponseWriter, r *http.Request) {
		id, err := getNumericPathComponent(r, 2)
		if err != nil {
			sendErrorResponse(t, w, ErrorResponse{
				ErrorMessage: fmt.Sprintf("error getting article ID from URL: %v", err),
				Status:       http.StatusBadRequest,
			})
			return
		}

		if id != 164198 {
			sendErrorResponse(t, w, ErrorResponse{
				ErrorMessage: "not found",
				Status:       http.StatusNotFound,
			})
			return
		}

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testdata.FindResponse))
	})

	// retrieve article 164198
	article, err := testClientPub.Articles.Find(ctx, 164198)
	if err != nil {
		t.Error(err)
	}

	if article.ID != 164198 {
		t.Error("article returned is not the one requested")
	}

	// test 404 response
	_, err = testClientPub.Articles.Find(ctx, 11235813)
	checkErrorResponse(t, err, http.StatusNotFound)
}

func TestArticlesResource_New(t *testing.T) {
	setup()
	defer teardown()
	testMux.HandleFunc("/api/articles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatal("invalid method for request")
		}

		var au ArticleUpdate
		if err := json.NewDecoder(r.Body).Decode(&au); err != nil {
			sendErrorResponse(t, w, ErrorResponse{
				ErrorMessage: "Bad Request",
				Status:       http.StatusBadRequest,
			})
			return
		} else if au.Title == "" {
			sendErrorResponse(t, w, ErrorResponse{
				ErrorMessage: "param is missing or the value is empty",
				Status:       http.StatusUnprocessableEntity,
			})
			return
		}

		a := Article{
			Title:     au.Title,
			Published: au.Published,
		}
		b, err := json.Marshal(&a)
		if err != nil {
			t.Errorf("error marshalling Article to JSON")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(b)
	})

	// Test successfully uploading an article
	res, err := testClientPro.Articles.New(ctx, ArticleUpdate{
		Title:     "Demo article",
		Published: false,
	})
	if err != nil {
		t.Fatalf("unexpected error publishing article: %v", err)
	}

	if res.Title != "Demo article" {
		t.Error("article parsing failed")
	}

	// Test uploading an invalid article
	_, err = testClientPro.Articles.New(ctx, ArticleUpdate{})
	checkErrorResponse(t, err, http.StatusUnprocessableEntity)
}

func TestArticlesResource_NewFailsWhenInsecure(t *testing.T) {
	setup()
	defer teardown()

	_, err := testClientPub.Articles.New(ctx, ArticleUpdate{
		Title:     "Demo article",
		Published: false,
	})

	if err != ErrProtectedEndpoint {
		t.Error("auth check failed")
	}
}

func TestArticlesResource_Update(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api/articles/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			sendErrorResponse(t, w, ErrorResponse{
				ErrorMessage: "invalid method for request",
				Status:       http.StatusMethodNotAllowed,
			})
			return
		}

		id, err := getNumericPathComponent(r, 2)
		if err != nil {
			sendErrorResponse(t, w, ErrorResponse{
				ErrorMessage: fmt.Sprintf("error getting article ID from URL: %v", err),
				Status:       http.StatusBadRequest,
			})
			return
		}

		if id != 164198 {
			sendErrorResponse(t, w, ErrorResponse{
				ErrorMessage: "not found",
				Status:       http.StatusNotFound,
			})
			return
		}

		var au ArticleUpdate
		if err := json.NewDecoder(r.Body).Decode(&au); err != nil {
			sendErrorResponse(t, w, ErrorResponse{
				ErrorMessage: fmt.Sprintf("error unmarshalling ArticleUpdate from JSON: %v", err),
				Status:       http.StatusBadRequest,
			})
			return
		}

		a := Article{
			ID:        uint32(id),
			Title:     au.Title,
			Published: au.Published,
		}
		b, err := json.Marshal(&a)
		if err != nil {
			sendErrorResponse(t, w, ErrorResponse{
				ErrorMessage: fmt.Sprintf("error marshalling Article to JSON: %v", err),
				Status:       http.StatusInternalServerError,
			})
			return
		}

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(b)
	})

	// Test successfully updating article
	res, err := testClientPro.Articles.Update(ctx, ArticleUpdate{
		Title:     "Demo article",
		Published: false,
	}, 164198)
	if err != nil {
		t.Fatalf("unexpected error publishing article: %v", err)
	}

	if res.Title != "Demo article" {
		t.Error("article parsing failed")
	}

	// Test updating an article that does not exist
	_, err = testClientPro.Articles.Update(ctx, ArticleUpdate{
		Title:     "Demo article",
		Published: false,
	}, 11235813)
	checkErrorResponse(t, err, http.StatusNotFound)
}

func TestArticlesResource_UpdateFailsWhenInsecure(t *testing.T) {
	setup()
	defer teardown()

	_, err := testClientPub.Articles.Update(ctx, ArticleUpdate{
		Title:     "Demo article",
		Published: false,
	}, 164198)

	if !reflect.DeepEqual(err, ErrProtectedEndpoint) {
		t.Error("auth check failed")
	}
}

//
// helper functions
//

func getNumericPathComponent(r *http.Request, index int) (int, error) {
	p := r.URL.Path
	pathComponents := strings.Split(p, "/")

	// Remove leading slash path component
	if pathComponents[0] == "" {
		if len(pathComponents) == 0 {
			return 0, errors.New("path was empty string")
		}
		pathComponents = pathComponents[1:]
	}
	if len(pathComponents) <= index {
		return 0, errors.New("n-th path component (zero-indexed) not found")
	}
	return strconv.Atoi(pathComponents[index])
}

func sendErrorResponse(t *testing.T, w http.ResponseWriter, res ErrorResponse) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(res.Status)
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		t.Fatalf("error marshalling error status: %v", err)
	}
}

func checkErrorResponse(t *testing.T, err error, expStatus int) {
	switch err := err.(type) {
	case nil:
		t.Errorf(
			"got nil error from find article endpoint when we were expecting "+
				"%d ErrorResponse",
			expStatus,
		)
	case *ErrorResponse:
		if err.Status != expStatus {
			t.Errorf("error should be %d, got %d", expStatus, err.Status)
		}
	default:
		t.Errorf("error should be of type *ErrorResponse; was %T", err)
	}
}
