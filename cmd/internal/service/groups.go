package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/lewvy/gopk/cmd/internal/data"
)

func CreateGroup(q *data.Queries, name string) error {
	ctx := context.Background()
	_, err := q.CreateGroup(ctx, name)
	if err != nil {
		return err
	}
	return nil
}

func ListGroups(q *data.Queries) ([]data.Group, error) {
	ctx := context.Background()
	return q.ListGroups(ctx)
}

func ListPackagesByGroupOrderByFreq(ctx context.Context, q *data.Queries, group string) ([]data.Package, error) {
	pkgs, err := q.ListPackagesByGroup(ctx, group)
	if err != nil {
		return nil, err
	}
	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].Freq.Int64 > pkgs[j].Freq.Int64
	})
	return pkgs, nil

}

func ListPackagesByGroupOrderByLU(ctx context.Context, queries *data.Queries, groupName string) ([]data.Package, error) {
	pkgs, err := queries.ListPackagesByGroup(ctx, groupName)
	if err != nil {
		return nil, err
	}
	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].LastUsed.Time.Unix() > pkgs[j].LastUsed.Time.Unix()
	})
	return pkgs, nil
}

func AssignToGroup(q *data.Queries, pkgs []string, group string) error {
	ctx := context.Background()

	groupID, err := q.GetGroupIDByName(ctx, group)
	if err != nil {
		return fmt.Errorf("group not found: %s", group)
	}

	for _, url := range pkgs {
		pkgID, err := q.GetPackageIDByURL(ctx, url)
		if err != nil {
			return fmt.Errorf("package not found: %s", url)
		}

		if err := q.AssignPackageToGroup(ctx, data.AssignPackageToGroupParams{
			GroupID:   groupID,
			PackageID: pkgID,
		}); err != nil {
			return err
		}
	}

	return nil
}

func InstallGroup(ctx context.Context, q *data.Queries, groupName string) error {
	pkgs, err := ListPackagesByGroupOrderByFreq(ctx, q, groupName)
	if err != nil {
		return err
	}
	urls := []string{}
	for _, pkg := range pkgs {
		urls = append(urls, pkg.Url)
	}

	return GetFromUrl(urls)
}
