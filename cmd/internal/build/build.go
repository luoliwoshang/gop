/*
 * Copyright (c) 2021 The GoPlus Authors (goplus.org). All rights reserved.
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

// Package build implements the “gop build” command.
package build

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/goplus/gogen"
	"github.com/goplus/gop"
	"github.com/goplus/gop/cl"
	"github.com/goplus/gop/cmd/internal/base"
	"github.com/goplus/gop/x/gocmd"
	"github.com/goplus/gop/x/gopprojs"
)

// gop build
var Cmd = &base.Command{
	UsageLine: "gop build [-debug -o output] [packages]",
	Short:     "Build Go+ files",
}

var (
	flagDebug  = flag.Bool("debug", false, "print debug information")
	flagOutput = flag.String("o", "", "gop build output file")
	flag       = &Cmd.Flag
)

func init() {
	Cmd.Run = runCmd
}

func runCmd(cmd *base.Command, args []string) {
	pass := base.PassBuildFlags(cmd)
	err := flag.Parse(args)
	if err != nil {
		log.Panicln("parse input arguments failed:", err)
	}

	if *flagDebug {
		gogen.SetDebug(gogen.DbgFlagAll &^ gogen.DbgFlagComments)
		cl.SetDebug(cl.DbgFlagAll)
		cl.SetDisableRecover(true)
	}

	args = flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	proj, args, err := gopprojs.ParseOne(args...)
	if err != nil {
		log.Panicln(err)
	}
	if len(args) != 0 {
		log.Panicln("too many arguments:", args)
	}

	conf, err := gop.NewDefaultConf(".", gop.ConfFlagNoTestFiles)
	if err != nil {
		log.Panicln("gop.NewDefaultConf:", err)
	}
	defer conf.UpdateCache()

	confCmd := &gocmd.BuildConfig{Gop: conf.Gop}
	if *flagOutput != "" {
		output, err := filepath.Abs(*flagOutput)
		if err != nil {
			log.Panicln(err)
		}
		confCmd.Flags = []string{"-o", output}
	}
	confCmd.Flags = append(confCmd.Flags, pass.Args...)
	build(proj, conf, confCmd)
}

// proj:项目的文件列表
// conf:gop的位置，初始化一个文件树
func build(proj gopprojs.Proj, conf *gop.Config, build *gocmd.BuildConfig) {
	const flags = gop.GenFlagPrompt
	var obj string
	var err error
	// 判断是何种构建
	switch v := proj.(type) {
	case *gopprojs.DirProj:
		obj = v.Dir
		err = gop.BuildDir(obj, conf, build, flags)
	case *gopprojs.PkgPathProj:
		obj = v.Path
		err = gop.BuildPkgPath("", v.Path, conf, build, flags)
	case *gopprojs.FilesProj:
		err = gop.BuildFiles(v.Files, conf, build)
	default:
		log.Panicln("`gop build` doesn't support", reflect.TypeOf(v))
	}
	if gop.NotFound(err) {
		fmt.Fprintf(os.Stderr, "gop build %v: not found\n", obj)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		return
	}
	os.Exit(1)
}

// -----------------------------------------------------------------------------
