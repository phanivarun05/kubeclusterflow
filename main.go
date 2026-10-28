package main

import (
	// Add a blank identifier prefix to protect them from your IDE
	_ "helm.sh/helm/v3/pkg/action"
	_ "helm.sh/helm/v3/pkg/chart/loader"
	_ "helm.sh/helm/v3/pkg/cli"
	_ "helm.sh/helm/v3/pkg/storage/driver"
	_ metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Your code here
}
