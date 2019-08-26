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
)

const (
	Start   = "start"
	Stop    = "stop"
	Restart = "restart"
	Metrics = "metrics"

	executorType        = "docker"
	failedStartPrefix   = "Error starting service"
	failedRestartPrefix = "Error restarting service"
	failedStopPrefix    = "Error stopping service"
)

type CommandExecutor func(arg ...string) ([]byte, error)

func messageExecutorOperationNotSupported() string {
	return "operation not supported by executor"
}

func messageMissingArguments() string {
	return fmt.Sprintf("missing <service> and <operation> command line arguments")
}

func Execute(args []string, executor CommandExecutor) string {
	if len(args) > 2 {
		service := args[1]
		operation := args[2]

		var result string
		switch operation {
		case Start:
			result = executeACommand(operation, service, executor, failedStartPrefix, true)
		case Restart:
			result = executeACommand(operation, service, executor, failedRestartPrefix, true)
		case Stop:
			result = executeACommand(operation, service, executor, failedStopPrefix, false)
		case Metrics:
			result = gatherMetrics(service, executor)
		default:
			result = Failure(messageExecutorOperationNotSupported())
		}
		return CreateResult(operation, service, executorType, result)
	}
	return CreateResult("", "", executorType, Failure(messageMissingArguments()))
}
