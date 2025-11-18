package handlers

import (
	"testing"

	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

func TestGoFeatureFlagClient_JSONVariation(t *testing.T) {
	client := &GoFeatureFlagClient{}
	_, err := client.JSONVariation("non_existent_flag", ffcontext.NewEvaluationContext("anonymous"), map[string]interface{}{})
	if err == nil {
		t.Error("expected an error but got nil")
	}
}
