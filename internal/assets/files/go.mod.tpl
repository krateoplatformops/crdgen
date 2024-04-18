module {{ .module }}

go 1.21.6

require (
	k8s.io/apimachinery v0.29.3
	sigs.k8s.io/controller-runtime v0.17.3
	sigs.k8s.io/controller-tools v0.14.0
)
