package service

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/lewvy/gopk/cmd/internal/data"
)

func Get(pkgs []string, db *data.Queries) error {

	rows, err := db.GetURLsByNames(context.Background(), pkgs)
	if err != nil {
		return fmt.Errorf("db error: %q", err)
	}

	foundMap := make(map[string]struct{})
	for _, row := range rows {
		foundMap[row.Name] = struct{}{}
	}

	var missing []string
	for _, req := range pkgs {
		if _, exists := foundMap[req]; !exists {
			missing = append(missing, req)
		}
	}

	if len(rows) == 0 {
		return fmt.Errorf("packages not found: %s", strings.Join(missing, ", "))
	}

	args := []string{"get"}
	for _, row := range rows {
		args = append(args, row.Url)
	}

	cmd := exec.Command("go", args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		output := strings.TrimSpace(string(out))
		if len(output) > 200 {
			output = output[:197] + "..."
		}
		return fmt.Errorf("install failed: %s", output)
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing packages: %s", strings.Join(missing, ", "))
	}

	return nil
}
