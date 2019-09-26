module sigs.k8s.io/kustomize/kustomize/v3

go 1.12

require (
	github.com/argoproj/argo v2.3.0-rc3.0.20190921165324-a317fbf1412c+incompatible
	github.com/argoproj/argo-rollouts v0.4.2
	github.com/keleustes/armada-crd v0.2.7
	// github.com/kubeflow/kubeflow/bootstrap/v3 v3.0.0-20190924054925-7a74c8752315
	github.com/openshift/api v3.9.1-0.20190911180052-9f80b7806f58+incompatible
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	istio.io/istio v0.0.0-20190914032905-41204513d2e8
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/kustomize/v3 v3.2.0
	sigs.k8s.io/yaml v1.1.0
)

replace (
	// Latest version of ./travis/xxx.sh force recompilation of mdrip
	// which modules conflicts with the kustomize ones
	github.com/gorilla/sessions => github.com/gorilla/sessions v1.2.0
	github.com/gorilla/websocket => github.com/gorilla/websocket v1.2.0
	github.com/monopole/mdrip => github.com/monopole/mdrip v0.2.48
	github.com/russross/blackfriday => github.com/russross/blackfriday v2.0.0+incompatible
	github.com/shurcooL/sanitized_anchor_name => github.com/shurcooL/sanitized_anchor_name v1.0.0

	// kubeflow causes some kind of loop
	// github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.0.5
	// github.com/kubeflow/kubeflow/bootstrap/v3 => github.com/prorates/kubeflow/bootstrap/v3 v3.0.0-20190925142407-584822919945

	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20181203042331-505ab145d0a9
	k8s.io/api => k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190918201827-3de75813f604
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190918200908-1e17798da8c1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190918202139-0b14c719ca62
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190918203125-ae665f80358a
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190918202959-c340507a5d48
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190918200425-ed2f0867c778
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190817025403-3ae76f584e79
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190918203248-97c07dcbb623
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190918201136-c3a845f1fbb2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190918202837-c54ce30c680e
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190918202429-08c8357f8e2d
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190918202713-c34a54b3ec8e
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190918202550-958285cf3eef
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190918203421-225f0541b3ea
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190918202012-3c1ca76f5bda
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190918201353-5cc279503896
	sigs.k8s.io/kustomize/v3 v3.2.0 => ../../kustomize

)
