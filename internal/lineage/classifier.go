package lineage

import (
	"strings"

	v1 "k8s.io/api/core/v1"
)

// represents deployment engine.
type DeploymentEngine string

const (
	EngineArgoCD     DeploymentEngine = "ArgoCD"
	EngineHelm       DeploymentEngine = "Helm"
	EngineStandalone DeploymentEngine = "Standalone"
)

// PodLineageInfo defines the structured classification metadata for a single Pod.
type PodLineageInfo struct {
	Engine     DeploymentEngine
	ParentName string
	Namespace  string
	PodName    string
	Phase      string
}

// ClassifyPodOwnership analyzes the ownership of a given Pod and returns a structured PodLineageInfo object.

func ClassifyPodOwnership(pod *v1.Pod) PodLineageInfo {

	// Extract metadata safely
	podName := pod.Name
	namespace := pod.Namespace
	phase := string(pod.Status.Phase)
	annotations := pod.Annotations
	labels := pod.Labels

	// Default values if unmanaged
	engine := EngineStandalone
	parentName := "none"

	// ----------------------------
	// Check for ArgoCD Annotations
	// ----------------------------
	if trackingID, exists := annotations["argocd.argoproj.io/tracking-id"]; exists {

		parts := strings.Split(trackingID, ":")
		if len(parts) > 0 {
			parentName = parts[0]
			engine = EngineArgoCD
			goto BUILD_RESULT
		}
	}

	// -------------------------------------
	// Check for ArgoCD Instance Annotations
	// -------------------------------------
	if _, exists := annotations["argocd.argoproj.io/instance"]; exists {
		parentName = annotations["argocd.argoproj.io/instance"]
		engine = EngineArgoCD
		goto BUILD_RESULT
	}

	// ---------------------------------------
	// Check for Helm & GitOps Combined Labels
	// ---------------------------------------
	if instanceLabel, exists := labels["app.kubernetes.io/instance"]; exists {
		if _, exists := labels["meta.helm.sh/release-name"]; exists {
			parentName = labels["meta.helm.sh/release-name"]
			engine = EngineHelm
		} else {
			parentName = instanceLabel
			engine = EngineArgoCD
		}
		goto BUILD_RESULT
	}

	// -----------------------------------
	// Check for Native Helm Release Label
	// -----------------------------------
	if _, exists := labels["meta.helm.sh/release-name"]; exists {
		parentName = labels["meta.helm.sh/release-name"]
		engine = EngineHelm
		goto BUILD_RESULT
	}

BUILD_RESULT:
	return PodLineageInfo{
		Engine:     engine,
		ParentName: parentName,
		Namespace:  namespace,
		PodName:    podName,
		Phase:      phase,
	}
}
