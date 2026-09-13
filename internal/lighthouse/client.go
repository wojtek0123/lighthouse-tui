package lighthouse

import (
	"context"
	"encoding/json"
	"os/exec"
)

// check if lightyhouse version is supported
type LighthouseResult struct {
	FetchTime         string `json:"fetchTime"`
	LighthouseVersion string `json:"lighthouseVersion"`
	Audits            map[string]struct {
		NumericValue float64 `json:"numericValue"`
		DisplayValue string  `json:"displayValue"`
		Title        string  `json:"title"`
		Description  string  `json:"description"`
		Unit         string  `json:"numericUnit"`
		Score        string  `json:"score"`
	} `json:"audits"`
	Categories map[string]struct {
		Title string  `json:"title"`
		Id    string  `json:"id"`
		Score float64 `json:"score"`
	}
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
