package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/finnhub-api/mcp-server/config"
	"github.com/finnhub-api/mcp-server/models"
	"github.com/mark3labs/mcp-go/mcp"
)

func Etfs_holdingsHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["symbol"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("symbol=%v", val))
		}
		if val, ok := args["isin"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("isin=%v", val))
		}
		if val, ok := args["skip"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("skip=%v", val))
		}
		if val, ok := args["date"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("date=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/etf/holdings%s", cfg.BaseURL, queryString)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to create request", err), nil
		}
		// No authentication required for this endpoint
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Request failed", err), nil
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to read response body", err), nil
		}

		if resp.StatusCode >= 400 {
			return mcp.NewToolResultError(fmt.Sprintf("API error: %s", body)), nil
		}
		// Use properly typed response
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			// Fallback to raw text if unmarshaling fails
			return mcp.NewToolResultText(string(body)), nil
		}

		prettyJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to format JSON", err), nil
		}

		return mcp.NewToolResultText(string(prettyJSON)), nil
	}
}

func CreateEtfs_holdingsTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_etf_holdings",
		mcp.WithDescription("ETFs Holdings"),
		mcp.WithString("symbol", mcp.Description("ETF symbol.")),
		mcp.WithString("isin", mcp.Description("ETF isin.")),
		mcp.WithString("skip", mcp.Description("Skip the first n results. You can use this parameter to query historical constituents data. The latest result is returned if skip=0 or not set.")),
		mcp.WithString("date", mcp.Description("Query holdings by date. You can use either this param or <code>skip</code> param, not both.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    Etfs_holdingsHandler(cfg),
	}
}
