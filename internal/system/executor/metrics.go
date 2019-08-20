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

import "strings"

func metricsExecutorCommands(serviceName string) []string {
	return []string{
		"stats",
		serviceName,
		"--no-stream",
		"--format",
		"{\"cpu_perc\":\"{{ .CPUPerc }}\",\"mem_usage\":\"{{ .MemUsage }}\",\"mem_perc\":\"{{ .MemPerc }}\",\"net_io\":\"{{ .NetIO }}\",\"block_io\":\"{{ .BlockIO }}\",\"pids\":\"{{ .PIDs }}\"}",
	}
}

func metricsSuccess(result string) string {
	return success() + ",\"result\":" + result
}

func gatherMetrics(serviceName string, executor CommandExecutor) string {
	output, err := executor(metricsExecutorCommands(serviceName)...)
	if err != nil {
		return failure(err.Error())
	}
	return metricsSuccess(strings.TrimRight(string(output), "\n"))
}
