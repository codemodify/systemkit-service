#!/bin/bash

# Remove unused modules from "go.mod / go.sum"
go mod tidy

# Clean up
export OUTPUTFOLDER=./temp-build
rm -rf $OUTPUTFOLDER

# Set compiler flags
export LDFLAGS="-w -s"

SUPPORTED_TARGETS=(
	"darwin,amd64"
	"linux,386"
	"linux,amd64"
	"linux,arm"
	"linux,arm64"
	"linux,mips"
	"linux,mipsle"
	"linux,mips64,true"
	"linux,mips64le,true"
	"windows,amd64"
	"dragonfly,amd64"
	"freebsd,386,true"
	"freebsd,amd64,true"
	"freebsd,arm,true"
	"netbsd,amd64,true"
	"netbsd,arm,true"
	"openbsd,amd64,true"
	"openbsd,arm,true"
	"openbsd,arm64,true"
	"netbsd,386,true"
	"openbsd,386,true"
)

clear
echo ""
echo "Build PARAMS"
echo "    -> OUTPUTFOLDER             : ${OUTPUTFOLDER}"
echo ""

for target in "${SUPPORTED_TARGETS[@]}"; do

  # get the OS/Arch array
  IFS=',' read -ra targ <<<"$target"
  os="${targ[0]}"
  arch="${targ[1]}"
  ext=""

  if [ $os = "windows" ]; then
    ext=".exe"
  fi

  # build
  UPX=YES
  if [ "${targ[2]}" = "true" ]; then
    UPX=NO
  fi

  printf "Building %-22s : UPX=$UPX" ${os}.${arch}
  echo ""

  fullOutputFilePath="$OUTPUTFOLDER/systemkit-service-cli.${os}.${arch}${ext}"
  GOOS=${os} GOARCH=${arch} go build -ldflags "${LDFLAGS}" -o ${fullOutputFilePath} .

  if [ $UPX = YES ]; then
    upx ${fullOutputFilePath} 1>/dev/null
  fi

done

echo ""

SUPPORTED_TARGETS_NOT=(
  "darwin,386"
  "linux,ppc64"
  "linux,ppc64le"
  "linux,s390x,true"
  "illumos,amd64"
  "windows,386"
  "aix,ppc64"
  "android,386"
  "android,amd64"
  "android,arm"
  "android,arm64"
  "darwin,arm"
  "darwin,arm64"
  "js,wasm"
  "plan9,386"
  "plan9,amd64"
  "plan9,arm"
  "solaris,amd64"
)
