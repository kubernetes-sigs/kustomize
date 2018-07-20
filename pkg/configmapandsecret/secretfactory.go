package configmapandsecret

import (
	"context"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"os/exec"
	"path/filepath"
	"time"
)

// SecretFactory makes Secrets.
type SecretFactory struct {
	args types.SecretArgs
	fSys fs.FileSystem
}

// NewSecretFactory returns a new SecretFactory.
func NewSecretFactory(args types.SecretArgs, fSys fs.FileSystem) *SecretFactory {
	return &SecretFactory{args: args, fSys: fSys}
}

// MakeSecret returns a new secret.
func (f *SecretFactory) MakeSecret(wd string) (*corev1.Secret, error) {
	s := &corev1.Secret{}
	s.APIVersion = "v1"
	s.Kind = "Secret"
	s.Name = f.args.Name
	s.Type = corev1.SecretType(f.args.Type)
	if s.Type == "" {
		s.Type = corev1.SecretTypeOpaque
	}
	s.Data = map[string][]byte{}
	for k, v := range f.args.Commands {
		out, err := f.createSecretKey(wd, v)
		if err != nil {
			return nil, errors.Wrap(err, "createSecretKey")
		}
		s.Data[k] = out
	}
	return s, nil
}

// Run a command, return its output as the secret.
func (f *SecretFactory) createSecretKey(wd string, command string) ([]byte, error) {
	if !f.fSys.IsDir(wd) {
		wd = filepath.Dir(wd)
		if !f.fSys.IsDir(wd) {
			return nil, errors.New("not a directory: " + wd)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = wd
	return cmd.Output()
}
