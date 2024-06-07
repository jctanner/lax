#!/bin/bash

#GO=/opt/go/go/bin/go
GO=$(which go)

$GO run cmd/lax/main.go $@
