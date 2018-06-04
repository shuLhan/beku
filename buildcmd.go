// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

type buildMode uint

const (
	buildModeDep buildMode = 1 << iota
	buildModeGdm
	buildModeMake
)

const (
	buildFileDep  = "Gopkg.toml"
	buildFileGdm  = "Godeps"
	buildFileMake = "Makefile"
)

var (
	buildCmdDep  = []string{"dep", "ensure"}
	buildCmdGdm  = []string{"gdm", "restore"}
	buildCmdMake = []string{"make"}
)
