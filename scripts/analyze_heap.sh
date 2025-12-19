#!/bin/bash

BIN=$1
PROF=$2

go tool pprof -top -unit=mb $BIN $PROF
