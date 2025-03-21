// Copyright 2020, OpenTelemetry Authors
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

package kubeletstatsreceiver

import (
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"k8s.io/client-go/kubernetes"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
	kube "github.com/open-telemetry/opentelemetry-collector-contrib/internal/kubelet"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver/internal/kubelet"
)

var _ config.Receiver = (*Config)(nil)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`

	kube.ClientConfig `mapstructure:",squash"`
	confignet.TCPAddr `mapstructure:",squash"`

	// ExtraMetadataLabels contains list of extra metadata that should be taken from /pods endpoint
	// and put as extra labels on metrics resource.
	// No additional metadata is fetched by default, so there are no extra calls to /pods endpoint.
	// Supported values include container.id and k8s.volume.type.
	ExtraMetadataLabels []kubelet.MetadataLabel `mapstructure:"extra_metadata_labels"`

	// MetricGroupsToCollect provides a list of metrics groups to collect metrics from.
	// "container", "pod", "node" and "volume" are the only valid groups.
	MetricGroupsToCollect []kubelet.MetricGroup `mapstructure:"metric_groups"`

	// Configuration of the Kubernetes API client.
	K8sAPIConfig *k8sconfig.APIConfig `mapstructure:"k8s_api_config"`
}

func (cfg *Config) Validate() error {
	if err := cfg.ReceiverSettings.Validate(); err != nil {
		return err
	}
	if cfg.K8sAPIConfig != nil {
		if err := cfg.K8sAPIConfig.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// getReceiverOptions returns scraperOptions is the config is valid,
// otherwise it will return an error.
func (cfg *Config) getReceiverOptions() (*scraperOptions, error) {
	err := kubelet.ValidateMetadataLabelsConfig(cfg.ExtraMetadataLabels)
	if err != nil {
		return nil, err
	}

	mgs, err := getMapFromSlice(cfg.MetricGroupsToCollect)
	if err != nil {
		return nil, err
	}

	var k8sAPIClient kubernetes.Interface
	if cfg.K8sAPIConfig != nil {
		k8sAPIClient, err = k8sconfig.MakeClient(*cfg.K8sAPIConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create K8s API client: %w", err)
		}
	}

	return &scraperOptions{
		id:                    cfg.ID(),
		collectionInterval:    cfg.CollectionInterval,
		extraMetadataLabels:   cfg.ExtraMetadataLabels,
		metricGroupsToCollect: mgs,
		k8sAPIClient:          k8sAPIClient,
	}, nil
}

// getMapFromSlice returns a set of kubelet.MetricGroup values from
// the provided list. Returns an err if invalid entries are encountered.
func getMapFromSlice(collect []kubelet.MetricGroup) (map[kubelet.MetricGroup]bool, error) {
	out := make(map[kubelet.MetricGroup]bool, len(collect))
	for _, c := range collect {
		if !kubelet.ValidMetricGroups[c] {
			return nil, errors.New("invalid entry in metric_groups")
		}
		out[c] = true
	}

	return out, nil
}

func (cfg *Config) Unmarshal(componentParser *config.Map) error {
	if componentParser == nil {
		// Nothing to do if there is no config given.
		return nil
	}

	if err := componentParser.UnmarshalExact(cfg); err != nil {
		return err
	}

	// custom unmarhalling is required to get []kubelet.MetricGroup, the default
	// unmarshaller does not correctly overwrite slices.
	if !componentParser.IsSet(metricGroupsConfig) {
		cfg.MetricGroupsToCollect = defaultMetricGroups
	}

	return nil
}
