package webhooks

import (
	"fmt"
	"strings"
)

type git struct {
	branch string
}

func (g *git) validateBranch(ref string) error {
	r := strings.Split(ref, "/")
	branch := r[len(r)-1]

	if branch != g.branch {
		return fmt.Errorf("branch doesn't match, expected: %s, got: %s", g.branch, branch)
	}

	return nil
}
