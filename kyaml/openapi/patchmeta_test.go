// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"reflect"
	"strings"
	"testing"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const equivalenceMaxDepth = 12

// TestPrecomputedPatchMetaEquivalence proves the precomputed table is
// merge-equivalent to the full parsed schema: at every path reachable from
// every builtin resource root, both sides report the same patch strategy and
// merge keys, and wherever the table prunes a subtree the full schema carries
// no patch strategy anywhere below it.
func TestPrecomputedPatchMetaEquivalence(t *testing.T) {
	ResetOpenAPI()
	defer ResetOpenAPI()

	// index completeness: every GVK in the full schema is in the table
	// (kustomization types come from a separate asset and are intentionally
	// not in the table; they fall through to the full-schema path)
	initSchema()
	for tm := range globalSchema.schemaByResourceType {
		if strings.HasPrefix(tm.APIVersion, "kustomize.config.k8s.io/") {
			continue
		}
		if _, ok := precomputedGVKToDef[tm]; !ok {
			t.Errorf("GVK %v present in full schema but missing from precomputedGVKToDef", tm)
		}
	}

	for tm, def := range precomputedGVKToDef {
		full := schemaForResourceTypeFull(tm)
		if full == nil {
			t.Errorf("GVK %v in table but not in full schema", tm)
			continue
		}
		pm := pmDefNode(def)
		seen := map[*spec.Schema]bool{}
		walkCompare(t, tm.APIVersion+"/"+tm.Kind, full, pm, equivalenceMaxDepth, seen)
	}
}

func walkCompare(t *testing.T, path string, full, pm *ResourceSchema, depth int, seen map[*spec.Schema]bool) {
	t.Helper()
	fullStrategy, fullKeys := full.PatchStrategyAndKeyList()
	if pm == nil {
		if fullStrategy != "" {
			t.Errorf("%s: full schema has strategy %q but table pruned this node", path, fullStrategy)
		}
		if hasPatchBelow(full, depth, map[*spec.Schema]bool{}) {
			t.Errorf("%s: table pruned a subtree that contains patch metadata", path)
		}
		return
	}
	pmStrategy, pmKeys := pm.PatchStrategyAndKeyList()
	if fullStrategy != pmStrategy || !reflect.DeepEqual(fullKeys, pmKeys) {
		t.Errorf("%s: full (%q,%v) != table (%q,%v)", path, fullStrategy, fullKeys, pmStrategy, pmKeys)
	}
	if depth == 0 || full.Schema == nil || seen[full.Schema] {
		return
	}
	seen[full.Schema] = true

	for fname := range full.Schema.Properties {
		walkCompare(t, path+"."+fname, full.Field(fname), pm.Field(fname), depth-1, seen)
	}
	if full.Schema.AdditionalProperties != nil && full.Schema.AdditionalProperties.Schema != nil {
		walkCompare(t, path+".*", full.Field("arbitraryKey"), pm.Field("arbitraryKey"), depth-1, seen)
	}
	if fullElems := full.Elements(); fullElems != nil {
		walkCompare(t, path+"[]", fullElems, pm.Elements(), depth-1, seen)
	}
}

// hasPatchBelow reports whether any node at or below rs carries a patch
// strategy in the full schema.
func hasPatchBelow(rs *ResourceSchema, depth int, seen map[*spec.Schema]bool) bool {
	if rs == nil || rs.Schema == nil || depth == 0 || seen[rs.Schema] {
		return false
	}
	seen[rs.Schema] = true
	if s, _ := rs.PatchStrategyAndKeyList(); s != "" {
		return true
	}
	for fname := range rs.Schema.Properties {
		if hasPatchBelow(rs.Field(fname), depth-1, seen) {
			return true
		}
	}
	if rs.Schema.AdditionalProperties != nil && rs.Schema.AdditionalProperties.Schema != nil {
		if hasPatchBelow(rs.Field("arbitraryKey"), depth-1, seen) {
			return true
		}
	}
	if elems := rs.Elements(); elems != nil {
		if hasPatchBelow(elems, depth-1, seen) {
			return true
		}
	}
	return false
}

// TestSchemaForResourceTypeDoesNotParseSwaggerByDefault proves the default
// lookup path serves builtin types from the precomputed table without ever
// parsing the embedded swagger.
func TestSchemaForResourceTypeDoesNotParseSwaggerByDefault(t *testing.T) {
	ResetOpenAPI()
	defer ResetOpenAPI()

	rs := PatchMetaSchemaForResourceType(yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})
	if rs == nil {
		t.Fatal("no schema for apps/v1 Deployment")
	}
	containers := rs.Lookup("spec", "template", "spec", "containers")
	if containers == nil {
		t.Fatal("no schema for containers")
	}
	if s, k := containers.PatchStrategyAndKeyList(); s != "merge" || !reflect.DeepEqual(k, []string{"name"}) {
		t.Fatalf("containers: got (%q,%v), want (merge,[name])", s, k)
	}
	ports := containers.Elements().Field("ports")
	if ports == nil {
		t.Fatal("no schema for container ports")
	}
	if s, k := ports.PatchStrategyAndKeyList(); s != "merge" || !reflect.DeepEqual(k, []string{"containerPort", "protocol"}) {
		t.Fatalf("ports: got (%q,%v), want (merge,[containerPort protocol])", s, k)
	}

	schemaLock.RLock()
	parsed := globalSchema.schemaInit
	schemaLock.RUnlock()
	if parsed {
		t.Fatal("default builtin lookup parsed the full swagger; expected table-only service")
	}
}
