package main

import (
	"context"
	"dagger/crossplane-function/internal/dagger"
)

func New() *CrossplaneFunction {
	return &CrossplaneFunction{
		RegistryConfig: dag.RegistryConfig(),
	}
}

type CrossplaneFunction struct {
	// +private
	RegistryConfig *dagger.RegistryConfig
}

// Add credentials for a registry.
func (m *CrossplaneFunction) WithRegistryAuth(
	// The address of the registry.
	address string,
	// The username to authenticate with.
	username string,
	// The environment variable name containing the password to authenticate with. NOT the password itself.
	secret *dagger.Secret,
) *CrossplaneFunction {
	m.RegistryConfig = m.RegistryConfig.WithRegistryAuth(address, username, secret)

	return m
}

// Uses unauthenticated access to a registry.
func (m *CrossplaneFunction) WithoutRegistryAuth(
	// The address of the registry.
	address string,
) *CrossplaneFunction {
	m.RegistryConfig = m.RegistryConfig.WithoutRegistryAuth(address)

	return m
}

// Build and push a Crossplane function.
func (m *CrossplaneFunction) BuildCrossplaneFunction(
	// The context to run the container in.
	ctx context.Context,
	// The directory containing the function code.
	directory *dagger.Directory,
	// The platform to build the function for. E.g. "linux/amd64".
	platform string,
	// The Docker socket to use. E.g. "/var/run/docker.sock".
	sock *dagger.Socket,
	// The image registry to push the function to.
	imageRegistry string,
	// The image tag to push the function with. E.g. "latest".
	imageTag string,
) *dagger.Container {
	tarFile := dag.Container().
		WithDirectory("/src", directory).
		WithWorkdir("/src").
		Directory("/src").
		DockerBuild(dagger.DirectoryDockerBuildOpts{
			Platform: dagger.Platform(platform),
		}).AsTarball()

	return dag.Container().
		From("alpine:latest").
		With(func(c *dagger.Container) *dagger.Container {
			return m.RegistryConfig.SecretMount("/root/.docker/config.json").Mount(c)
		}).
		WithUnixSocket("/var/run/docker.sock", sock).
		WithFile("/src/runtime", tarFile).
		WithDirectory("/src/package", directory.Directory("package")).
		WithoutEntrypoint().
		WithExec([]string{"apk", "update"}).
		WithExec([]string{"apk", "add", "curl"}).
		WithExec([]string{"curl", "-sL", "https://raw.githubusercontent.com/crossplane/crossplane/master/install.sh", "-o", "install.sh"}).
		WithExec([]string{"chmod", "+x", "install.sh"}).
		WithExec([]string{"chmod", "700", "/src/runtime"}).
		WithExec([]string{"./install.sh"}).
		WithExec([]string{"./crossplane", "xpkg", "build", "--package-root=/src/package", "--embed-runtime-image-tarball=/src/runtime", "--package-file=function.xpkg"}).
		WithExec([]string{"./crossplane", "xpkg", "push", "--package-files=function.xpkg", imageRegistry + ":" + imageTag})
}
