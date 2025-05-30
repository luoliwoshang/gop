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

package ast

import (
	"sort"
	"strconv"

	"github.com/goplus/xgo/token"
)

// SortImports sorts runs of consecutive import lines in import blocks in f.
// It also removes duplicate imports when it is possible to do so without data loss.
func SortImports(fset *token.FileSet, f *File) {
	for _, d := range f.Decls {
		d, ok := d.(*GenDecl)
		if !ok || d.Tok != token.IMPORT {
			// Not an import declaration, so we're done.
			// Imports are always first.
			break
		}

		if !d.Lparen.IsValid() {
			// Not a block: sorted by default.
			continue
		}

		// Identify and sort runs of specs on successive lines.
		i := 0
		specs := d.Specs[:0]
		for j, s := range d.Specs {
			if j > i && lineAt(fset, s.Pos()) > 1+lineAt(fset, d.Specs[j-1].End()) {
				// j begins a new run. End this one.
				specs = append(specs, sortSpecs(fset, f, d.Specs[i:j])...)
				i = j
			}
		}
		specs = append(specs, sortSpecs(fset, f, d.Specs[i:])...)
		d.Specs = specs

		// Deduping can leave a blank line before the rparen; clean that up.
		if len(d.Specs) > 0 {
			lastSpec := d.Specs[len(d.Specs)-1]
			lastLine := lineAt(fset, lastSpec.Pos())
			rParenLine := lineAt(fset, d.Rparen)
			for rParenLine > lastLine+1 {
				rParenLine--
				fset.File(d.Rparen).MergeLine(rParenLine)
			}
		}
	}
}

func lineAt(fset *token.FileSet, pos token.Pos) int {
	return fset.PositionFor(pos, false).Line
}

func importPath(s Spec) string {
	t, err := strconv.Unquote(s.(*ImportSpec).Path.Value)
	if err == nil {
		return t
	}
	return ""
}

func importName(s Spec) string {
	n := s.(*ImportSpec).Name
	if n == nil {
		return ""
	}
	return n.Name
}

func importComment(s Spec) string {
	c := s.(*ImportSpec).Comment
	if c == nil {
		return ""
	}
	return c.Text()
}

// collapse indicates whether prev may be removed, leaving only next.
func collapse(prev, next Spec) bool {
	if importPath(next) != importPath(prev) || importName(next) != importName(prev) {
		return false
	}
	return prev.(*ImportSpec).Comment == nil
}

type posSpan struct {
	Start token.Pos
	End   token.Pos
}

type cgPos struct {
	left bool // true if comment is to the left of the spec, false otherwise.
	cg   *CommentGroup
}

func sortSpecs(fset *token.FileSet, f *File, specs []Spec) []Spec {
	// Can't short-circuit here even if specs are already sorted,
	// since they might yet need deduplication.
	// A lone import, however, may be safely ignored.
	if len(specs) <= 1 {
		return specs
	}

	// Record positions for specs.
	pos := make([]posSpan, len(specs))
	for i, s := range specs {
		pos[i] = posSpan{s.Pos(), s.End()}
	}

	// Identify comments in this range.
	begSpecs := pos[0].Start
	endSpecs := pos[len(pos)-1].End
	beg := fset.File(begSpecs).LineStart(lineAt(fset, begSpecs))
	endLine := lineAt(fset, endSpecs)
	endFile := fset.File(endSpecs)
	var end token.Pos
	if endLine == endFile.LineCount() {
		end = endSpecs
	} else {
		end = endFile.LineStart(endLine + 1) // beginning of next line
	}
	first := len(f.Comments)
	last := -1
	for i, g := range f.Comments {
		if g.End() >= end {
			break
		}
		// g.End() < end
		if beg <= g.Pos() {
			// comment is within the range [beg, end[ of import declarations
			if i < first {
				first = i
			}
			if i > last {
				last = i
			}
		}
	}

	var comments []*CommentGroup
	if last >= 0 {
		comments = f.Comments[first : last+1]
	}

	// Assign each comment to the import spec on the same line.
	importComments := map[*ImportSpec][]cgPos{}
	specIndex := 0
	for _, g := range comments {
		for specIndex+1 < len(specs) && pos[specIndex+1].Start <= g.Pos() {
			specIndex++
		}
		var left bool
		// A block comment can appear before the first import spec.
		if specIndex == 0 && pos[specIndex].Start > g.Pos() {
			left = true
		} else if specIndex+1 < len(specs) && // Or it can appear on the left of an import spec.
			lineAt(fset, pos[specIndex].Start)+1 == lineAt(fset, g.Pos()) {
			specIndex++
			left = true
		}
		s := specs[specIndex].(*ImportSpec)
		importComments[s] = append(importComments[s], cgPos{left: left, cg: g})
	}

	// Sort the import specs by import path.
	// Remove duplicates, when possible without data loss.
	// Reassign the import paths to have the same position sequence.
	// Reassign each comment to the spec on the same line.
	// Sort the comments by new position.
	sort.Slice(specs, func(i, j int) bool {
		ipath := importPath(specs[i])
		jpath := importPath(specs[j])
		if ipath != jpath {
			return ipath < jpath
		}
		iname := importName(specs[i])
		jname := importName(specs[j])
		if iname != jname {
			return iname < jname
		}
		return importComment(specs[i]) < importComment(specs[j])
	})

	// Dedup. Thanks to our sorting, we can just consider
	// adjacent pairs of imports.
	deduped := specs[:0]
	for i, s := range specs {
		if i == len(specs)-1 || !collapse(s, specs[i+1]) {
			deduped = append(deduped, s)
		} else {
			p := s.Pos()
			fset.File(p).MergeLine(lineAt(fset, p))
		}
	}
	specs = deduped

	// Fix up comment positions
	for i, s := range specs {
		s := s.(*ImportSpec)
		if s.Name != nil {
			s.Name.NamePos = pos[i].Start
		}
		s.Path.ValuePos = pos[i].Start
		s.EndPos = pos[i].End
		for _, g := range importComments[s] {
			for _, c := range g.cg.List {
				if g.left {
					c.Slash = pos[i].Start - 1
				} else {
					// An import spec can have both block comment and a line comment
					// to its right. In that case, both of them will have the same pos.
					// But while formatting the AST, the line comment gets moved to
					// after the block comment.
					c.Slash = pos[i].End
				}
			}
		}
	}

	sort.Slice(comments, func(i, j int) bool {
		return comments[i].Pos() < comments[j].Pos()
	})

	return specs
}
