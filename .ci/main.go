package main

import (
	"context"

	"dagger/nucleus-ci/internal/dagger"
)

// NucleusCi represents the entry point for the CI pipeline.
type NucleusCi struct{}

// Actions encapsulates our CI/CD pipeline steps. Defined here so both files can use it.
type Actions struct{}

// Actions provides access to the modular, pipeline-specific commands.
// By returning a pointer to Actions, Dagger will automatically nest these commands in the CLI.
func (m *NucleusCi) Actions() *Actions {
	return &Actions{}
}

// Returns a container that echoes whatever string argument is provided
func (m *NucleusCi) ContainerEcho(stringArg string) *dagger.Container {
	return dag.Container().From("alpine:latest").WithExec([]string{"echo", stringArg})
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *NucleusCi) GrepDir(ctx context.Context, directoryArg *dagger.Directory, pattern string) (string, error) {
	return dag.Container().
		From("alpine:latest").
		WithMountedDirectory("/mnt", directoryArg).
		WithWorkdir("/mnt").
		WithExec([]string{"grep", "-R", pattern, "."}).
		Stdout(ctx)
}
