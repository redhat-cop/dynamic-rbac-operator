module github.com/redhat-cop/dynamic-rbac-operator

go 1.15

require (
	github.com/go-logr/logr v0.1.0
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.10.1
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
)
