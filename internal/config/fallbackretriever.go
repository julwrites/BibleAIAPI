package config

import (
	"context"
	"log"

	"github.com/thomaspoignant/go-feature-flag/retriever"
)

type FallbackRetriever struct {
	Primary   retriever.Retriever
	Secondary retriever.Retriever
}

func (r *FallbackRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	data, err := r.Primary.Retrieve(ctx)
	if err != nil {
		log.Printf("Primary retriever failed: %v. Falling back to secondary.", err)
		return r.Secondary.Retrieve(ctx)
	}
	return data, nil
}
