package service

import (
	"context"

	"github.com/lewvy/gopk/cmd/internal/data"
)

func DeletePackage(ctx context.Context, queries *data.Queries, pkgs []string) error {
	return queries.MarkDeleteByName(ctx, pkgs)
}

func DeleteGroup(ctx context.Context, queries *data.Queries, group data.Group) error {
	return queries.DeleteGroup(ctx, group.Name)

}
