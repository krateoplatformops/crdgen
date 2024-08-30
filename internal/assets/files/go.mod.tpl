module {{ .module }}

go 1.21.6

require (
	github.com/krateoplatformops/provider-runtime v0.9.0
	k8s.io/apimachinery v0.31.0
	sigs.k8s.io/controller-runtime v0.19.0
	sigs.k8s.io/controller-tools v0.16.1
)
