/*
 * Copyright (c) 2021 The XGo Authors (xgo.dev). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// This file implements scopes and the objects they contain.

package ast

import (
	"bytes"
	"fmt"

	"github.com/goplus/xgo/token"
)

// A Scope maintains the set of named language entities declared
// in the scope and a link to the immediately surrounding (outer)
// scope.
type Scope struct {
	Outer   *Scope
	Objects map[string]*Object
}

// NewScope creates a new scope nested in the outer scope.
func NewScope(outer *Scope) *Scope {
	const n = 4 // initial scope capacity
	return &Scope{outer, make(map[string]*Object, n)}
}

// Lookup returns the object with the given name if it is
// found in scope s, otherwise it returns nil. Outer scopes
// are ignored.
func (s *Scope) Lookup(name string) *Object {
	return s.Objects[name]
}

// Insert attempts to insert a named object obj into the scope s.
// If the scope already contains an object alt with the same name,
// Insert leaves the scope unchanged and returns alt. Otherwise
// it inserts obj and returns nil.
func (s *Scope) Insert(obj *Object) (alt *Object) {
	if alt = s.Objects[obj.Name]; alt == nil {
		s.Objects[obj.Name] = obj
	}
	return
}

// Debugging support
func (s *Scope) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "scope %p {", s)
	if s != nil && len(s.Objects) > 0 {
		fmt.Fprintln(&buf)
		for _, obj := range s.Objects {
			fmt.Fprintf(&buf, "\t%s %s\n", obj.Kind, obj.Name)
		}
	}
	fmt.Fprintf(&buf, "}\n")
	return buf.String()
}

// ----------------------------------------------------------------------------
// Objects

// An Object describes a named language entity such as a package,
// constant, type, variable, function (incl. methods), or label.
//
// The Data fields contains object-specific data:
//
//	Kind    Data type         Data value
//	Pkg     *Scope            package scope
//	Con     int               iota for the respective declaration
type Object struct {
	Kind ObjKind
	Name string // declared name
	Decl any    // corresponding Field, XxxSpec, FuncDecl, LabeledStmt, AssignStmt, Scope; or nil
	Data any    // object-specific data; or nil
	Type any    // placeholder for type information; may be nil
}

// NewObj creates a new object of a given kind and name.
func NewObj(kind ObjKind, name string) *Object {
	return &Object{Kind: kind, Name: name}
}

// Pos computes the source position of the declaration of an object name.
// The result may be an invalid position if it cannot be computed
// (obj.Decl may be nil or not correct).
func (obj *Object) Pos() token.Pos {
	name := obj.Name
	switch d := obj.Decl.(type) {
	case *Field:
		for _, n := range d.Names {
			if n.Name == name {
				return n.Pos()
			}
		}
	case *ImportSpec:
		if d.Name != nil && d.Name.Name == name {
			return d.Name.Pos()
		}
		return d.Path.Pos()
	case *ValueSpec:
		for _, n := range d.Names {
			if n.Name == name {
				return n.Pos()
			}
		}
	case *TypeSpec:
		if d.Name.Name == name {
			return d.Name.Pos()
		}
	case *FuncDecl:
		if d.Name.Name == name {
			return d.Name.Pos()
		}
	case *LabeledStmt:
		if d.Label.Name == name {
			return d.Label.Pos()
		}
	case *AssignStmt:
		for _, x := range d.Lhs {
			if ident, isIdent := x.(*Ident); isIdent && ident.Name == name {
				return ident.Pos()
			}
		}
	case *Scope:
		// predeclared object - nothing to do for now
	}
	return token.NoPos
}

// ObjKind describes what an object represents.
type ObjKind int

// The list of possible Object kinds.
const (
	Bad ObjKind = iota // for error handling
	Pkg                // package
	Con                // constant
	Typ                // type
	Var                // variable
	Fun                // function or method
	Lbl                // label
	implicitBase

	ImplicitPkg = ObjKind(implicitBase)
	ImplicitFun = ObjKind(implicitBase + 1)
)

var objKindStrings = [...]string{
	Bad: "bad",
	Pkg: "package",
	Con: "const",
	Typ: "type",
	Var: "var",
	Fun: "func",
	Lbl: "label",

	ImplicitPkg: "package",
	ImplicitFun: "func",
}

func (kind ObjKind) String() string { return objKindStrings[kind] }
