// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package sets

type String map[string]interface{}
type StringList [][]string

func (s StringList) Insert(val []string) StringList {
	return append(s, val)
}

func (s StringList) Has(val []string) bool {
	if len(s) == 0 {
		return false
	}

	for _, v := range s {
		if len(v) != len(val) {
			continue
		}
		// check every value
		var same bool
		for i := range v {
			same = true
			if v[i] != val[i] {
				same = false
				break
			}
		}
		if same {
			return true
		}
	}
	return false
}

func (s String) Len() int {
	return len(s)
}

func (s String) List() []string {
	var val []string
	for k := range s {
		val = append(val, k)
	}
	return val
}

func (s String) Has(val string) bool {
	_, found := s[val]
	return found
}

func (s String) Insert(vals ...string) {
	for _, val := range vals {
		s[val] = nil
	}
}

func (s String) Difference(s2 String) String {
	s3 := String{}
	for k := range s {
		if _, found := s2[k]; !found {
			s3.Insert(k)
		}
	}
	return s3
}

func (s String) SymmetricDifference(s2 String) String {
	s3 := String{}
	for k := range s {
		if _, found := s2[k]; !found {
			s3.Insert(k)
		}
	}
	for k := range s2 {
		if _, found := s[k]; !found {
			s3.Insert(k)
		}
	}
	return s3
}

func (s String) Intersection(s2 String) String {
	s3 := String{}
	for k := range s {
		if _, found := s2[k]; found {
			s3.Insert(k)
		}
	}
	return s3
}
