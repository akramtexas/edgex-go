/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package executor

import (
	"fmt"
	"strings"
)

const separator = ";"

func metricsExecutorCommands(serviceName string) []string {
	return []string{
		"stats",
		serviceName,
		"--no-stream",
		"--format",
		"{{ .CPUPerc }}" + separator +
			"{{ .MemUsage }}" + separator +
			"{\"cpu_perc\":\"{{ .CPUPerc }}\",\"mem_usage\":\"{{ .MemUsage }}\",\"mem_perc\":\"{{ .MemPerc }}\",\"net_io\":\"{{ .NetIO }}\",\"block_io\":\"{{ .BlockIO }}\",\"pids\":\"{{ .PIDs }}\"}",
	}
}

func dockerMemoryToStringInt(value string) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	var memory float64
	var scale string
	n, err := fmt.Sscanf(value, "%f%s", &memory, &scale)
	if err != nil || n != 2 {
		return "-1"
	}
	switch scale {
	case "KiB":
		memory *= kb
	case "MiB":
		memory *= mb
	case "GiB":
		memory *= gb
	}
	return fmt.Sprintf("%.0f", memory)
}

func resultToFields(result string) (cpuUsedPercent, memoryUsed, raw string) {
	resultFields := strings.Split(strings.TrimRight(result, "\n"), separator)
	cpuUsedPercent = strings.TrimRight(resultFields[0], "%")
	memoryUsed = dockerMemoryToStringInt(strings.Split(resultFields[1], " ")[0])
	raw = resultFields[2]
	return
}

func gatherMetrics(serviceName string, executor CommandExecutor) string {
	result, err := executor(metricsExecutorCommands(serviceName)...)
	if err != nil {
		return Failure(err.Error())
	}
	return MetricsSuccess(resultToFields(string(result)))
}
