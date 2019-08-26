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
	"encoding/json"
	"fmt"
	"strings"
)

const inspect = "inspect"

func messageExecutorCommandFailed(operationPrefix string, result string, errorMessage string) string {
	return fmt.Sprintf("%s: %s (%s)", operationPrefix, errorMessage, strings.ReplaceAll(result, "\n", " "))
}

func messageExecutorInspectFailed(operationPrefix string, errorMessage string) string {
	return fmt.Sprintf("%s: %s", operationPrefix, errorMessage)
}

func messageServiceIsNotRunningButShouldBe(operationPrefix string) string {
	return fmt.Sprintf("%s: service is not running but should be", operationPrefix)
}

func messageServiceIsRunningButShouldNotBe(operationPrefix string) string {
	return fmt.Sprintf("%s: service is running but shouldn't be", operationPrefix)
}

func messageContainerNotFound(serviceName string) string {
	return fmt.Sprintf("container %s not found", serviceName)
}

func messageMoreThanOneContainerFound(serviceName string) string {
	return fmt.Sprintf("multiple containers found with name %s", serviceName)
}

func isContainerRunning(service string, executor CommandExecutor) (bool, string) {
	// check the status of the container using the json format - include all
	// containers as the container we want to check may be Exited
	stringOutput, err := executor(inspect, service)
	if err != nil {
		return false, err.Error()
	}

	var containerStatus []struct {
		State struct {
			Running bool
		}
	}
	jsonOutput := json.NewDecoder(strings.NewReader(string(stringOutput)))
	if err = jsonOutput.Decode(&containerStatus); err != nil {
		return false, err.Error()
	}

	switch {
	case len(containerStatus) < 1:
		return false, messageContainerNotFound(service)
	case len(containerStatus) > 1:
		return false, messageMoreThanOneContainerFound(service)
	default:
		return containerStatus[0].State.Running, ""
	}
}

func executeACommand(
	operation string,
	service string,
	executor CommandExecutor,
	operationPrefix string,
	shouldBeRunning bool) string {

	if output, err := executor(operation, service); err != nil {
		return Failure(messageExecutorCommandFailed(operationPrefix, string(output), err.Error()))
	}

	isRunning, errorMessage := isContainerRunning(service, executor)
	switch {
	case len(errorMessage) > 0:
		return Failure(messageExecutorInspectFailed(operationPrefix, errorMessage))
	case isRunning != shouldBeRunning:
		if isRunning {
			return Failure(messageServiceIsRunningButShouldNotBe(operationPrefix))
		}
		return Failure(messageServiceIsNotRunningButShouldBe(operationPrefix))
	default:
		return success()
	}
}
