#!/bin/bash

#GO=/opt/go/go/bin/go
GO=$(which go)

$GO run -race cmd/lax/main.go $@
