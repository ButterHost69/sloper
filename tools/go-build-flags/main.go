package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// go-build-flags prints the -ldflags snippet that injects the current git
// describe version into the version package. It is used by the Makefile:
//
//	go build -ldflags "$(go run ./tools/go-build-flags)"
func main() {
	version := git("describe", "--tags", "--always", "--dirty")
	if version == "" {
		version = "dev"
	}
	fmt.Printf("-X github.com/ButterHost69/sloper/internal/version.Version=%s", version)
}

func git(args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}