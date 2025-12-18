package service

import (
	"context"
	"database/sql"
	"path"
	"regexp"
	"strings"

	"github.com/lewvy/gopk/cmd/internal/data"
)

var moduleVerRe = regexp.MustCompile(`^v\d+$`)

func Add(url, name, version string, iflag bool, queries *data.Queries) error {

	url = normalizeURL(url)

	if name == "" {
		name = getAlias(url)
	}

	var err error
	if version == "" {
		a := &data.AddPackageWithoutVersionParams{
			Name: name,
			Url:  url,
		}

		_, err = queries.AddPackageWithoutVersion(context.Background(), *a)
	} else {
		a := &data.AddPackageWithVersionParams{
			Name:    name,
			Url:     url,
			Version: sql.NullString{Valid: true, String: version},
		}

		_, err = queries.AddPackageWithVersion(context.Background(), *a)
	}
	if err != nil {
		return err
	}

	if iflag {
		return Get([]string{name}, queries)
	}

	return nil
}

func getAlias(u string) string {
	name := path.Base(u)

	for moduleVerRe.MatchString(name) {
		u = strings.TrimSuffix(u, "/"+name)
		name = path.Base(u)
	}

	if name == "." || name == "/" {
		return u
	}

	return name
}

func normalizeURL(u string) string {
	u, _ = strings.CutPrefix(u, "https://")
	u, _ = strings.CutPrefix(u, "http://")
	return u
}
