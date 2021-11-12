package configmanager

import (
	"sync"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	webhookutil "k8s.io/apiserver/pkg/util/webhook"
	"k8s.io/client-go/rest"

	configv1alpha1 "github.com/karmada-io/karmada/pkg/apis/config/v1alpha1"
)

// WebhookAccessor provides a common interface to get webhook configuration.
type WebhookAccessor interface {
	// GetUID gets a string that uniquely identifies the webhook.
	GetUID() string
	// GetConfigurationName gets the name of the webhook configuration that owns this webhook.
	GetConfigurationName() string
	// GetName gets the webhook Name field.
	GetName() string
	// GetClientConfig gets the webhook ClientConfig field.
	GetClientConfig() admissionregistrationv1.WebhookClientConfig
	// GetRules gets the webhook Rules field.
	GetRules() []configv1alpha1.RuleWithOperations
	// GetFailurePolicy gets the webhook FailurePolicy field.
	GetFailurePolicy() *admissionregistrationv1.FailurePolicyType
	// GetTimeoutSeconds gets the webhook TimeoutSeconds field.
	GetTimeoutSeconds() *int32
	// GetExploreReviewVersions gets the webhook ExploreReviewVersions field.
	GetExploreReviewVersions() []string

	// GetRESTClient gets the webhook client.
	GetRESTClient(clientManager *webhookutil.ClientManager) (*rest.RESTClient, error)
}

type resourceExploringAccessor struct {
	*configv1alpha1.ResourceExploringWebhook
	uid               string
	configurationName string

	initClient sync.Once
	client     *rest.RESTClient
	clientErr  error
}

// NewResourceExploringAccessor create an accessor for webhook.
func NewResourceExploringAccessor(uid, configurationName string, hook *configv1alpha1.ResourceExploringWebhook) WebhookAccessor {
	return &resourceExploringAccessor{uid: uid, configurationName: configurationName, ResourceExploringWebhook: hook}
}

// GetUID gets a string that uniquely identifies the webhook.
func (a *resourceExploringAccessor) GetUID() string {
	return a.uid
}

// GetConfigurationName gets the name of the webhook configuration that owns this webhook.
func (a *resourceExploringAccessor) GetConfigurationName() string {
	return a.configurationName
}

// GetName gets the webhook Name field.
func (a *resourceExploringAccessor) GetName() string {
	return a.Name
}

// GetClientConfig gets the webhook ClientConfig field.
func (a *resourceExploringAccessor) GetClientConfig() admissionregistrationv1.WebhookClientConfig {
	return a.ClientConfig
}

// GetRules gets the webhook Rules field.
func (a *resourceExploringAccessor) GetRules() []configv1alpha1.RuleWithOperations {
	return a.Rules
}

// GetFailurePolicy gets the webhook FailurePolicy field.
func (a *resourceExploringAccessor) GetFailurePolicy() *admissionregistrationv1.FailurePolicyType {
	return a.FailurePolicy
}

// GetTimeoutSeconds gets the webhook TimeoutSeconds field.
func (a *resourceExploringAccessor) GetTimeoutSeconds() *int32 {
	return a.TimeoutSeconds
}

// GetExploreReviewVersions gets the webhook ExploreReviewVersions field.
func (a *resourceExploringAccessor) GetExploreReviewVersions() []string {
	return a.ExploreReviewVersions
}

// GetRESTClient gets the webhook client.
func (a *resourceExploringAccessor) GetRESTClient(clientManager *webhookutil.ClientManager) (*rest.RESTClient, error) {
	a.initClient.Do(func() {
		a.client, a.clientErr = clientManager.HookClient(hookClientConfigForWebhook(a.Name, a.ClientConfig))
	})
	return a.client, a.clientErr
}

// hookClientConfigForWebhook construct a webhookutil.ClientConfig using an admissionregistrationv1.WebhookClientConfig
// to access v1alpha1.ResourceExploringWebhook. webhookutil.ClientConfig is used to create a HookClient
// and the purpose of the config struct is to share that with other packages that need to create a HookClient.
func hookClientConfigForWebhook(hookName string, config admissionregistrationv1.WebhookClientConfig) webhookutil.ClientConfig {
	clientConfig := webhookutil.ClientConfig{Name: hookName, CABundle: config.CABundle}
	if config.URL != nil {
		clientConfig.URL = *config.URL
	}
	if config.Service != nil {
		clientConfig.Service = &webhookutil.ClientConfigService{
			Name:      config.Service.Name,
			Namespace: config.Service.Namespace,
		}
		if config.Service.Port != nil {
			clientConfig.Service.Port = *config.Service.Port
		} else {
			clientConfig.Service.Port = 443
		}
		if config.Service.Path != nil {
			clientConfig.Service.Path = *config.Service.Path
		}
	}
	return clientConfig
}
