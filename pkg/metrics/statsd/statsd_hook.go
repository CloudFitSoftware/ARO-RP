package statsd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Azure/ARO-RP/pkg/metrics"

	"github.com/sirupsen/logrus"
)

type StatsdHook struct {
	mEmitter metrics.Emitter
}

func NewStatsdHook(m metrics.Emitter) *StatsdHook {
	return &StatsdHook{m}
}

func (hook *StatsdHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read log entry, %v", err)
		return err
	}

	if len(entry.Data) == 0 {
		return nil
	}

	fmt.Fprintf(os.Stdout, "\nLog Entry Captured by StatsdHook: %s", line)

	isE2EEmittableMetricStr := fmt.Sprint(entry.Data["IsE2EEmittableMetric"])
	isE2EEmittableMetric, isEmittableErr := strconv.ParseBool(isE2EEmittableMetricStr)

	if isEmittableErr != nil {
		fmt.Fprintf(os.Stderr, "Unable to read IsE2EEmittableMetric value, %v", err)
		return isEmittableErr
	}

	if isE2EEmittableMetric {
		metricName := fmt.Sprint(entry.Data["MetricName"])
		metricStatusStr := fmt.Sprint(entry.Data["MetricStatus"])
		metricStatusBool, metricStatusErr := strconv.ParseBool(metricStatusStr)

		if metricStatusErr != nil {
			fmt.Fprintf(os.Stderr, "Unable to parse MetricStatus value, %v", err)
			return metricStatusErr
		}

		metricStatusInt := int64(Btoi(metricStatusBool))
		dimensions := map[string]string{"armResourceID": fmt.Sprint(entry.Data["armResourceID"]), "armGeoLocation": fmt.Sprint(entry.Data["armGeoLocation"]), "resourceType": fmt.Sprint(entry.Data["resourceType"])}

		fmt.Fprintf(os.Stdout, "\nAttempting to Emit Metric: %s", metricName)
		hook.mEmitter.EmitGauge(metricName, metricStatusInt, dimensions)
	}

	return nil
}

func (hook *StatsdHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func Btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
