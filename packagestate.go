// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

type packageState uint

const (
	packageStateNew packageState = 1 << iota
	packageStateLoad
	packageStateChange
	packageStateDirty
)
