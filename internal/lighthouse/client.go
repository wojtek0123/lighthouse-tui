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
		Title        string  `json:"title"`
		Description  string  `json:"description"`
		Unit         string  `json:"numericUnit"`
		Score        string  `json:"score"`
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
		"--chrome-flags=--headless=new",
	)

	output, err := cmd.Output()
	if err != nil {
		return LighthouseResult{}, err
	}

	var result LighthouseResult
	err = json.Unmarshal(output, &result)
	return result, err
}
