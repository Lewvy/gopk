package service

import (
	"context"
	"fmt"

	"github.com/lewvy/gopk/cmd/internal/data"
)

func List(q *data.Queries, limit int, sortByFreq bool) error {
	var packages []data.Package
	var err error

	if sortByFreq {
		packages, err = q.ListPackagesByFrequency(context.Background(), int64(limit))
	} else {
		packages, err = q.ListPackagesByLastUsed(context.Background(), int64(limit))
	}
	if err != nil {
		return err
	}
	fmt.Println(packages)
	return nil
}
