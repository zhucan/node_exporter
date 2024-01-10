// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs/sysfs"
)

var (
	desayGPU = "/sys/devices/gpu.0"
)

const (
	desayGPUCollectorSubsystem = "desay_gpu"
)

type gpuDesayCollector struct {
	fs           sysfs.FS
	desayGPUInfo *prometheus.Desc
	logger       log.Logger
}

func init() {
	if runtime.GOARCH == "arm64" {
		if _, err := os.Stat(desayGPU); err == nil {
			registerCollector("desay_gpu", defaultEnabled, NewDesayGPUCollector)
		}
	}
}

func NewDesayGPUCollector(logger log.Logger) (Collector, error) {
	fs, err := sysfs.NewFS(*sysPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sysfs: %w", err)
	}
	c := &gpuDesayCollector{
		fs: fs,
		desayGPUInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, desayGPUCollectorSubsystem, "load"),
			"GPU load",
			[]string{"name"}, nil,
		),
		logger: logger,
	}
	return c, nil
}

func getDesayGPULoad() (string, error) {
	load, err := os.ReadFile(filepath.Join(desayGPU, "load"))
	if err != nil {
		return "", err
	}
	return string(load), nil
}

func (c *gpuDesayCollector) Update(ch chan<- prometheus.Metric) error {
	load, err := getDesayGPULoad()
	if err != nil {
		return err
	}
	fload, err := strconv.ParseFloat(strings.Replace(load, "\n", "", -1), 64)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(c.desayGPUInfo,
		prometheus.GaugeValue,
		fload,
		desayGPUCollectorSubsystem,
	)
	return nil
}
