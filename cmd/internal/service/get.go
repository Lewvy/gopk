package service

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/lewvy/gopk/cmd/internal/data"
)

func GetFromName(pkgs []string, db *data.Queries) error {
	rows, err := db.GetURLsByNames(context.Background(), pkgs)
	if err != nil {
		return fmt.Errorf("db error: %q", err)
	}

	foundMap := make(map[string]struct{})
	var urls []string

	for _, row := range rows {
		foundMap[row.Name] = struct{}{}
		urls = append(urls, row.Url)
	}

	var missing []string
	for _, req := range pkgs {
		if _, exists := foundMap[req]; !exists {
			missing = append(missing, req)
		}
	}

	if len(urls) == 0 {
		return fmt.Errorf("packages not found: %s", strings.Join(missing, ", "))
	}

	if err := runGoGet(urls); err != nil {
		return err
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing packages: %s", strings.Join(missing, ", "))
	}

	return nil
}

func GetFromUrl(urls []string) error {
	if len(urls) == 0 {
		return nil
	}
	return runGoGet(urls)
}

func runGoGet(urls []string) error {
	args := append([]string{"get"}, urls...)
	cmd := exec.Command("go", args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		output := strings.TrimSpace(string(out))
		if len(output) > 200 {
			output = output[:197] + "..."
		}
		return fmt.Errorf("install failed: %s", output)
	}
	return nil
}
