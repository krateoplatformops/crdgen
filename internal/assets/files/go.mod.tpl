module {{ .module }}

go 1.24.0

require (
	github.com/krateoplatformops/provider-runtime v0.9.1
	k8s.io/apimachinery v0.33.0
	sigs.k8s.io/controller-runtime v0.20.0
	sigs.k8s.io/controller-tools v0.18.0
)
