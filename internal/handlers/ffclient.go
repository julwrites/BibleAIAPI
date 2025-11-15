package handlers

import (
	gofeatureflag "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

type FFClient interface {
	JSONVariation(flagKey string, context ffcontext.EvaluationContext, defaultValue map[string]interface{}) (map[string]interface{}, error)
}

type GoFeatureFlagClient struct{}

func (c *GoFeatureFlagClient) JSONVariation(flagKey string, context ffcontext.EvaluationContext, defaultValue map[string]interface{}) (map[string]interface{}, error) {
	return gofeatureflag.JSONVariation(flagKey, context, defaultValue)
}
