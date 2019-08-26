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

package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
	"github.com/edgexfoundry/edgex-go/internal/system/executor"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
)

func normalizeAndWrap(serviceName string, result string) (string, error) {
	var s telemetry.SystemUsage
	if err := json.NewDecoder(bytes.NewBuffer([]byte(result))).Decode(&s); err != nil {
		return "", fmt.Errorf("error decoding telemetry.SystemUsage: %s", err.Error())
	}

	return executor.CreateResult(
		metrics,
		serviceName,
		executorTypeDirectService,
		executor.MetricsSuccess(fmt.Sprintf("%.2f", s.CpuBusyAvg), fmt.Sprintf("%d", s.Memory.Sys), result)), nil
}

func fetchMetrics(serviceName string, ctx context.Context) (string, error) {
	result, err := generalClients[serviceName].FetchMetrics(ctx)
	if err != nil {
		return "", err
	}
	return normalizeAndWrap(serviceName, result)
}

func handleUnknownService(serviceName string, ctx context.Context) (string, error) {
	LoggingClient.Info(fmt.Sprintf("service %s not known to SMA as being in the ready-made list of clients", serviceName))

	if registryClient == nil {
		return "", fmt.Errorf("registryClient not initialized; required to handle unknown service %s", serviceName)
	}

	// Service unknown to SMA, so ask the Registry whether `serviceName` is available.
	if err := registryClient.IsServiceAvailable(serviceName); err != nil {
		return "", err
	}

	LoggingClient.Info(fmt.Sprintf("Registry responded with %s serviceName available", serviceName))

	// Since serviceName is unknown to SMA, ask the Registry for a ServiceEndpoint associated with `serviceName`
	endpoint, err := registryClient.GetServiceEndpoint(serviceName)
	if err != nil {
		return "", fmt.Errorf("on attempting to get ServiceEndpoint for serviceName %s, got error: %v", serviceName, err.Error())
	}

	// add the specified key to the map where the value will be the respective GeneralClient
	Configuration.Clients[endpoint.ServiceId] = config.ClientInfo{
		Protocol: Configuration.Service.Protocol,
		Host:     endpoint.Host,
		Port:     endpoint.Port,
	}

	params := types.EndpointParams{
		ServiceKey:  endpoint.ServiceId,
		Path:        "/",
		UseRegistry: true,
		Url:         Configuration.Clients[endpoint.ServiceId].Url() + clients.ApiMetricsRoute,
		Interval:    internal.ClientMonitorDefault,
	}

	// Add the serviceName key to the map where the value is the respective GeneralClient
	generalClients[endpoint.ServiceId] = general.NewGeneralClient(params, startup.Endpoint{RegistryClient: &registryClient})

	return fetchMetrics(endpoint.ServiceId, ctx)
}

func handleKnownService(serviceName string, ctx context.Context) (string, error) {
	// Service is known to SMA, so no need to ask the Registry for a ServiceEndpoint associated with `serviceName`
	// Simply use one of the ready-made list of clients.
	LoggingClient.Info(fmt.Sprintf("serviceName %s is known to SMA as being in the ready-made list of clients", serviceName))
	return fetchMetrics(serviceName, ctx)
}

func MetricsViaDirect(serviceName string, ctx context.Context) (string, error) {
	if _, ok := generalClients[serviceName]; ok {
		return handleKnownService(serviceName, ctx)
	}
	return handleUnknownService(serviceName, ctx)
}
