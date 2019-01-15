package digest

import (
	"context"

	"github.com/containers/image/docker"
	"github.com/containers/image/docker/reference"
	"github.com/containers/image/manifest"
	"github.com/containers/image/pkg/docker/config"
	imageTypes "github.com/containers/image/types"
)

func Fetch(name, tag string) (string, error) {
	named, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return "", err
	}

	tagged, err := reference.WithTag(named, tag)
	if err != nil {
		return "", err
	}

	ref, err := docker.NewReference(tagged)
	if err != nil {
		return "", err
	}

	username, password, err := config.GetAuthentication(nil, reference.Domain(named))
	if err != nil {
		return "", err
	}

	ctx := context.Background()

	img, err := ref.NewImage(ctx, &imageTypes.SystemContext{
		OSChoice: "linux",
		DockerAuthConfig: &imageTypes.DockerAuthConfig{
			Username: username, Password: password,
		},
	})
	if err != nil {
		return "", err
	}

	defer img.Close()

	rawManifest, _, err := img.Manifest(ctx)
	if err != nil {
		return "", err
	}

	digest, err := manifest.Digest(rawManifest)
	if err != nil {
		return "", err
	}

	return digest.String(), nil
}
