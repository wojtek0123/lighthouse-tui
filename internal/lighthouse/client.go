package lighthouse

import (
	"context"
	"encoding/json"
	"os/exec"
)

type LighthouseResult struct {
	Audits map[string]struct {
		NumericValue float64 `json:"numericValue"`
		DisplayValue string  `json:"displayValue"`
	} `json:"audits"`
}

func RunLighthouse(ctx context.Context, target string) (LighthouseResult, error) {
	cmd := exec.CommandContext(
		ctx,
		"npx",
		"lighthouse",
		target,
		"--quiet",
		"--output=json",
		"--chrome-flags=--headless=new --no-sandbox --disable-gpu --disable-dev-shm-usage",
	)

	// "--chrome-flags=--headless",
	output, err := cmd.Output()
	if err != nil {
		return LighthouseResult{}, err
	}

	var result LighthouseResult
	err = json.Unmarshal(output, &result)
	return result, err
}
