package configmapandsecret

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

// SecretFactory makes Secrets.
type SecretFactory struct {
	fSys fs.FileSystem
	wd   string
}

// NewSecretFactory returns a new SecretFactory.
func NewSecretFactory(fSys fs.FileSystem, wd string) *SecretFactory {
	return &SecretFactory{fSys: fSys, wd: wd}
}

// MakeSecret returns a new secret.
func (f *SecretFactory) MakeSecret(args types.SecretArgs) (*corev1.Secret, error) {
	s := &corev1.Secret{}
	s.APIVersion = "v1"
	s.Kind = "Secret"
	s.Name = args.Name
	s.Type = corev1.SecretType(args.Type)
	if s.Type == "" {
		s.Type = corev1.SecretTypeOpaque
	}
	s.Data = map[string][]byte{}
	for k, v := range args.Commands {
		out, err := f.createSecretKey(v)
		if err != nil {
			errMsg := fmt.Sprintf("createSecretKey: couldn't make secret %s for key %s", s.Name, k)
			return nil, errors.Wrap(err, errMsg)
		}
		s.Data[k] = out
	}
	return s, nil
}

// Run a command, return its output as the secret.
func (f *SecretFactory) createSecretKey(command string) ([]byte, error) {
	if !f.fSys.IsDir(f.wd) {
		f.wd = filepath.Dir(f.wd)
		if !f.fSys.IsDir(f.wd) {
			return nil, errors.New("not a directory: " + f.wd)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = f.wd
	return cmd.Output()
}
