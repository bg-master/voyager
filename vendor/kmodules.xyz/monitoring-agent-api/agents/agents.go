/*
Copyright The Kmodules Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package agents

import (
	"kmodules.xyz/monitoring-agent-api/agents/coreosprometheusoperator"
	"kmodules.xyz/monitoring-agent-api/agents/prometheusbuiltin"
	api "kmodules.xyz/monitoring-agent-api/api/v1"

	prom "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	ecs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
)

func New(at api.AgentType, k8sClient kubernetes.Interface, extClient ecs.ApiextensionsV1beta1Interface, promClient prom.MonitoringV1Interface) api.Agent {
	switch at {
	case api.AgentCoreOSPrometheus, api.DeprecatedAgentCoreOSPrometheus:
		return coreosprometheusoperator.New(at, k8sClient, extClient, promClient)
	case api.AgentPrometheusBuiltin:
		return prometheusbuiltin.New(k8sClient)
	}
	return nil
}
