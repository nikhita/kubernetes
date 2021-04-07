#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${KUBE_ROOT}/hack/lib/init.sh"
source "${KUBE_ROOT}/hack/lib/util.sh"

kube::golang::verify_go_version
kube::golang::setup_env


echo 'installing depstat'
pushd "${KUBE_ROOT}/hack/tools" >/dev/null
  GO111MODULE=on go install github.com/RinkiyaKeDad/depstat
popd >/dev/null

cd "${KUBE_ROOT}"

# <(...) is process substitution which converts output of a 
# command into a file like object which is what diff expects
# -s reports if the files are same
diff -s ${KUBE_ROOT}/dependency-stats.json <(depstat stats --json)
