// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

//go:build darwin || freebsd || linux || windows

package process_summary

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/beats/v7/metricbeat/mb"
	mbtest "github.com/elastic/beats/v7/metricbeat/mb/testing"
	_ "github.com/elastic/beats/v7/metricbeat/module/system"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/elastic/elastic-agent-system-metrics/metric/system/process"
	"github.com/elastic/elastic-agent-system-metrics/metric/system/resolve"
)

func TestData(t *testing.T) {
	f := mbtest.NewReportingMetricSetV2Error(t, getConfig())
	err := mbtest.WriteEventsReporterV2Error(f, t, ".")
	if err != nil {
		t.Fatal("write", err)
	}
}

func TestFetch(t *testing.T) {
	f := mbtest.NewReportingMetricSetV2Error(t, getConfig())
	events, errs := mbtest.ReportingFetchV2Error(f)

	for _, err := range errs {
		assert.ErrorIsf(t, err, process.NonFatalErr{}, "Expected non-fatal error, got %v", err)
	}
	require.NotEmpty(t, events)
	event := events[0].BeatEvent("system", "process_summary").Fields
	t.Logf("%s/%s event: %+v", f.Module().Name(), f.Name(),
		event.StringToPrint())

	_, err := event.GetValue("system.process.summary")
	require.NoError(t, err)

}
func TestFetchDegradeOnPartial(t *testing.T) {
	config := getConfig()
	config["degrade_on_partial"] = true

	f := mbtest.NewReportingMetricSetV2Error(t, config)
	events, errs := mbtest.ReportingFetchV2Error(f)
	if len(errs) > 0 {
		for _, err := range errs {
			assert.NotErrorIsf(t, err, &mb.PartialMetricsError{}, "Expected non-fatal error, got %v", err)
		}
	} else {
		require.NotEmpty(t, events)
		event := events[0].BeatEvent("system", "process_summary").Fields
		t.Logf("%s/%s event: %+v", f.Module().Name(), f.Name(),
			event.StringToPrint())

		_, err := event.GetValue("system.process.summary")
		require.NoError(t, err)
	}
}

func TestStateNames(t *testing.T) {
	f := mbtest.NewReportingMetricSetV2Error(t, getConfig())
	events, errs := mbtest.ReportingFetchV2Error(f)

	for _, err := range errs {
		assert.ErrorIsf(t, err, process.NonFatalErr{}, "Expected non-fatal error, got %v", err)
	}
	require.NotEmpty(t, events)
	event := events[0].BeatEvent("system", "process_summary").Fields

	summary, err := event.GetValue("system.process.summary")
	require.NoError(t, err)

	event, ok := summary.(mapstr.M)
	require.True(t, ok)

	// if there's nothing marked as sleeping or idle, something weird is happening
	assert.NotZero(t, event["total"])

	var sum int
	total, ok := event["total"].(int)
	require.Truef(t, ok, "Expected int got %T", event["total"])
	for key, val := range event {
		if key == "total" {
			continue
		}
		if _, ok := val.(int); !ok {
			continue
		}
		// Check to make sure the values we got actually exist
		exists := false
		for _, proc := range process.PidStates {
			if string(proc) == key {
				exists = true
				break
			}
		}
		assert.True(t, exists, "could not find value %s in event #%v", key, event.StringToPrint())

		sum = val.(int) + sum
	}
	assert.Equal(t, total, sum)

}

func TestThreads(t *testing.T) {
	root := resolve.NewTestResolver("_meta/testdata")
	stats, err := threadStats(root)
	require.NoError(t, err)
	require.Equal(t, int64(1), stats["blocked"])
	require.Equal(t, int64(3), stats["running"])
	t.Logf("metrics: %#v", stats)
}

func getConfig() map[string]interface{} {
	return map[string]interface{}{
		"module":     "system",
		"metricsets": []string{"process_summary"},
	}
}
