// Copyright 2020 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datareceivers

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/signalfxreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/testbed/testbed"
)

// SFxMetricsDataReceiver implements SignalFx format receiver.
type SFxMetricsDataReceiver struct {
	testbed.DataReceiverBase
	receiver component.MetricsReceiver
}

// Ensure SFxMetricsDataReceiver implements MetricDataSender.
var _ testbed.DataReceiver = (*SFxMetricsDataReceiver)(nil)

// NewSFxMetricsDataReceiver creates a new SFxMetricsDataReceiver that will listen on the
// specified port after Start is called.
func NewSFxMetricsDataReceiver(port int) *SFxMetricsDataReceiver {
	return &SFxMetricsDataReceiver{DataReceiverBase: testbed.DataReceiverBase{Port: port}}
}

// Start the receiver.
func (sr *SFxMetricsDataReceiver) Start(_ consumer.Traces, mc consumer.Metrics, _ consumer.Logs) error {
	config := signalfxreceiver.Config{
		HTTPServerSettings: confighttp.HTTPServerSettings{
			Endpoint: fmt.Sprintf("localhost:%d", sr.Port),
		},
	}
	var err error
	f := signalfxreceiver.NewFactory()
	sr.receiver, err = f.CreateMetricsReceiver(context.Background(), componenttest.NewNopReceiverCreateSettings(), &config, mc)
	if err != nil {
		return err
	}

	return sr.receiver.Start(context.Background(), sr)
}

// Stop the receiver.
func (sr *SFxMetricsDataReceiver) Stop() error {
	return sr.receiver.Shutdown(context.Background())
}

// GenConfigYAMLStr returns exporter config for the agent.
func (sr *SFxMetricsDataReceiver) GenConfigYAMLStr() string {
	// Note that this generates an exporter config for agent.
	return fmt.Sprintf(`
    signalfx:
      ingest_url: "http://localhost:%d"
      api_url: "http://localhost/"
      access_token: "access_token"`, sr.Port)
}

// ProtocolName returns protocol name as it is specified in Collector config.
func (sr *SFxMetricsDataReceiver) ProtocolName() string {
	return "signalfx"
}
