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
	"github.com/edgexfoundry/edgex-go/internal/system/agent"
)

const (
	start   = "start"
	stop    = "stop"
	restart = "restart"
	metrics = "metrics"

	failedStartPrefix   = "Error starting service"
	failedRestartPrefix = "Error restarting service"
	failedStopPrefix    = "Error stopping service"
)

type CommandExecutor func(arg ...string) ([]byte, error)

func messageExecutorOperationNotSupported() string {
	return "operation not supported by executor"
}

func messageSpecifiedServiceIsUnknown() string {
	return "Specified service is unknown"
}

func messageMissingArguments(executableName string) string {
	return fmt.Sprintf("Usage: ./%s <service> <operation>\t\tStart app with requested {service} and {operation}\n", executableName)
}

func Execute(args []string, executor CommandExecutor) string {
	if len(args) > 2 {
		service := args[1]
		if agent.IsKnownServiceKey(service) {
			operation := args[2]

			switch operation {
			case start:
				return createResult(operation, service, executeACommand(operation, service, executor, failedStartPrefix, true))
			case restart:
				return createResult(operation, service, executeACommand(operation, service, executor, failedRestartPrefix, true))
			case stop:
				return createResult(operation, service, executeACommand(operation, service, executor, failedStopPrefix, false))
			case metrics:
				return createResult(operation, service, gatherMetrics(service, executor))
			default:
				return createResult(operation, service, failure(messageExecutorOperationNotSupported()))
			}
		}
		return createResult("", service, failure(messageSpecifiedServiceIsUnknown()))
	}
	return createResult("", "", failure(messageMissingArguments(args[0])))
}
