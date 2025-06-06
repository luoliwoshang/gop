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

package mod

import (
	"fmt"
	"os"

	"github.com/goplus/xgo/cmd/internal/base"
	"github.com/goplus/xgo/tool"
	"github.com/goplus/xgo/x/xgoenv"
)

// gop mod tidy
var CmdTidy = &base.Command{
	UsageLine: "gop mod tidy [-e -v]",
	Short:     "add missing and remove unused modules",
}

func init() {
	CmdTidy.Run = runTidy
}

func runTidy(cmd *base.Command, args []string) {
	err := tool.Tidy(".", xgoenv.Get())
	if err != nil {
		if tool.NotFound(err) {
			fmt.Fprintln(os.Stderr, "go.mod not found")
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
