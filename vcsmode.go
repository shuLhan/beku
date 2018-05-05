// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

//
// VCSMode define the mode of package's version control system (VCS).
// Currently only supporting Git.
//
type VCSMode uint

// List of VCS modes.
const (
	VCSModeGit VCSMode = 1 << iota
)
