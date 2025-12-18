package service

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/lewvy/gopk/cmd/internal/data"
)

func Get(pkgs []string, db *data.Queries) error {
	errs := []error{}
	args_map := make(map[string]struct{})
	for _, pkg := range pkgs {
		url, err := db.GetPackageURLByName(context.Background(), pkg)
		if err != nil {
			e := fmt.Errorf("error finding %s: %q", pkg, err)
			errs = append(errs, e)
			continue
		}
		args_map[url] = struct{}{}
	}
	args := make([]string, 0, len(args_map)+1)
	args = append(args, "get")

	for i := range args_map {
		args = append(args, i)
	}
	if len(args_map) == 0 {
		if len(errs) > 0 {
			fmt.Println("Errors encountered:", errs)
		}
		return fmt.Errorf("no valid packages to install")
	}

	cmd := exec.Command("go", args...)

	out, err := cmd.CombinedOutput()

	if len(errs) > 0 {
		fmt.Println("lookup errors:", errs)
	}
	if len(out) > 0 {

		fmt.Println(string(out))
	}
	if err != nil {
		return err
	}

	return nil
}
