package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSRF_MissingTokenForbidden(t *testing.T) {
	csrf := NewCSRF("test-secret")
	h := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/admin/albums", strings.NewReader(""))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Result().StatusCode)
	}
}

func TestCSRF_ValidHeaderPasses(t *testing.T) {
	csrf := NewCSRF("test-secret")
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := csrf.Middleware()(inner)

	getReq := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)

	cookies := getRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected csrf cookie to be set")
	}
	csrfCookie := cookies[0]

	postReq := httptest.NewRequest(http.MethodPost, "/admin/albums", strings.NewReader(""))
	postReq.AddCookie(csrfCookie)
	postReq.Header.Set(csrfHeaderName, csrfCookie.Value)
	postRec := httptest.NewRecorder()
	h.ServeHTTP(postRec, postReq)

	if postRec.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", postRec.Result().StatusCode)
	}
}

func TestCSRF_InvalidHeaderRejected(t *testing.T) {
	csrf := NewCSRF("test-secret")
	h := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	getReq := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)

	cookies := getRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected csrf cookie to be set")
	}

	postReq := httptest.NewRequest(http.MethodPost, "/admin/albums", strings.NewReader(""))
	postReq.AddCookie(cookies[0])
	postReq.Header.Set(csrfHeaderName, "invalid.token")
	postRec := httptest.NewRecorder()
	h.ServeHTTP(postRec, postReq)

	if postRec.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", postRec.Result().StatusCode)
	}
}
