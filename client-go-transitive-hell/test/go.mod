module test2

go 1.14

require (
	github.com/ori-edge/deploy-controllers v0.0.0-00010101000000-000000000000
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	k8s.io/klog v1.0.0
	sigs.k8s.io/cluster-api v0.3.5
	sigs.k8s.io/controller-runtime v0.6.0
)

replace github.com/ori-edge/deploy-controllers => ../

replace k8s.io/api => k8s.io/api v0.17.4

replace k8s.io/apimachinery => k8s.io/apimachinery v0.17.4

replace k8s.io/client-go => k8s.io/client-go v0.17.4
