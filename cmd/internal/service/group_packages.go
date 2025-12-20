package service

import (
	"context"

	"github.com/lewvy/gopk/cmd/internal/data"
)

func RemovePackagesFromGroups(ctx context.Context, queries *data.Queries, pkgs map[data.Package]struct{}, group int64) error {

	pkgIDs := []int64{}

	for pkg := range pkgs {
		pkgIDs = append(pkgIDs, pkg.ID)
	}

	args := data.RemovePackagesFromGroupParams{
		GroupID:    group,
		PackageIds: pkgIDs,
	}
	return queries.RemovePackagesFromGroup(ctx, args)

}
