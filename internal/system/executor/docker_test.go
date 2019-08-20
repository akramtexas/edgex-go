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
	"errors"
	"github.com/edgexfoundry/edgex-go/internal/system/agent"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	executableName       = "executableName"
	errorMessage         = "errorMessage"
	invalidOperation     = "invalidOperation"
	metricsSuccessResult = "metricsSuccessResult"
)

type executorStubCall struct {
	expectedArgs []string // expected arg value for specific executor call
	outBytes     []byte   // return value for specific executor call
	outError     error    // return value for specific executor call
}

type executorStub struct {
	Called         int                // number of times stub is called
	capturedArgs   [][]string         // captures arg values for each stub call
	perCallResults []executorStubCall // expected arg value and return values for each stub call
}

func newExecutor(results []executorStubCall) executorStub {
	return executorStub{
		perCallResults: results,
	}
}

func (e *executorStub) commandExecutor(arg ...string) ([]byte, error) {
	e.Called++
	e.capturedArgs = append(e.capturedArgs, arg)
	return e.perCallResults[e.Called-1].outBytes, e.perCallResults[e.Called-1].outError
}

func assertArgsAreEqual(t *testing.T, expected []string, actual []string) {
	assert.Equal(t, len(expected), len(actual))
	for key, expectedValue := range expected {
		assert.Equal(t, expectedValue, actual[key])
	}
}

func firstCommandCallFails(serviceName string, operation string) []executorStubCall {
	return []executorStubCall{
		{[]string{serviceName, operation}, []byte(nil), errors.New(errorMessage)},
	}
}

func secondCommandCallFails(serviceName string, operation string) []executorStubCall {
	return []executorStubCall{
		{[]string{serviceName, operation}, []byte(nil), nil},
		{[]string{serviceName, inspect}, []byte(nil), errors.New(errorMessage)},
	}
}

func secondCommandCallSucceeds(serviceName string, operation string, result string) []executorStubCall {
	return []executorStubCall{
		{[]string{serviceName, operation}, []byte(nil), nil},
		{[]string{serviceName, inspect}, []byte(result), nil},
	}
}

func firstMetricsCallFails(serviceName string, operation string) []executorStubCall {
	return []executorStubCall{
		{metricsExecutorCommands(serviceName), []byte(nil), errors.New(errorMessage)},
	}
}

func firstMetricsCallSucceeds(serviceName string, operation string, result string) []executorStubCall {
	return []executorStubCall{
		{metricsExecutorCommands(serviceName), []byte(result), nil},
	}
}

func executeArguments(serviceName string, operation string) []string {
	return []string{executableName, serviceName, operation}
}

func TestExecute(t *testing.T) {
	for serviceName := range agent.KnownServices() {
		tests := []struct {
			name           string
			operation      string
			expectedResult string
			executorCalls  []executorStubCall
		}{
			// start command test cases

			{
				"Start: first executor call fails",
				start,
				createResult(start, serviceName, failure(messageExecutorCommandFailed(failedStartPrefix, string([]byte(nil)), errorMessage))),
				firstCommandCallFails(serviceName, start),
			},
			{
				"Start: second executor call fails",
				start,
				createResult(start, serviceName, failure(messageExecutorInspectFailed(failedStartPrefix, errorMessage))),
				secondCommandCallFails(serviceName, start),
			},
			{
				"Start: container not found in inspect result",
				start,
				createResult(start, serviceName, failure(messageExecutorInspectFailed(failedStartPrefix, messageContainerNotFound(serviceName)))),
				secondCommandCallSucceeds(serviceName, start, "[]"),
			},
			{
				"Start: more than one container instance found in inspect result",
				start,
				createResult(start, serviceName, failure(messageExecutorInspectFailed(failedStartPrefix, messageMoreThanOneContainerFound(serviceName)))),
				secondCommandCallSucceeds(serviceName, start, "[{\"State\": {\"Running\": false}}, {\"State\": {\"Running\": false}}]"),
			},
			{
				"Start: inspect result says service is not running as expected",
				start,
				createResult(start, serviceName, failure(messageServiceIsNotRunningButShouldBe(failedStartPrefix))),
				secondCommandCallSucceeds(serviceName, start, "[{\"State\": {\"Running\": false}}]"),
			},
			{
				"Start: isContainerRunning json.Decode failure",
				start,
				createResult(start, serviceName, failure(messageExecutorInspectFailed(failedStartPrefix, "EOF"))),
				secondCommandCallSucceeds(serviceName, start, ""),
			},
			{
				"Start: success",
				start,
				createResult(start, serviceName, success()),
				secondCommandCallSucceeds(serviceName, start, "[{\"State\": {\"Running\": true}}]"),
			},

			// restart command test cases

			{
				"Restart: first executor call fails",
				restart,
				createResult(restart, serviceName, failure(messageExecutorCommandFailed(failedRestartPrefix, string([]byte(nil)), errorMessage))),
				firstCommandCallFails(serviceName, restart),
			},
			{
				"Restart: second executor call fails",
				restart,
				createResult(restart, serviceName, failure(messageExecutorInspectFailed(failedRestartPrefix, errorMessage))),
				secondCommandCallFails(serviceName, restart),
			},
			{
				"Restart: container not found in inspect result",
				restart,
				createResult(restart, serviceName, failure(messageExecutorInspectFailed(failedRestartPrefix, messageContainerNotFound(serviceName)))),
				secondCommandCallSucceeds(serviceName, restart, "[]"),
			},
			{
				"Restart: more than one container instance found in inspect result",
				restart,
				createResult(restart, serviceName, failure(messageExecutorInspectFailed(failedRestartPrefix, messageMoreThanOneContainerFound(serviceName)))),
				secondCommandCallSucceeds(serviceName, restart, "[{\"State\": {\"Running\": false}}, {\"State\": {\"Running\": false}}]"),
			},
			{
				"Restart: inspect result says service is not running as expected",
				restart,
				createResult(restart, serviceName, failure(messageServiceIsNotRunningButShouldBe(failedRestartPrefix))),
				secondCommandCallSucceeds(serviceName, restart, "[{\"State\": {\"Running\": false}}]"),
			},
			{
				"Restart: isContainerRunning json.Decode failure",
				restart,
				createResult(restart, serviceName, failure(messageExecutorInspectFailed(failedRestartPrefix, "EOF"))),
				secondCommandCallSucceeds(serviceName, restart, ""),
			},
			{
				"Restart: success",
				restart,
				createResult(restart, serviceName, success()),
				secondCommandCallSucceeds(serviceName, restart, "[{\"State\": {\"Running\": true}}]"),
			},

			// stop command test cases

			{
				"Stop: first executor call fails",
				stop,
				createResult(stop, serviceName, failure(messageExecutorCommandFailed(failedStopPrefix, string([]byte(nil)), errorMessage))),
				firstCommandCallFails(serviceName, stop),
			},
			{
				"Stop: second executor call fails",
				stop,
				createResult(stop, serviceName, failure(messageExecutorInspectFailed(failedStopPrefix, errorMessage))),
				secondCommandCallFails(serviceName, stop),
			},
			{
				"Stop: container not found in inspect result",
				stop,
				createResult(stop, serviceName, failure(messageExecutorInspectFailed(failedStopPrefix, messageContainerNotFound(serviceName)))),
				secondCommandCallSucceeds(serviceName, stop, "[]"),
			},
			{
				"Stop: more than one container instance found in inspect result",
				stop,
				createResult(stop, serviceName, failure(messageExecutorInspectFailed(failedStopPrefix, messageMoreThanOneContainerFound(serviceName)))),
				secondCommandCallSucceeds(serviceName, stop, "[{\"State\": {\"Running\": true}}, {\"State\": {\"Running\": true}}]"),
			},
			{
				"Stop: inspect result says service is not running as expected",
				stop,
				createResult(stop, serviceName, failure(messageServiceIsRunningButShouldNotBe(failedStopPrefix))),
				secondCommandCallSucceeds(serviceName, stop, "[{\"State\": {\"Running\": true}}]"),
			},
			{
				"Stop: isContainerRunning json.Decode failure",
				stop,
				createResult(stop, serviceName, failure(messageExecutorInspectFailed(failedStopPrefix, "EOF"))),
				secondCommandCallSucceeds(serviceName, stop, ""),
			},
			{
				"Stop: success",
				stop,
				createResult(stop, serviceName, success()),
				secondCommandCallSucceeds(serviceName, stop, "[{\"State\": {\"Running\": false}}]"),
			},

			// metrics command test case

			{
				"Metrics: failure",
				metrics,
				createResult(metrics, serviceName, failure(errorMessage)),
				firstMetricsCallFails(serviceName, metrics),
			},
			{
				"Metrics: success",
				metrics,
				createResult(metrics, serviceName, metricsSuccess(metricsSuccessResult)),
				firstMetricsCallSucceeds(serviceName, metrics, metricsSuccessResult),
			},

			// invalid operation test case

			{
				"operation not supported by executor",
				invalidOperation,
				createResult(invalidOperation, serviceName, failure(messageExecutorOperationNotSupported())),
				[]executorStubCall{},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				executor := newExecutor(test.executorCalls)

				result := Execute(executeArguments(serviceName, test.operation), executor.commandExecutor)

				if assert.Equal(t, len(test.executorCalls), executor.Called) {
					for key, executorCall := range test.executorCalls {
						assertArgsAreEqual(t, executorCall.expectedArgs, executor.capturedArgs[key])
					}
				}
				assert.Equal(t, test.expectedResult, result)
			})
		}
	}
}

func TestUnknownService(t *testing.T) {
	const unknownServiceName = "unknownServiceName"
	executor := newExecutor([]executorStubCall{})

	result := Execute(executeArguments(unknownServiceName, ""), executor.commandExecutor)

	assert.Equal(t, 0, executor.Called)
	assert.Equal(t, createResult("", unknownServiceName, failure(messageSpecifiedServiceIsUnknown())), result)
}

func TestMissingArguments(t *testing.T) {
	missingArguments := []string{executableName}
	executor := newExecutor([]executorStubCall{})

	result := Execute(missingArguments, executor.commandExecutor)

	assert.Equal(t, 0, executor.Called)
	assert.Equal(t, createResult("", "", failure(messageMissingArguments(executableName))), result)
}
