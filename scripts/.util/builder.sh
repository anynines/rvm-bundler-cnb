#!/usr/bin/env bash

set -eu
set -o pipefail

# shellcheck source=SCRIPTDIR/print.sh
source "$(dirname "${BASH_SOURCE[0]}")/print.sh"

function util::builder::stack::build() {
  util::print::title "Building bionic stack..."
  docker pull ubuntu:bionic

  scripts_dir=$(cd "$(dirname $0)" && pwd)
  dir=${scripts_dir}/../integration/builder

  base_build_n_run=anyninesgmbh/build-n-run:base

  cnb_base_build_n_run=anyninesgmbh/build-n-run:base-cnb

  docker build -t "${base_build_n_run}" "$dir/dockerfile"
  docker build --build-arg "base_image=${base_build_n_run}" -t "${cnb_base_build_n_run}"  "$dir/cnb"

}


function util::builder::builder::build() {
  util::print::title "Building builder..."
  pack builder create anyninesgmbh/builder:bionic --pull-policy=never --config "$dir/builder.toml" 
}