package registry

import (
	expander "github.com/enterprise-contract/go-gather/expand"
	_ "github.com/enterprise-contract/go-gather/expand/bzip2"
	_ "github.com/enterprise-contract/go-gather/expand/tar"
	_ "github.com/enterprise-contract/go-gather/expand/zip"


)

func GetExpander(extension string) (expander.Expander) {
	return expander.GetExpander(extension)
}
