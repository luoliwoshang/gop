/*
 * Copyright (c) 2025 The GoPlus Authors (goplus.org). All rights reserved.
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

package reflect

import (
	"reflect"

	"github.com/goplus/gop/tpl/variant"
)

// -----------------------------------------------------------------------------

// Type returns the type of an value.
func Type(args ...any) any {
	if len(args) != 1 {
		panic("type: arity mismatch")
	}
	return reflect.TypeOf(variant.Eval(args[0]))
}

// -----------------------------------------------------------------------------

func init() {
	mod := variant.NewModule("reflect")
	mod.Insert("type", Type)
}

// -----------------------------------------------------------------------------
