module k8s.io/kubernetes/hack/tools

go 1.16

require (
	github.com/cespare/prettybench v0.0.0-20150116022406-03b8cfe5406c
	github.com/client9/misspell v0.3.4
	github.com/golangci/golangci-lint v1.36.0
	github.com/google/go-flow-levee v0.1.4-0.20201102181719-72c65d71b1d3
	gotest.tools v2.2.0+incompatible
	gotest.tools/gotestsum v0.3.5
	honnef.co/go/tools v0.0.1-2020.1.6
	k8s.io/klog/hack/tools v0.0.0-20210303110520-14dec3377f55
	k8s.io/repo-infra/cmd/depstat v0.0.0-20210409033219-2d6ef9e18972
	sigs.k8s.io/zeitgeist v0.2.0
)

replace k8s.io/repo-infra/cmd/depstat => github.com/nikhita/repo-infra/cmd/depstat v0.0.0-20210409033219-2d6ef9e18972
