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

const (
	metricsResult               = "result"
	metricsResultCpuUsedPercent = "cpuUsedPercent"
	metricsResultMemoryUsed     = "memoryUsed"
	metricsResultRaw            = "raw"
)

func Failure(errorMessage string) string {
	return "\"success\":false,\"errorMessage\":\"" + errorMessage + "\""
}

func success() string {
	return "\"success\":true"
}

func MetricsSuccess(cpuUsedPercent string, memoryUsed string, raw string) string {
	return success() + "," +
		"\"" + metricsResult + "\":{" +
		"\"" + metricsResultCpuUsedPercent + "\":" + cpuUsedPercent + "," +
		"\"" + metricsResultMemoryUsed + "\":" + memoryUsed + "," +
		"\"" + metricsResultRaw + "\":" + raw +
		"}"
}

func CreateResult(operation string, service string, executorType string, result string) string {
	return "{\"operation\":\"" + operation + "\",\"service\":\"" + service + "\",\"executor\":\"" + executorType + "\"," + result + "}"
}
