// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

type debugMode uint

// List of debug levels.
const (
	DebugL1 debugMode = 1 << iota
	DebugL2
)
