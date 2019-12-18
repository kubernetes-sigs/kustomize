// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/base64"
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	corev1 "k8s.io/api/core/v1"
)

func Test_generateDockerConfigJson(t *testing.T) {
	type args struct {
		username string
		password string
		email    string
		server   string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "simplesecret", args: args{
				username: "dockeruser",
				password: "dockerpassword",
				email:    "docker@google.com",
				server:   "https://index.docker.io/v1/",
			},
			want:    []byte("{\"auths\":{\"https://index.docker.io/v1/\":{\"username\":\"dockeruser\",\"password\":\"dockerpassword\",\"email\":\"docker@google.com\",\"auth\":\"ZG9ja2VydXNlcjpkb2NrZXJwYXNzd29yZA==\"}}}"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateDockerConfigJson(tt.args.username, tt.args.password, tt.args.email, tt.args.server)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateDockerConfigJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateDockerConfigJson() got = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

var opaqueSecret = map[string]interface{}{
	"keyA": "valueA",
	"keyB": "valueB",
}

type opaqueSecretReceiver struct{}

func (opaqueSecretReceiver) getData(path string) (map[string]interface{}, error) {
	return opaqueSecret, nil
}

var dockerSecret = map[string]interface{}{
	"password": "securepass",
	"username": "user",
	"server": "https://index.docker.io/v1/",
	"email": "user@gmail.com",
}

var dockerSecretData = map[string]interface{}{
	".dockerconfigjson": "eyJhdXRocyI6eyJodHRwczovL2luZGV4LmRvY2tlci5pby92MS8iOnsidXNlcm5hbWUiOiJ1c2VyIiwicGFzc3dvcmQiOiJzZWN1cmVwYXNzIiwiZW1haWwiOiJ1c2VyQGdtYWlsLmNvbSIsImF1dGgiOiJkWE5sY2pwelpXTjFjbVZ3WVhOeiJ9fX0=",
}

type dockerSecretReceiver struct{}

func (dockerSecretReceiver) getData(path string) (map[string]interface{}, error) {
	return dockerSecret, nil
}

var factory = resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), nil)

func Test_plugin_buildSecretResource(t *testing.T) {
	type fields struct {
		rmf            *resmap.Factory
		ldr            ifc.Loader
		secretReceiver secretReceiver
		TypeMeta       types.TypeMeta
		ObjectMeta     types.ObjectMeta
	}
	type args struct {
		name       string
		namespace  string
		path       string
		secretType string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		secretType corev1.SecretType
	}{
		{
			name: "opaque secret",
			fields: fields{
				rmf:            factory,
				ldr:            nil,
				secretReceiver: opaqueSecretReceiver{},
				TypeMeta:       types.TypeMeta{},
				ObjectMeta:     types.ObjectMeta{},
			},
			args: args{
				name:       "opaquesecret",
				namespace:  "mynamespace",
				path:       "/secrets/mynamespace/opaquesecret",
				secretType: string(corev1.SecretTypeOpaque),
			},
			secretType: corev1.SecretTypeOpaque,
		}, {
			name: "docker secret",
			fields: fields{
				rmf:            factory,
				ldr:            nil,
				secretReceiver: dockerSecretReceiver{},
				TypeMeta:       types.TypeMeta{},
				ObjectMeta:     types.ObjectMeta{},
			},
			args: args{
				name:       "dockersecret",
				namespace:  "mynamespace",
				path:       "/secrets/mynamespace/dockersecret",
				secretType: string(corev1.SecretTypeDockerConfigJson),
			},
			secretType: corev1.SecretTypeDockerConfigJson,
		}, {
			name: "implicit docker secret",
			fields: fields{
				rmf:            factory,
				ldr:            nil,
				secretReceiver: dockerSecretReceiver{},
				TypeMeta:       types.TypeMeta{},
				ObjectMeta:     types.ObjectMeta{},
			},
			args: args{
				name:       "dockersecret",
				namespace:  "mynamespace",
				path:       "/secrets/mynamespace/dockersecret",
				secretType: "",
			},
			secretType: corev1.SecretTypeDockerConfigJson,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &plugin{
				rmf:            tt.fields.rmf,
				ldr:            tt.fields.ldr,
				secretReceiver: tt.fields.secretReceiver,
				TypeMeta:       tt.fields.TypeMeta,
				ObjectMeta:     tt.fields.ObjectMeta,
			}
			got, err := p.buildSecretResource(tt.args.name, tt.args.namespace, tt.args.path, tt.args.secretType)
			if err != nil || got == nil {
				t.Errorf("resource is failed to build: %s", err)
			}

			dataMap := got.Map()["data"].(map[string]interface{})
			if (tt.secretType == corev1.SecretTypeDockerConfigJson) {
				if !reflect.DeepEqual(dataMap, dockerSecretData) {
					t.Errorf("objects are not equal")
				}
			} else {
				var decodedMap = make(map[string]interface{})
				for k, v := range dataMap {
					switch v.(type) {
					case string:
						decoded, _ := base64.StdEncoding.DecodeString(v.(string))
						decodedMap[k] = string(decoded)
					}
				}

				if !reflect.DeepEqual(decodedMap, opaqueSecret) {
					t.Errorf("objects are not equal")
				}
			}
		})
	}
}
