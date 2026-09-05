package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"featureflags/internal/model"
	"featureflags/internal/store"
)

func TestEvaluateFlagDisabled(t *testing.T) {
	flag := model.Flag{Key: "feature-x", Enabled: false, RolloutPercent: 100}
	if evaluateFlag(flag, "alice") {
		t.Error("disabled flag must evaluate to false")
	}
}

func TestEvaluateFlagRollout100(t *testing.T) {
	flag := model.Flag{Key: "feature-x", Enabled: true, RolloutPercent: 100}
	if !evaluateFlag(flag, "alice") {
		t.Error("rollout_percent 100 must evaluate to true")
	}
}

func TestEvaluateFlagRollout0(t *testing.T) {
	flag := model.Flag{Key: "feature-x", Enabled: true, RolloutPercent: 0}
	if evaluateFlag(flag, "alice") {
		t.Error("rollout_percent 0 must evaluate to false")
	}
}

func TestEvaluateFlagDeterministic(t *testing.T) {
	flag := model.Flag{Key: "feature-x", Enabled: true, RolloutPercent: 50}
	first := evaluateFlag(flag, "alice")
	for i := 0; i < 100; i++ {
		if got := evaluateFlag(flag, "alice"); got != first {
			t.Fatalf("evaluation not deterministic: %v then %v", first, got)
		}
	}
}

func TestEvaluateResponseShape(t *testing.T) {
	flag := model.Flag{Key: "feature-x", Enabled: true, RolloutPercent: 100}
	resp := evaluateResponse(flag, "alice")

	if got := resp["key"]; got != "feature-x" {
		t.Errorf("key = %v, want feature-x", got)
	}
	if got := resp["enabled"]; got != true {
		t.Errorf("enabled = %v, want true", got)
	}
	result, ok := resp["result"].(bool)
	if !ok {
		t.Fatalf("result is %T, want bool", resp["result"])
	}
	if !result {
		t.Error("result must be true for rollout_percent 100")
	}
	if len(resp) != 3 {
		t.Errorf("expected exactly 3 fields, got %d", len(resp))
	}
}

func TestEvaluateMissingUser(t *testing.T) {
	s := store.New()
	req := httptest.NewRequest(http.MethodGet, "/flags/feature-x/evaluate", nil)
	req.SetPathValue("key", "feature-x")
	rr := httptest.NewRecorder()

	Evaluate(s)(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestEvaluateEmptyUser(t *testing.T) {
	s := store.New()
	req := httptest.NewRequest(http.MethodGet, "/flags/feature-x/evaluate?user=", nil)
	req.SetPathValue("key", "feature-x")
	rr := httptest.NewRecorder()

	Evaluate(s)(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestEvaluateUnknownKey(t *testing.T) {
	s := store.New()
	req := httptest.NewRequest(http.MethodGet, "/flags/feature-x/evaluate?user=alice", nil)
	req.SetPathValue("key", "feature-x")
	rr := httptest.NewRecorder()

	Evaluate(s)(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}
