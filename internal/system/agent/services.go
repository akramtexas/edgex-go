/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
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
 *
 *******************************************************************************/

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/system/executor"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"net/http"
)

func processResponse(response string) map[string]interface{} {
	rsp := make(map[string]interface{})
	err := json.Unmarshal([]byte(response), &rsp)
	if err != nil {
		LoggingClient.Error("error unmarshalling response from JSON: %v", err.Error())
	}
	return rsp
}

func InvokeOperation(operation string, serviceNames []string) (interface{}, error) {
	var result []interface{}

	// Loop through requested operation, along with respectively-supplied parameters.
	for _, service := range serviceNames {
		out, err := OperationViaExecutor(service, operation)
		if err != nil {
			return nil, err
		}
		result = append(result, processResponse(out))
	}
	return result, nil
}

func InvokeMetrics(services []string, ctx context.Context) (interface{}, error) {
	var result []interface{}

	// Loop through requested actions, along with (any) respectively-supplied parameters.
	for _, service := range services {
		LoggingClient.Debug("invoke metrics")

		switch Configuration.MetricsMechanism {
		case metricsOptionViaDirectService:
			out, err := MetricsViaDirect(service, ctx)
			if err != nil {
				out = executor.CreateResult(metrics, service, executorTypeDirectService, executor.Failure(err.Error()))
			}
			result = append(result, processResponse(out))
		case metricsOptionViaExecutor:
			out, err := MetricsViaExecutor(service)
			if err != nil {
				out = executor.CreateResult(metrics, service, executorTypeUnknown, executor.Failure(err.Error()))
			}
			result = append(result, processResponse(out))
		case metricsOptionViaCustom:
			return nil, fmt.Errorf("the requested custom executor (e.g. snap) has not been integrated")
		default:
			return nil, fmt.Errorf("the requested metrics mechanism is not supported")
		}
	}
	return result, nil
}

func getConfig(services []string, ctx context.Context) (interface{}, error) {
	result := struct {
		Configuration map[string]interface{} `json:"configuration"`
	}{
		Configuration: map[string]interface{}{},
	}

	// Loop through requested actions, along with (any) respectively-supplied parameters.
	for _, service := range services {

		// Check whether SMA does _not_ know of ServiceKey ("service") as being one for one of its ready-made list of clients.
		if !IsKnownServiceKey(service) {
			LoggingClient.Info(fmt.Sprintf("service %s not known to SMA as being in the ready-made list of clients", service))

			// Service unknown to SMA, so ask the Registry whether `service` is available.
			err := registryClient.IsServiceAvailable(service)
			if err != nil {
				result.Configuration[service] = fmt.Sprintf(err.Error())
				LoggingClient.Error(err.Error())
			} else {
				LoggingClient.Info(fmt.Sprintf("Registry responded with %s service available", service))

				// Since service is unknown to SMA, ask the Registry for a ServiceEndpoint associated with `service`
				e, err := registryClient.GetServiceEndpoint(service)
				if err != nil {
					result.Configuration[service] = fmt.Sprintf("on attempting to get ServiceEndpoint for service %s, got error: %v", service, err.Error())
					LoggingClient.Error(err.Error())
				} else {
					// Preparing to add the specified key to the map where the value will be the respective GeneralClient
					clientInfo := config.ClientInfo{}
					clientInfo.Protocol = Configuration.Service.Protocol
					clientInfo.Host = e.Host
					clientInfo.Port = e.Port

					// This code will evolve to take into account a manifest-like functionality in future. So
					// rather than assume that the runtime bool flag useRegistry has been initialized to true,
					// given that the flow has reached this point, having already called functions on the Registry,
					// such as registryClient.IsServiceAvailable(service), we test for its truthiness. I expect
					// this code to be refactored as we evolve toward a manifest-like functionality in future.
					usingRegistry := false
					if registryClient != nil {
						usingRegistry = true
					}

					Configuration.Clients[e.ServiceId] = clientInfo
					params := types.EndpointParams{
						ServiceKey:  e.ServiceId,
						Path:        "/",
						UseRegistry: usingRegistry,
						Url:         Configuration.Clients[e.ServiceId].Url() + clients.ApiConfigRoute,
						Interval:    internal.ClientMonitorDefault,
					}
					// Add the service key to the map where the value is the respective GeneralClient
					generalClients[e.ServiceId] = general.NewGeneralClient(params, startup.Endpoint{RegistryClient: &registryClient})

					responseJSON, err := generalClients[e.ServiceId].FetchConfiguration(ctx)
					if err != nil {
						result.Configuration[service] = fmt.Sprintf(err.Error())
						LoggingClient.Error(err.Error())
					} else {
						result.Configuration[service] = processResponse(responseJSON)
					}
					return result, nil
				}
			}
		} else {
			// Service is known to SMA, so no need to ask the Registry for a ServiceEndpoint associated with `service`
			// Simply use one of the ready-made list of clients.
			LoggingClient.Info(fmt.Sprintf("service %s is known to SMA as being in the ready-made list of clients", service))

			responseJSON, err := generalClients[service].FetchConfiguration(ctx)
			if err != nil {
				result.Configuration[service] = fmt.Sprintf(err.Error())
				LoggingClient.Error(err.Error())
			} else {
				result.Configuration[service] = processResponse(responseJSON)
			}
		}
	}
	return result, nil
}

func getHealth(services []string) (map[string]interface{}, error) {
	health := make(map[string]interface{})

	for _, service := range services {

		if !IsKnownServiceKey(service) {
			LoggingClient.Warn(fmt.Sprintf("unknown service %s found while getting health", service))
		}

		err := registryClient.IsServiceAvailable(service)
		// the registry service returns nil for a healthy service
		if err != nil {
			health[service] = err.Error()
		} else {
			health[service] = true
		}
	}

	return health, nil
}

// Test if the service is working
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}

func IsKnownServiceKey(serviceKey string) bool {
	knownServices := map[string]struct{}{
		clients.SupportNotificationsServiceKey: {},
		clients.CoreCommandServiceKey:          {},
		clients.CoreDataServiceKey:             {},
		clients.CoreMetaDataServiceKey:         {},
		clients.ExportClientServiceKey:         {},
		clients.ExportDistroServiceKey:         {},
		clients.SupportLoggingServiceKey:       {},
		clients.SupportSchedulerServiceKey:     {},
		clients.ConfigSeedServiceKey:           {},
	}

	_, exists := knownServices[serviceKey]
	return exists
}
