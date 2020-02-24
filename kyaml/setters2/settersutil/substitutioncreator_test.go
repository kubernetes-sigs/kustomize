// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package settersutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/setters2"
)

func TestGetValuesForMarkersPositive(t *testing.T) {
	value1 := setters2.Value{
		Marker: "IMAGE",
	}

	value2 := setters2.Value{
		Marker: "VERSION",
	}

	values := []setters2.Value{value1, value2}

	c := SubstitutionCreator{
		Pattern:    "something/IMAGE::VERSION/otherthing/IMAGE::VERSION/",
		Values:     values,
		FieldValue: "something/nginx::0.1.0/otherthing/nginx::0.1.0/",
	}

	m, err := c.GetValuesForMarkers()

	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, m["IMAGE"], "nginx")
	assert.Equal(t, m["VERSION"], "0.1.0")
}

func TestGetValuesForMarkersDiffMarkerValues(t *testing.T) {
	value1 := setters2.Value{
		Marker: "IMAGE",
	}

	value2 := setters2.Value{
		Marker: "VERSION",
	}

	values := []setters2.Value{value1, value2}

	c := SubstitutionCreator{
		Pattern:    "something/IMAGE:VERSION/IMAGE",
		Values:     values,
		FieldValue: "something/nginx:0.1.0/ubuntu",
	}

	_, err := c.GetValuesForMarkers()

	if !assert.Error(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, err.Error(), "Same marker is found to have different values in field value.") {
		t.FailNow()
	}
}

func TestGetValuesForMarkersNoMatch(t *testing.T) {
	value1 := setters2.Value{
		Marker: "IMAGE",
	}

	value2 := setters2.Value{
		Marker: "VERSION",
	}

	values := []setters2.Value{value1, value2}

	c := SubstitutionCreator{
		Pattern:    "something/IMAGE:VERSION",
		Values:     values,
		FieldValue: "otherthing/nginx:0.1.0",
	}

	_, err := c.GetValuesForMarkers()

	if !assert.Error(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, err.Error(), "Unable to derive values for markers. Create setters for all markers and then try again.") {
		t.FailNow()
	}
}

func TestGetValuesForMarkersNoMatch2(t *testing.T) {
	value1 := setters2.Value{
		Marker: "IMAGE",
	}

	value2 := setters2.Value{
		Marker: "VERSION",
	}

	values := []setters2.Value{value1, value2}

	c := SubstitutionCreator{
		Pattern:    "something/IMAGE:VERSION/abc",
		Values:     values,
		FieldValue: "something/nginx:0.1.0/abcd",
	}

	_, err := c.GetValuesForMarkers()

	if !assert.Error(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, err.Error(), "Unable to derive values for markers. Create setters for all markers and then try again.") {
		t.FailNow()
	}
}

func TestGetValuesForMarkersSubStngMarkers(t *testing.T) {
	value1 := setters2.Value{
		Marker: "IMAGE",
	}

	value2 := setters2.Value{
		Marker: "VERSION",
	}

	value3 := setters2.Value{
		Marker: "MAGE",
	}

	values := []setters2.Value{value1, value2, value3}

	c := SubstitutionCreator{
		Pattern:    "something/IMAGE:VERSION/abc/MAGE",
		Values:     values,
		FieldValue: "something/nginx:0.1.0/abc/ubuntu",
	}

	m, err := c.GetValuesForMarkers()

	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, m["IMAGE"], "nginx")
	assert.Equal(t, m["VERSION"], "0.1.0")
	assert.Equal(t, m["MAGE"], "ubuntu")
}
