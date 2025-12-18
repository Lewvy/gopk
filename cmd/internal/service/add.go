package service

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/lewvy/gopk/cmd/internal/data"
)

var PkgExp = regexp.MustCompile(`^v\d+$`)

func Add(args []string, db *data.Queries) error {
	var alias, url string

	url = args[0]

	url, _ = strings.CutPrefix(url, "https://")
	url, _ = strings.CutPrefix(url, "http://")

	if len(args) > 1 {
		alias = args[1]
	}
	if alias == "" {
		base := path.Base(url)

		for PkgExp.MatchString(base) {
			url = strings.TrimSuffix(url, "/"+base)
			base = path.Base(url)
			fmt.Printf("url: %s, base: %s\n", url, base)
		}
		alias = base

		if alias == "." || alias == "/" {
			alias = url
		}
	}

	a := &data.AddPackageWithoutVersionParams{
		Name: alias,
		Url:  url,
	}

	_, err := db.AddPackageWithoutVersion(context.Background(), *a)
	if err != nil {
		return err
	}

	return nil
}
