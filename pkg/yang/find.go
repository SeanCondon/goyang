// Copyright 2015 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yang

// This file has functions that search the AST for specified nodes.

import (
	"reflect"
	"strings"
)

// localPrefix returns the local prefix used by the containing (sub)module to
// refer to its own module.
func localPrefix(n Node) string {
	return RootNode(n).GetPrefix()
}

// trimLocalPrefix trims the current module's prefix from the given name. If the
// name is not prefixed with it, the same string is returned unchanged.
func trimLocalPrefix(n Node, name string) string {
	pfx := localPrefix(n)
	if pfx != "" {
		pfx += ":"
	}
	return strings.TrimPrefix(name, pfx)
}

// FindGrouping finds the grouping named name in one of the parent node's
// grouping fields, seen provides a list of the modules previously seen
// by FindGrouping during traversal.  If no parent has the named grouping,
// nil is returned. Imported and included modules are also checked.
func FindGrouping(n Node, name string, seen map[string]bool) *Grouping {
	name = trimLocalPrefix(n, name)
	for n != nil {
		// Grab the Grouping field of the underlying structure.  n is
		// always a pointer to a structure,
		e := reflect.ValueOf(n).Elem()
		v := e.FieldByName("Grouping")
		if v.IsValid() {
			for _, g := range v.Interface().([]*Grouping) {
				if g.Name == name {
					return g
				}
			}
		}
		v = e.FieldByName("Import")
		if v.IsValid() {
			for _, i := range v.Interface().([]*Import) {
				pname := strings.TrimPrefix(name, i.Prefix.Name+":")
				if pname == name {
					continue
				}
				if g := FindGrouping(i.Module, pname, seen); g != nil {
					return g
				}
			}
		}
		v = e.FieldByName("Include")
		if v.IsValid() {
			for _, i := range v.Interface().([]*Include) {
				if seen[i.Module.Name] {
					// Prevent infinite loops in the case that we have already looked at
					// this submodule. This occurs where submodules have include statements
					// in them, or there is a circular dependency.
					continue
				}
				seen[i.Module.Name] = true
				if g := FindGrouping(i.Module, name, seen); g != nil {
					return g
				}
			}
		}
		n = n.ParentNode()
	}
	return nil
}
