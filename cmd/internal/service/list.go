package service

import (
	"context"

	"github.com/lewvy/gopk/cmd/internal/data"
)

func List(q *data.Queries, limit int, sortByFreq bool) ([]data.Package, error) {
	var packages []data.Package
	var err error

	if sortByFreq {
		packages, err = q.ListPackagesByFrequency(context.Background(), int64(limit))
	} else {
		packages, err = q.ListPackagesByLastUsed(context.Background(), int64(limit))
	}
	if err != nil {
		return nil, err
	}

	return packages, nil
}
