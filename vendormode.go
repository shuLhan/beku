// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

type vendorMode uint

const (
	vendorModeDep vendorMode = 1 << iota
	vendorModeGovendor
)

const (
	vendorFileDep      = "Gopkg.toml"
	vendorFileGovendor = "vendor/vendor.json"
)

var (
	vendorCmdDep      = []string{"dep", "ensure"}
	vendorCmdGovendor = []string{"govendor", "sync"}
)
