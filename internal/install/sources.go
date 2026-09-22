// Copyright 2026 Perruer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package install

// KnownSources maps the provider names used in resource type prefixes to
// their registry addresses. Terraformer wrote most of these providers into
// required_providers without a source, which silently meant hashicorp/<name>
// and broke `init` for every provider not published by HashiCorp.
var KnownSources = map[string]string{
	"alicloud":      "aliyun/alicloud",
	"auth0":         "auth0/auth0",
	"aws":           "hashicorp/aws",
	"azuread":       "hashicorp/azuread",
	"azuredevops":   "microsoft/azuredevops",
	"azurerm":       "hashicorp/azurerm",
	"cloudflare":    "cloudflare/cloudflare",
	"commercetools": "labd/commercetools",
	"datadog":       "DataDog/datadog",
	"digitalocean":  "digitalocean/digitalocean",
	"fastly":        "fastly/fastly",
	"github":        "integrations/github",
	"gitlab":        "gitlabhq/gitlab",
	"gmailfilter":   "yamamoto-febc/gmailfilter",
	"google":        "hashicorp/google",
	"google-beta":   "hashicorp/google-beta",
	"grafana":       "grafana/grafana",
	"heroku":        "heroku/heroku",
	"honeycombio":   "honeycombio/honeycombio",
	"ibm":           "IBM-Cloud/ibm",
	"ionoscloud":    "ionos-cloud/ionoscloud",
	"keycloak":      "keycloak/keycloak",
	"kubernetes":    "hashicorp/kubernetes",
	"launchdarkly":  "launchdarkly/launchdarkly",
	"linode":        "linode/linode",
	"logzio":        "logzio/logzio",
	"mackerel":      "mackerelio-labs/mackerel",
	"metal":         "equinix/metal",
	"mikrotik":      "ddelnano/mikrotik",
	"myrasec":       "Myra-Security-GmbH/myrasec",
	"newrelic":      "newrelic/newrelic",
	"ns1":           "ns1-terraform/ns1",
	"octopusdeploy": "OctopusDeployLabs/octopusdeploy",
	"okta":          "okta/okta",
	"opal":          "opalsecurity/opal",
	"openstack":     "terraform-provider-openstack/openstack",
	"opsgenie":      "opsgenie/opsgenie",
	"pagerduty":     "PagerDuty/pagerduty",
	"panos":         "PaloAltoNetworks/panos",
	"rabbitmq":      "cyrilgdn/rabbitmq",
	"tencentcloud":  "tencentcloudstack/tencentcloud",
	"vault":         "hashicorp/vault",
	"vultr":         "vultr/vultr",
	"xenorchestra":  "vatesfr/xenorchestra",
	// Not mirrored by the OpenTofu registry. Yandex also runs its own mirror
	// at terraform-mirror.yandexcloud.net for users who configure it.
	"yandex": "registry.terraform.io/yandex-cloud/yandex",
}

// SourceFor returns the registry address for a provider name, falling back
// to the hashicorp namespace the way required_providers does.
func SourceFor(name string) Source {
	if s, ok := KnownSources[name]; ok {
		src, _ := ParseSource(s)
		return src
	}
	return Source{Namespace: "hashicorp", Type: name}
}
