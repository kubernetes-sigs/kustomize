// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package parser contains implementations of the framework.TemplateParser and framework.SchemaParser interfaces.
// Typically, you would use these in a framework.TemplateProcessor.
//
// Example:
//
//	processor := framework.TemplateProcessor{
//		ResourceTemplates: []framework.ResourceTemplate{{
//			Templates: parser.TemplateFiles("path/to/templates"),
//		}},
//		PatchTemplates: []framework.PatchTemplate{
//			&framework.ResourcePatchTemplate{
//				Templates: parser.TemplateFiles("path/to/patches/ingress.template.yaml"),
//			},
//		},
//		AdditionalSchemas: parser.SchemaFiles("path/to/crd-schemas"),
//	}
//
//
// All the parser in this file are compatible with embed.FS filesystems. To load from an embed.FS
// instead of local disk, use `.FromFS`. For example, if you embed filesystems as follows:
//
// //go:embed resources/* patches/*
// var templateFS embed.FS
// //go:embed schemas/*
// var schemaFS embed.FS
//
// Then you can use them like so:
//
//	processor := framework.TemplateProcessor{
//		ResourceTemplates: []framework.ResourceTemplate{{
//			Templates: parser.TemplateFiles("resources").FromFS(templateFS),
//		}},
//		PatchTemplates: []framework.PatchTemplate{
//			&framework.ResourcePatchTemplate{
//				Templates: parser.TemplateFiles("patches/ingress.template.yaml").FromFS(templateFS),
//			},
//		},
//		AdditionalSchemas: parser.SchemaFiles("schemas").FromFS(schemaFS),
//	}
package parser
