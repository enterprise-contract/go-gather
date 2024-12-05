package registry

import (
	"github.com/enterprise-contract/go-gather/gather"
	_ "github.com/enterprise-contract/go-gather/gather/file"
	_ "github.com/enterprise-contract/go-gather/gather/git"
	_ "github.com/enterprise-contract/go-gather/gather/http"
	_ "github.com/enterprise-contract/go-gather/gather/oci"
	
)

func GetGatherer(uri string) (gather.Gatherer, error) {
	return gather.GetGatherer(uri)
}
