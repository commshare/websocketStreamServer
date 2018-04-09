#!/bin/bash

set -x 
CWD=`pwd`

echo $CWD

PA=$CWD:$CWD/../:$CWD/../../:$GOPATH

echo $PA

export GOPATH="$PA"

go build -o main.exe main.go



