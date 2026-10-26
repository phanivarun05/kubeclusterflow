package helm

import (
	"errors"
	"fmt"
	"time"

	"github.com/phanivarun05/kubeclusterflow/config"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// Helm Client
type HelmClient struct {
	actionConfig *action.Configuration
	namespace    string
}

// Constructor to initialize Helm action configuration linked to client-go
func NewHelmClient(kubeContext string, targetNamespace string) (*HelmClient, error) {

	// Setup Helm CLI settings
	envSettings := cli.New()
	envSettings.KubeConfig = config.GetKubeConfigPath()
	envSettings.KubeContext = kubeContext

	actionConfig := new(action.Configuration)

	// Initialize action Configuration using Kubernetes REST Client getter
	// Helm stores release revisions inside Kubernetes secrets by default ("secret" driver)
	err := actionConfig.Init(
		envSettings.RESTClientGetter(),
		targetNamespace,
		"secret",
		func(f string, v ...interface{}) {},
	)

	if err != nil {
		return nil, err
	}

	return &HelmClient{
		actionConfig: actionConfig,
		namespace:    targetNamespace,
	}, nil
}

// ----------------------------------------
// Deployment or Upgrade for a helm release
// ----------------------------------------
func (h *HelmClient) DeployOrUpgrade(releaseName, chartPath string, values map[string]interface{}) error {

	//Load and parse the Helm chart from local disk
	chart, err := loader.Load(chartPath)
	if err != nil {
		return fmt.Errorf("failed to load chart from %s: %w", chartPath, err)
	}

	// check release history if it already exists
	histClient := action.NewHistory(h.actionConfig)
	histClient.Max = 1
	_, historyErr := histClient.Run(releaseName)

	// Conditional branching: Install vs Upgrade
	if errors.Is(historyErr, driver.ErrReleaseNotFound) {

		// Fresh install
		installClient := action.NewInstall(h.actionConfig)
		installClient.ReleaseName = releaseName
		installClient.Namespace = h.namespace
		installClient.Timeout = 2 * time.Minute
		installClient.Wait = false

		rel, err := installClient.Run(chart, values)
		if err != nil {
			return fmt.Errorf("helm install failed: %w", err)
		}

		fmt.Print("Successfully installed release: ", rel.Name, " Revision: ", rel.Version)
	} else if historyErr == nil {

		// --- Existing Release Upgrade ---
		upgradeClient := action.NewUpgrade(h.actionConfig)
		upgradeClient.Namespace = h.namespace
		upgradeClient.Wait = false

		rel, err := upgradeClient.Run(releaseName, chart, values)
		if err != nil {
			return fmt.Errorf("helm upgrade failed: %w", err)
		}

		fmt.Print("Successfully upgraded release: ", rel.Name, " to Revision: ", rel.Version)
	} else {
		return fmt.Errorf("failed to check release history: %w", historyErr)
	}

	return nil
}
