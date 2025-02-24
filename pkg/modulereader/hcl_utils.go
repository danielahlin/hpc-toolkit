// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package modulereader

import (
	"fmt"
	"hpc-toolkit/pkg/sourcereader"
	"log"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/zclconf/go-cty/cty"
)

// getHCLInfo is wrapped by SourceReader interface which supports multiple
// sources and stores remote modules locally, so the given source parameter to
// getHCLInfo is only a local path.
func getHCLInfo(source string) (ModuleInfo, error) {
	var module *tfconfig.Module
	ret := ModuleInfo{}

	if sourcereader.IsEmbeddedPath(source) {
		if !tfconfig.IsModuleDirOnFilesystem(tfconfig.WrapFS(ModuleFS), source) {
			return ret, fmt.Errorf("Source is not a terraform or packer module: %s", source)
		}
		module, _ = tfconfig.LoadModuleFromFilesystem(tfconfig.WrapFS(ModuleFS), source)
	} else {
		fileInfo, err := os.Stat(source)
		if os.IsNotExist(err) {
			return ret, fmt.Errorf("Source to module does not exist: %s", source)
		}
		if err != nil {
			return ret, fmt.Errorf("Failed to read source of module: %s", source)
		}
		if !fileInfo.IsDir() {
			return ret, fmt.Errorf("Source of module must be a directory: %s", source)
		}
		if !tfconfig.IsModuleDir(source) {
			return ret, fmt.Errorf("Source is not a terraform or packer module: %s", source)
		}
		module, _ = tfconfig.LoadModule(source)
	}

	var vars []VarInfo
	var outs []VarInfo
	for _, v := range module.Variables {
		vInfo := VarInfo{
			Name:        v.Name,
			Type:        v.Type,
			Description: v.Description,
			Default:     v.Default,
			Required:    v.Required,
		}
		vars = append(vars, vInfo)
	}
	ret.Inputs = vars
	for _, v := range module.Outputs {
		vInfo := VarInfo{
			Name:        v.Name,
			Description: v.Description,
		}
		outs = append(outs, vInfo)
	}
	ret.Outputs = outs
	return ret, nil
}

// Transforms HCL type string into cty.Type
func getCtyType(hclType string) (cty.Type, error) {
	expr, diags := hclsyntax.ParseExpression([]byte(hclType), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return cty.Type{}, diags
	}
	typ, diags := typeexpr.TypeConstraint(expr)
	if diags.HasErrors() {
		return cty.Type{}, diags
	}
	return typ, nil
}

// NormalizeType attempts to bring semantically equal types to equal string representation. E.g.:
//
//	 NormalizeType("object({count=number,kind=string})")
//		== NormalizeType("object({kind=string,count=number})").
//
// Additionally it removes any comments, whitespaces and line breaks.
//
// This method is fail-safe, if error arises passed type will be returned without changes.
func NormalizeType(hclType string) string {
	ctyType, err := getCtyType(hclType)
	if err != nil {
		log.Printf("Failed to parse HCL type='%s', got %v", hclType, err)
		return hclType
	}
	return typeexpr.TypeString(ctyType)
}
