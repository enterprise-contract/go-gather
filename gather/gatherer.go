package gather

import (
	"context"
	"fmt"

	"github.com/enterprise-contract/go-gather/metadata"
)

type Gatherer interface {
	Gather(ctx context.Context, src, dst string) (metadata.Metadata, error)
	Matcher(uri string) bool
}

var gatherers []Gatherer

func GetGatherer(uri string) (Gatherer, error) {
	for _, gatherer := range gatherers {
		if gatherer.Matcher(uri) {
			return gatherer, nil
		}
	}
	return nil, fmt.Errorf("no gatherer found for URI: %s", uri)
}

func RegisterGatherer(g Gatherer) {
	gatherers = append(gatherers, g)
}