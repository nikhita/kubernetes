// This is a generated file. Do not edit directly.

module k8s.io/kube-proxy

go 1.15

require (
	k8s.io/apimachinery v0.0.0
	k8s.io/component-base v0.0.0
)

replace (
	k8s.io/api => ../api
	k8s.io/apimachinery => ../apimachinery
	k8s.io/client-go => ../client-go
	k8s.io/component-base => ../component-base
	k8s.io/kube-openapi => github.com/nikhita/kube-openapi v0.0.0-20201202112845-d16cf5751bb0
	k8s.io/kube-proxy => ../kube-proxy
)
