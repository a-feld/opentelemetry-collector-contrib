// Copyright 2019, OpenTelemetry Authors
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

package carbonreceiver

import (
	"context"

	"go.opencensus.io/trace"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/obsreport"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver/transport"
)

// reporter struct implements the transport.Reporter interface to give consistent
// observability per Collector metric observability package.
type reporter struct {
	id            config.ComponentID
	spanName      string
	logger        *zap.Logger
	sugaredLogger *zap.SugaredLogger // Used for generic debug logging
	obsrecv       *obsreport.Receiver
}

var _ transport.Reporter = (*reporter)(nil)

func newReporter(id config.ComponentID, set component.ReceiverCreateSettings) transport.Reporter {
	return &reporter{
		id:            id,
		spanName:      id.String() + ".receiver",
		logger:        set.Logger,
		sugaredLogger: set.Logger.Sugar(),
		obsrecv: obsreport.NewReceiver(obsreport.ReceiverSettings{
			ReceiverID:             id,
			Transport:              "tcp",
			ReceiverCreateSettings: set,
		}),
	}
}

// OnDataReceived is called when a message or request is received from
// a client. The returned context should be used in other calls to the same
// reporter instance. The caller code should include a call to end the
// returned span.
func (r *reporter) OnDataReceived(ctx context.Context) context.Context {
	return r.obsrecv.StartMetricsOp(ctx)
}

// OnTranslationError is used to report a translation error from original
// format to the internal format of the Collector. The context and span
// passed to it should be the ones returned by OnDataReceived.
func (r *reporter) OnTranslationError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	r.logger.Debug("Carbon translation error", zap.Error(err))

	// Using annotations since multiple translation errors can happen in the
	// same client message/request. The time itself is not relevant.
	span := trace.FromContext(ctx)
	span.Annotate([]trace.Attribute{
		trace.StringAttribute("error", err.Error())},
		"translation",
	)
}

// OnMetricsProcessed is called when the received data is passed to next
// consumer on the pipeline. The context and span passed to it should be the
// ones returned by OnDataReceived. The error should be error returned by
// the next consumer - the reporter is expected to handle nil error too.
func (r *reporter) OnMetricsProcessed(
	ctx context.Context,
	numReceivedMetricPoints int,
	err error,
) {
	if err != nil {
		r.logger.Debug(
			"Carbon receiver failed to push metrics into pipeline",
			zap.Int("numReceivedMetricPoints", numReceivedMetricPoints),
			zap.Error(err))

		span := trace.FromContext(ctx)
		span.SetStatus(trace.Status{
			Code:    trace.StatusCodeUnknown,
			Message: err.Error(),
		})
	}

	r.obsrecv.EndMetricsOp(ctx, "carbon", numReceivedMetricPoints, err)
}

func (r *reporter) OnDebugf(template string, args ...interface{}) {
	if r.logger.Check(zap.DebugLevel, "debug") != nil {
		r.sugaredLogger.Debugf(template, args...)
	}
}
