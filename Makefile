## Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
## Use of this source code is governed by a BSD-style
## license that can be found in the LICENSE file.

SRC:=$(shell go list -f '{{$$d:=.Dir}} {{ range .GoFiles }}{{$$d}}/{{.}} {{end}}' ./...)
SRC_TEST:=$(shell go list -f '{{$$d:=.Dir}} {{ range .TestGoFiles }}{{$$d}}/{{.}} {{end}}' ./...)

COVER_OUT:=cover.out
COVER_HTML:=cover.html
CPU_PROF:=cpu.prof
MEM_PROF:=mem.prof

.PHONY: all install lint
.PHONE: test test.prof coverbrowse

all: install

install: test
	go install ./...

test: $(COVER_HTML)

test.prof:
	go test -cpuprofile $(CPU_PROF) -memprofile $(MEM_PROF) .

$(COVER_HTML): $(COVER_OUT)
	go tool cover -html=$< -o $@

$(COVER_OUT): $(SRC) $(SRC_TEST)
	go test -coverprofile=$@ ./...

coverbrowse: $(COVER_HTML)
	xdg-open $<

lint:
	gometalinter ./...
