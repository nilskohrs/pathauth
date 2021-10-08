package pathauth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nilskohrs/pathauth"
)

func TestShouldAllowUser(t *testing.T) {
	cfg := pathauth.CreateConfig()
	cfg.Source.Name = "X-Roles"
	cfg.Source.Type = "header"
	cfg.Source.Delimiter = ","

	cfg.Authorization = append(cfg.Authorization, pathauth.Authorization{
		Path:     ".*/admin/.*",
		Priority: 1,
		Allowed:  []string{"admin"},
	})

	cfg.Authorization = append(cfg.Authorization, pathauth.Authorization{
		Path:     ".*/admin/health",
		Host:     "localhost",
		Priority: 0,
		Allowed:  []string{"monitoring"},
		Method:   []string{"Get"},
	})

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := pathauth.New(ctx, next, cfg, "pathauth")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/admin/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("X-Roles", "monitoring")

	handler.ServeHTTP(recorder, req)

	assertAllowed(t, recorder, true)
}

func assertAllowed(t *testing.T, recorder *httptest.ResponseRecorder, allowed bool) {
	t.Helper()
	if recorder.Result().StatusCode == 403 && allowed {
		t.Errorf("request was forbidden, expected allowed")
	} else if recorder.Result().StatusCode != 403 && !allowed {
		t.Errorf("request was allowed, expected forbidden")
	}
}
