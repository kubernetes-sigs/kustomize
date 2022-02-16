// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:build !no_go_plugins
// +build !no_go_plugins

package loader

import (
	"fmt"
	"log"
	"plugin"
	"reflect"

	"github.com/pkg/errors"

	"sigs.k8s.io/kustomize/api/internal/plugins/utils"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

func (l *Loader) loadGoPlugin(id resid.ResId, absPath string) (resmap.Configurable, error) {
	regId := relativePluginPath(id)
	if c, ok := registry[regId]; ok {
		return copyPlugin(c), nil
	}
	if !utils.FileExists(absPath) {
		return nil, fmt.Errorf(
			"expected file with Go object code at: %s", absPath)
	}
	log.Printf("Attempting plugin load from '%s'", absPath)
	p, err := plugin.Open(absPath)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s fails to load", absPath)
	}
	symbol, err := p.Lookup(konfig.PluginSymbol)
	if err != nil {
		return nil, errors.Wrapf(
			err, "plugin %s doesn't have symbol %s",
			regId, konfig.PluginSymbol)
	}
	c, ok := symbol.(resmap.Configurable)
	if !ok {
		return nil, fmt.Errorf("plugin '%s' not configurable", regId)
	}
	registry[regId] = c
	return copyPlugin(c), nil
}

func copyPlugin(c resmap.Configurable) resmap.Configurable {
	indirect := reflect.Indirect(reflect.ValueOf(c))
	newIndirect := reflect.New(indirect.Type())
	newIndirect.Elem().Set(reflect.ValueOf(indirect.Interface()))
	newNamed := newIndirect.Interface()
	return newNamed.(resmap.Configurable)
}
