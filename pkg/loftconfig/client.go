package loftconfig

import (
	"bytes"
	"encoding/json"
	"fmt"

	"dev.khulnasoft.com/pkg/credentials"
	"dev.khulnasoft.com/pkg/platform/client"
	"dev.khulnasoft.com/log"
)

func GetLoftConfig(context, provider string, port int, logger log.Logger) (*client.Config, error) {
	request := &LoftConfigRequest{
		Context:  context,
		Provider: provider,
	}

	rawJson, err := json.Marshal(request)
	if err != nil {
		logger.Errorf("Error parsing request: %w", err)
		return nil, err
	}

	configResponse := &LoftConfigResponse{}
	out, err := credentials.PostWithRetry(port, "loft-platform-credentials", bytes.NewReader(rawJson), logger)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out, configResponse)
	if err != nil {
		return nil, fmt.Errorf("decode loft config %s: %w", string(out), err)
	}

	return configResponse.LoftConfig, nil
}
