/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testing

import (
	"math/rand"
	"testing"

	admissionv1 "sigs.k8s.io/kustomize/pseudo/k8s/api/admission/v1"
	admissionv1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/admission/v1beta1"
	admissionregv1 "sigs.k8s.io/kustomize/pseudo/k8s/api/admissionregistration/v1"
	admissionregv1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/admissionregistration/v1beta1"
	appsv1 "sigs.k8s.io/kustomize/pseudo/k8s/api/apps/v1"
	appsv1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/apps/v1beta1"
	appsv1beta2 "sigs.k8s.io/kustomize/pseudo/k8s/api/apps/v1beta2"
	authenticationv1 "sigs.k8s.io/kustomize/pseudo/k8s/api/authentication/v1"
	authenticationv1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/authentication/v1beta1"
	authorizationv1 "sigs.k8s.io/kustomize/pseudo/k8s/api/authorization/v1"
	authorizationv1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/authorization/v1beta1"
	autoscalingv1 "sigs.k8s.io/kustomize/pseudo/k8s/api/autoscaling/v1"
	autoscalingv2beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/autoscaling/v2beta1"
	autoscalingv2beta2 "sigs.k8s.io/kustomize/pseudo/k8s/api/autoscaling/v2beta2"
	batchv1 "sigs.k8s.io/kustomize/pseudo/k8s/api/batch/v1"
	batchv1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/batch/v1beta1"
	batchv2alpha1 "sigs.k8s.io/kustomize/pseudo/k8s/api/batch/v2alpha1"
	certificatesv1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/certificates/v1beta1"
	coordinationv1 "sigs.k8s.io/kustomize/pseudo/k8s/api/coordination/v1"
	coordinationv1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/coordination/v1beta1"
	corev1 "sigs.k8s.io/kustomize/pseudo/k8s/api/core/v1"
	eventsv1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/events/v1beta1"
	extensionsv1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/extensions/v1beta1"
	imagepolicyv1alpha1 "sigs.k8s.io/kustomize/pseudo/k8s/api/imagepolicy/v1alpha1"
	networkingv1 "sigs.k8s.io/kustomize/pseudo/k8s/api/networking/v1"
	networkingv1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/networking/v1beta1"
	nodev1alpha1 "sigs.k8s.io/kustomize/pseudo/k8s/api/node/v1alpha1"
	nodev1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/node/v1beta1"
	policyv1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/policy/v1beta1"
	rbacv1 "sigs.k8s.io/kustomize/pseudo/k8s/api/rbac/v1"
	rbacv1alpha1 "sigs.k8s.io/kustomize/pseudo/k8s/api/rbac/v1alpha1"
	rbacv1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/rbac/v1beta1"
	schedulingv1 "sigs.k8s.io/kustomize/pseudo/k8s/api/scheduling/v1"
	schedulingv1alpha1 "sigs.k8s.io/kustomize/pseudo/k8s/api/scheduling/v1alpha1"
	schedulingv1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/scheduling/v1beta1"
	settingsv1alpha1 "sigs.k8s.io/kustomize/pseudo/k8s/api/settings/v1alpha1"
	storagev1 "sigs.k8s.io/kustomize/pseudo/k8s/api/storage/v1"
	storagev1alpha1 "sigs.k8s.io/kustomize/pseudo/k8s/api/storage/v1alpha1"
	storagev1beta1 "sigs.k8s.io/kustomize/pseudo/k8s/api/storage/v1beta1"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/pseudo/k8s/apimachinery/pkg/api/apitesting/fuzzer"
	"sigs.k8s.io/kustomize/pseudo/k8s/apimachinery/pkg/api/apitesting/roundtrip"
	genericfuzzer "sigs.k8s.io/kustomize/pseudo/k8s/apimachinery/pkg/apis/meta/fuzzer"
	"sigs.k8s.io/kustomize/pseudo/k8s/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/pseudo/k8s/apimachinery/pkg/runtime/serializer"
)

var groups = []runtime.SchemeBuilder{
	admissionv1beta1.SchemeBuilder,
	admissionv1.SchemeBuilder,
	admissionregv1beta1.SchemeBuilder,
	admissionregv1.SchemeBuilder,
	appsv1beta1.SchemeBuilder,
	appsv1beta2.SchemeBuilder,
	appsv1.SchemeBuilder,
	authenticationv1beta1.SchemeBuilder,
	authenticationv1.SchemeBuilder,
	authorizationv1beta1.SchemeBuilder,
	authorizationv1.SchemeBuilder,
	autoscalingv1.SchemeBuilder,
	autoscalingv2beta1.SchemeBuilder,
	autoscalingv2beta2.SchemeBuilder,
	batchv2alpha1.SchemeBuilder,
	batchv1beta1.SchemeBuilder,
	batchv1.SchemeBuilder,
	certificatesv1beta1.SchemeBuilder,
	coordinationv1.SchemeBuilder,
	coordinationv1beta1.SchemeBuilder,
	corev1.SchemeBuilder,
	eventsv1beta1.SchemeBuilder,
	extensionsv1beta1.SchemeBuilder,
	imagepolicyv1alpha1.SchemeBuilder,
	networkingv1.SchemeBuilder,
	networkingv1beta1.SchemeBuilder,
	nodev1alpha1.SchemeBuilder,
	nodev1beta1.SchemeBuilder,
	policyv1beta1.SchemeBuilder,
	rbacv1alpha1.SchemeBuilder,
	rbacv1beta1.SchemeBuilder,
	rbacv1.SchemeBuilder,
	schedulingv1alpha1.SchemeBuilder,
	schedulingv1beta1.SchemeBuilder,
	schedulingv1.SchemeBuilder,
	settingsv1alpha1.SchemeBuilder,
	storagev1alpha1.SchemeBuilder,
	storagev1beta1.SchemeBuilder,
	storagev1.SchemeBuilder,
}

func TestRoundTripExternalTypes(t *testing.T) {
	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)
	for _, builder := range groups {
		require.NoError(t, builder.AddToScheme(scheme))
	}
	seed := rand.Int63()
	// I'm only using the generic fuzzer funcs, but at some point in time we might need to
	// switch to specialized. For now we're happy with the current serialization test.
	fuzzer := fuzzer.FuzzerFor(genericfuzzer.Funcs, rand.NewSource(seed), codecs)

	roundtrip.RoundTripExternalTypes(t, scheme, codecs, fuzzer, nil)
}

func TestCompatibility(t *testing.T) {
	scheme := runtime.NewScheme()
	for _, builder := range groups {
		require.NoError(t, builder.AddToScheme(scheme))
	}
	roundtrip.NewCompatibilityTestOptions(scheme).Complete(t).Run(t)
}
