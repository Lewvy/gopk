package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/lewvy/gopk/cmd/internal/data"
	"github.com/mattn/go-sqlite3"
)

var (
	moduleVerRe         = regexp.MustCompile(`^v\d+$`)
	ErrConstraintUnique = errors.New("package already exists")
	ErrNotFound         = errors.New("package not found in the registry")
)

func Add(url, name, version string, iflag, force, outToTUI bool, queries *data.Queries) error {
	url = normalizeURL(url)
	if name == "" {
		name = getAlias(url)
	}

	addParams := data.AddPackageWithVersionParams{
		Name:    name,
		Url:     url,
		Version: sql.NullString{Valid: true, String: version},
	}

	_, err := queries.AddPackageWithVersion(context.Background(), addParams)

	if err != nil {
		if isUniqueConstraintErr(err) {
			if force {
				updateParams := data.UpdatePackageByNameParams{
					Url:     url,
					Name:    name,
					Version: sql.NullString{Valid: true, String: version},
				}
				if _, err := queries.UpdatePackageByName(context.Background(), updateParams); err != nil {
					return fmt.Errorf("failed to force update: %w", err)
				}
			} else {
				return ErrConstraintUnique
			}
		} else {
			return err
		}
	}

	if iflag {
		return Get([]string{name}, outToTUI, queries)
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
