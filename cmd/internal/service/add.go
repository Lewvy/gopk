package service

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/lewvy/gopk/internal/data"
)

func Add(args []string, db *data.Queries) error {
	var alias, url string

	url = args[0]

	url, _ = strings.CutPrefix(url, "https://")
	url, _ = strings.CutPrefix(url, "http://")

	if len(args) > 1 {
		alias = args[1]
	}
	if alias == "" {
		alias = path.Base(url)
		if alias == "." || alias == "/" {
			alias = url
		}
	}
	a := &data.AddPackageWithoutVersionParams{
		Name: alias,
		Url:  url,
	}

	p, err := db.AddPackageWithoutVersion(context.Background(), *a)
	if err != nil {
		return err
	}
	fmt.Println(p)

	return nil
}
