package service

import (
	"context"
	"database/sql"
	"errors"
	"path"
	"regexp"
	"strings"

	"github.com/lewvy/gopk/cmd/internal/data"
	"github.com/mattn/go-sqlite3"
)

var moduleVerRe = regexp.MustCompile(`^v\d+$`)
var ErrConstraintUnique = errors.New("package already exists")

func Add(url, name, version string, iflag, force bool, queries *data.Queries) error {

	url = normalizeURL(url)

	if name == "" {
		name = getAlias(url)
	}

	var err error

	if force {
		args := data.UpdatePackageByNameParams{
			Url:     url,
			Name:    name,
			Version: sql.NullString{Valid: true, String: version},
		}
		_, err = queries.UpdatePackageByName(context.Background(), args)

	} else {
		a := &data.AddPackageWithVersionParams{
			Name:    name,
			Url:     url,
			Version: sql.NullString{Valid: true, String: version},
		}

		_, err = queries.AddPackageWithVersion(context.Background(), *a)

	}

	if err != nil {
		if isUniqueConstraintErr(err) {
			return ErrConstraintUnique
		}
		return err
	}

	if iflag {
		return Get([]string{name}, queries)
	}

	return nil
}

func isUniqueConstraintErr(err error) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
	}

	return false
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
