// +build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Build mg.Namespace
type Lint mg.Namespace
type Test mg.Namespace
type Run mg.Namespace
type Clean mg.Namespace

// Build the frontend and backend
func (b Build) All() {
	mg.Deps(b.Backend, b.Frontend)
}

// Build the backend binaries
func (Build) Backend() error {
	cleanBackend()

	var version string
	var err error
	version = os.Getenv("VERSION")
	if version == "" {
		version, err = generateBackendVersion()
		if err != nil {
			return err
		}
	}

	flags := fmt.Sprintf(`-X go-webapp-example/internal/app.version=%s -s -w`, version)
	binary := "go-webapp-example"
	env := map[string]string{
		"CGO_ENABLED": "0",
		"GOOS":        "linux",
		"GOARCH":      "amd64",
	}

	if err := sh.RunWith(env, "go", "build", "-ldflags", flags, "-o", "output/"+binary); err != nil {
		return err
	}
	return nil
}

// Build the frontend
func (Build) Frontend() error {
	// Return this to really build your frontend.
	if true {
		return nil
	}
	env := map[string]string{"NODE_ENV": "production"}
	if err := sh.RunWith(env, "yarn", "--cwd", "web", "build", "--production"); err != nil {
		return err
	}
	return nil
}

// Lint all code
func (l Lint) All() {
	mg.Deps(l.Backend, l.Frontend)
}

// Lint all backend code
func (Lint) Backend() error {
	return sh.Run("golangci-lint", "run", "--fix")
}

// Lint all frontend code
func (Lint) Frontend() error {
	return sh.Run("yarn", "--cwd", "web", "lint")
}

// Test all code
func (t Test) All() {
	mg.Deps(t.Backend, t.Frontend)
}

// Run all backend unit tests
func (Test) Backend() error {
	return sh.Run("go", "test", "./...", "-test.short")
}

// Run all backend unit tests with race detection
func (Test) Race() error {
	return sh.Run("go", "test", "-race", "./...", "-test.short")
}

// Run all backend tests, include integration tests
func (Test) Integration() error {
	return sh.Run("go", "test", "./...")
}

// Run all backend system tests
func (Test) System() error {
	return sh.Run("go", "test", "./...")
}

// Run all frontend tests
func (Test) Frontend() error {
	return sh.Run("yarn", "--cwd", "web", "test:unit")
}

// Cleanup output directories
func (Clean) Filesystem() error {
	if err := cleanBackend(); err != nil {
		return err
	}
	return nil
}

// Remove all docker dependencies
func (Clean) Docker() error {
	sh.Run("pkill", "-e", "air")
	if err := sh.Run("docker-compose", "-f", "deployments/docker-compose.yml", "-f", "deployments/docker-compose.dev.yml", "down"); err != nil {
		return err
	}

	var output string
	var err error
	if output, err = sh.Output("docker", "ps"); err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

// Install all required dependencies
func (Run) Deps() error {
	if err := sh.Run("yarn", "--cwd", "web"); err != nil {
		return err
	}
	if err := sh.Run("go", "mod", "download"); err != nil {
		return err
	}
	return nil
}

// Start the docker stack
func (Run) Docker() error {
	return sh.Run("docker-compose", "-f", "deployments/docker-compose.yml", "-f", "deployments/docker-compose.dev.yml", "up", "-d")
}

// Start air
func (Run) Backend() error {
	return sh.Run("air")
}

// Run all DB migrations
func (Run) Migrate() error {
	if err := sh.Run("go", "run", "main.go", "migrate", "fresh"); err != nil {
		return err
	}
	return nil
}

// Run all code generations
func (Run) Generate() error {
	if err := sh.Run("rm", "-f", "internal/graphql/gqlresolvers/resolver.go"); err != nil {
		return err
	}
	if err := sh.Run("go", "generate", "./..."); err != nil {
		return err
	}
	if err := sh.Run("rm", "-f", "internal/graphql/gqlresolvers/resolver.go"); err != nil {
		return err
	}
	if err := sh.Run("rm", "-f", "internal/graphql/gqlresolvers/*.resolvers.go"); err != nil {
		return err
	}
	return nil
}

// Cleanup backend build output directories
func cleanBackend() error {
	return sh.Run("rm", "-rf", "output/*")
}

// Fetch the version information from git
func generateBackendVersion() (string, error) {
	commit, err := sh.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "no-repo", nil
	}
	branch, err := sh.Output("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "no-repo", nil
	}
	return fmt.Sprintf("%s-%s", commit, branch), nil
}
