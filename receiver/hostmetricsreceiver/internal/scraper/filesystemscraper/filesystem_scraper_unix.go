// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux || darwin || freebsd || openbsd || solaris
// +build linux darwin freebsd openbsd solaris

package filesystemscraper

import (
	"go.opentelemetry.io/collector/model/pdata"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/filesystemscraper/internal/metadata"
)

const fileSystemStatesLen = 3

func appendFileSystemUsageStateDataPoints(idps pdata.NumberDataPointSlice, now pdata.Timestamp, deviceUsage *deviceUsage) {
	initializeFileSystemUsageDataPoint(idps.AppendEmpty(), now, deviceUsage.partition, metadata.AttributeState.Used, int64(deviceUsage.usage.Used))
	initializeFileSystemUsageDataPoint(idps.AppendEmpty(), now, deviceUsage.partition, metadata.AttributeState.Free, int64(deviceUsage.usage.Free))
	initializeFileSystemUsageDataPoint(idps.AppendEmpty(), now, deviceUsage.partition, metadata.AttributeState.Reserved, int64(deviceUsage.usage.Total-deviceUsage.usage.Used-deviceUsage.usage.Free))
}

const systemSpecificMetricsLen = 1

func appendSystemSpecificMetrics(metrics pdata.MetricSlice, now pdata.Timestamp, deviceUsages []*deviceUsage) {
	metric := metrics.AppendEmpty()
	metadata.Metrics.SystemFilesystemInodesUsage.Init(metric)

	idps := metric.Sum().DataPoints()
	idps.EnsureCapacity(2 * len(deviceUsages))
	for _, deviceUsage := range deviceUsages {
		initializeFileSystemUsageDataPoint(idps.AppendEmpty(), now, deviceUsage.partition, metadata.AttributeState.Used, int64(deviceUsage.usage.InodesUsed))
		initializeFileSystemUsageDataPoint(idps.AppendEmpty(), now, deviceUsage.partition, metadata.AttributeState.Free, int64(deviceUsage.usage.InodesFree))
	}
}
