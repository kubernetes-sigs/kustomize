// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Code generated for package kustomizationapi by go-bindata DO NOT EDIT. (@generated)
// sources:
// kustomizationapi/swagger.json
package kustomizationapi

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _kustomizationapiSwaggerJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xdc\x54\xb1\x6e\xdb\x30\x10\xdd\xfd\x15\x04\xdb\xd1\x52\xe0\xad\xf0\x56\x74\xe8\x10\x04\x08\x90\x6e\x45\x86\xb3\x7c\x52\xae\x92\x49\xf6\x78\x12\xea\x16\xfe\xf7\x42\xac\xa5\x88\xb6\xd4\xb4\x46\x1c\x24\x1e\x0c\x18\xd4\xdd\x7b\xbc\x7b\x8f\xef\xd7\x4c\x29\xbd\xc6\x9c\x0c\x09\x59\xe3\xf5\x52\xb5\x47\x4a\x69\xb2\x69\xf9\xc1\xa7\xe0\x28\x05\xe7\x7c\xda\x2c\xd2\x4f\xd6\xe4\x54\xdc\x80\xfb\xc8\xc5\x63\xa5\x52\xda\xb1\x75\xc8\x42\x38\x3c\x55\x4a\x7f\x46\x83\x0c\x62\xf9\xa0\x21\x7c\x7c\xcf\x98\xeb\xa5\xd2\xef\xae\x06\xfc\x57\x23\xb4\x31\x4a\x0f\xb1\xdb\xff\xdb\xcd\xbb\x6b\xc0\x7a\x1d\x50\xa0\xba\x1d\x5e\x28\x87\xca\x63\x5f\x24\x5b\x87\x2d\xad\x5d\x7d\xc3\x4c\x74\x7f\xfe\x23\x29\xeb\x15\xb2\x41\x41\x9f\x14\x6c\x6b\x97\x34\xc8\x9e\xac\x49\x4a\x32\x6b\xbd\x54\x5f\x7b\xea\x68\x8e\x50\xdb\x22\x96\xb5\x17\xbb\xa1\x9f\x98\x66\x61\x51\x61\x10\xb2\x3d\x45\xa8\xde\x63\xe9\x78\x97\x51\xc9\x9e\xb6\xad\x6a\x16\x2b\x14\x58\x1c\x0f\x7d\x3f\x1b\x8c\x3e\xa6\xd5\x1d\x66\x8c\xf2\x3a\x84\x7a\x9c\xae\xdb\x7e\x84\xdf\x29\xe2\x85\xc9\x14\x97\x22\xf0\x40\x80\xe7\x57\x77\x4a\xaf\x49\x81\x0d\x6c\xd0\x3b\xc8\xfe\x7d\xf9\xf3\xb8\xf9\x94\xbe\x15\x3e\x40\x43\x96\x4f\xe9\xbd\x6e\x6e\x81\xf8\xce\xd6\x9c\xe1\xe9\x8e\x8c\x51\x2e\xc4\x59\xb1\xf8\xcf\x6f\xae\xeb\xfd\x65\x40\xfe\x40\xf5\xe6\x62\xfc\x5e\x13\x63\x3c\x90\xfe\xb2\x75\x78\x83\x02\x1d\xd3\xfd\xfc\x29\x33\x66\x5d\xf6\xf5\x93\x1c\x0a\x4c\x82\x9b\x43\xd5\xff\x47\xf7\x38\x5d\x07\x20\xbb\xf9\x98\x11\x81\x19\xb6\xf1\x26\x23\x4d\x1d\x48\xf6\x90\x6c\x90\x0b\x4c\x4a\xdc\xb6\x2d\xe1\x4d\x3c\xd5\xe1\x85\x41\xb0\x08\x0d\xa1\x7b\xdc\xeb\x3e\x44\xc5\xd9\x96\x31\x48\xa2\x57\xb9\x89\x37\xfd\x18\xe3\xc7\x72\x86\xc7\x38\x91\x83\x93\x8f\xab\x22\x41\x86\xea\x28\x33\x27\x5c\x34\x95\xc5\x7f\x37\xc8\xa8\x8d\x73\xaa\x8e\xa3\xfa\xfc\xb4\x68\x9a\x97\x62\x7d\xdb\x4e\x8d\x9c\x74\xaa\x53\x67\xed\x6f\xf7\x3b\x00\x00\xff\xff\xfe\x97\xce\xec\x37\x0c\x00\x00")

func kustomizationapiSwaggerJsonBytes() ([]byte, error) {
	return bindataRead(
		_kustomizationapiSwaggerJson,
		"kustomizationapi/swagger.json",
	)
}

func kustomizationapiSwaggerJson() (*asset, error) {
	bytes, err := kustomizationapiSwaggerJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "kustomizationapi/swagger.json", size: 3127, mode: os.FileMode(420), modTime: time.Unix(1602011464, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"kustomizationapi/swagger.json": kustomizationapiSwaggerJson,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"kustomizationapi": &bintree{nil, map[string]*bintree{
		"swagger.json": &bintree{kustomizationapiSwaggerJson, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
