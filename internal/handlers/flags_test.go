package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"featureflags/internal/store"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	s := store.New()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /flags", Create(s))
	mux.HandleFunc("GET /flags", List(s))
	mux.HandleFunc("GET /flags/{key}", Get(s))
	mux.HandleFunc("PUT /flags/{key}", Update(s))
	mux.HandleFunc("DELETE /flags/{key}", Delete(s))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts
}

func postJSON(t *testing.T, url, body string) *http.Response {
	t.Helper()
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	return resp
}

func doJSON(t *testing.T, method, url, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s failed: %v", method, err)
	}
	return resp
}

func decodeError(t *testing.T, resp *http.Response) map[string]string {
	t.Helper()
	defer resp.Body.Close()
	var e map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
		t.Fatalf("decode error object: %v", err)
	}
	return e
}

func TestCreateFlag(t *testing.T) {
	ts := newTestServer(t)

	resp := postJSON(t, ts.URL+"/flags", `{"key":"feature-x","enabled":true,"description":"test","rollout_percent":50}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST status = %d, want 201", resp.StatusCode)
	}
	resp.Body.Close()

	listResp, err := http.Get(ts.URL + "/flags")
	if err != nil {
		t.Fatalf("GET /flags failed: %v", err)
	}
	defer listResp.Body.Close()
	var flags []map[string]any
	if err := json.NewDecoder(listResp.Body).Decode(&flags); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(flags) != 1 {
		t.Fatalf("len(flags) = %d, want 1", len(flags))
	}
	if flags[0]["key"] != "feature-x" {
		t.Fatalf("key = %v, want feature-x", flags[0]["key"])
	}
}

func TestCreateDuplicateConflict(t *testing.T) {
	ts := newTestServer(t)

	body := `{"key":"dup","enabled":true}`
	resp := postJSON(t, ts.URL+"/flags", body)
	resp.Body.Close()

	resp2 := postJSON(t, ts.URL+"/flags", body)
	if resp2.StatusCode != http.StatusConflict {
		t.Fatalf("second POST status = %d, want 409", resp2.StatusCode)
	}
	e := decodeError(t, resp2)
	if e["error"] == "" {
		t.Fatalf("expected non-empty error field")
	}
}

func TestCreateValidationErrors(t *testing.T) {
	ts := newTestServer(t)

	cases := []struct {
		name string
		body string
	}{
		{"invalid json", `{"key":`},
		{"empty key", `{"key":"","enabled":true}`},
		{"rollout too low", `{"key":"a","enabled":true,"rollout_percent":-1}`},
		{"rollout too high", `{"key":"a","enabled":true,"rollout_percent":101}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := postJSON(t, ts.URL+"/flags", c.body)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", resp.StatusCode)
			}
			e := decodeError(t, resp)
			if e["error"] == "" {
				t.Fatalf("expected non-empty error field")
			}
		})
	}
}

func TestCreateDefaultRollout(t *testing.T) {
	ts := newTestServer(t)

	resp := postJSON(t, ts.URL+"/flags", `{"key":"defaulted","enabled":true}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	defer resp.Body.Close()
	var flag map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&flag); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if flag["rollout_percent"] != float64(100) {
		t.Fatalf("rollout_percent = %v, want 100", flag["rollout_percent"])
	}
}

func TestListEmpty(t *testing.T) {
	ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/flags")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body := new(bytes.Buffer)
	body.ReadFrom(resp.Body)
	trimmed := strings.TrimSpace(body.String())
	if trimmed != "[]" {
		t.Fatalf("body = %q, want []", trimmed)
	}
}

func TestGetFlag(t *testing.T) {
	ts := newTestServer(t)
	postJSON(t, ts.URL+"/flags", `{"key":"g","enabled":true}`).Body.Close()

	resp, err := http.Get(ts.URL + "/flags/g")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var flag map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&flag); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if flag["key"] != "g" {
		t.Fatalf("key = %v, want g", flag["key"])
	}
}

func TestGetNotFound(t *testing.T) {
	ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/flags/missing")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	e := decodeError(t, resp)
	if e["error"] == "" {
		t.Fatalf("expected non-empty error field")
	}
}

func TestUpdateFlag(t *testing.T) {
	ts := newTestServer(t)
	postJSON(t, ts.URL+"/flags", `{"key":"u","enabled":true,"description":"old","rollout_percent":10}`).Body.Close()

	resp := doJSON(t, http.MethodPut, ts.URL+"/flags/u", `{"enabled":false,"description":"new","rollout_percent":80}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	defer resp.Body.Close()
	var flag map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&flag); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if flag["enabled"] != false {
		t.Fatalf("enabled = %v, want false", flag["enabled"])
	}
	if flag["description"] != "new" {
		t.Fatalf("description = %v, want new", flag["description"])
	}
	if flag["rollout_percent"] != float64(80) {
		t.Fatalf("rollout_percent = %v, want 80", flag["rollout_percent"])
	}
}

func TestUpdateNotFound(t *testing.T) {
	ts := newTestServer(t)

	resp := doJSON(t, http.MethodPut, ts.URL+"/flags/missing", `{"enabled":true,"description":"","rollout_percent":50}`)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	decodeError(t, resp)
}

func TestUpdateInvalidRollout(t *testing.T) {
	ts := newTestServer(t)
	postJSON(t, ts.URL+"/flags", `{"key":"u","enabled":true}`).Body.Close()

	resp := doJSON(t, http.MethodPut, ts.URL+"/flags/u", `{"enabled":true,"description":"","rollout_percent":150}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	decodeError(t, resp)
}

func TestDeleteFlag(t *testing.T) {
	ts := newTestServer(t)
	postJSON(t, ts.URL+"/flags", `{"key":"d","enabled":true}`).Body.Close()

	resp := doJSON(t, http.MethodDelete, ts.URL+"/flags/d", "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	resp.Body.Close()

	getResp, err := http.Get(ts.URL + "/flags/d")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET after delete = %d, want 404", getResp.StatusCode)
	}
}

func TestDeleteNotFound(t *testing.T) {
	ts := newTestServer(t)

	resp := doJSON(t, http.MethodDelete, ts.URL+"/flags/missing", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	decodeError(t, resp)
}

func TestCreateBodyTooLarge(t *testing.T) {
	ts := newTestServer(t)

	big := strings.Repeat("a", maxBodyBytes+10)
	body := `{"key":"big","enabled":true,"description":"` + big + `"}`

	resp := postJSON(t, ts.URL+"/flags", body)
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", resp.StatusCode)
	}
	e := decodeError(t, resp)
	if e["error"] == "" {
		t.Fatalf("expected non-empty error field")
	}
}

func TestUpdateBodyTooLarge(t *testing.T) {
	ts := newTestServer(t)
	postJSON(t, ts.URL+"/flags", `{"key":"big","enabled":true}`).Body.Close()

	big := strings.Repeat("a", maxBodyBytes+10)
	body := `{"enabled":true,"description":"` + big + `","rollout_percent":50}`

	resp := doJSON(t, http.MethodPut, ts.URL+"/flags/big", body)
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", resp.StatusCode)
	}
	e := decodeError(t, resp)
	if e["error"] == "" {
		t.Fatalf("expected non-empty error field")
	}
}
