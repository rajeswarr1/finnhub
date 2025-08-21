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

func Bond_tickHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["isin"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("isin=%v", val))
		}
		if val, ok := args["date"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("date=%v", val))
		}
		if val, ok := args["limit"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("limit=%v", val))
		}
		if val, ok := args["skip"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("skip=%v", val))
		}
		if val, ok := args["exchange"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("exchange=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/bond/tick%s", cfg.BaseURL, queryString)
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

func CreateBond_tickTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_bond_tick",
		mcp.WithDescription("Bond Tick Data"),
		mcp.WithString("isin", mcp.Required(), mcp.Description("ISIN.")),
		mcp.WithString("date", mcp.Required(), mcp.Description("Date: 2020-04-02.")),
		mcp.WithString("limit", mcp.Required(), mcp.Description("Limit number of ticks returned. Maximum value: <code>25000</code>")),
		mcp.WithString("skip", mcp.Required(), mcp.Description("Number of ticks to skip. Use this parameter to loop through the entire data.")),
		mcp.WithString("exchange", mcp.Required(), mcp.Description("Currently support the following values: <code>trace</code>.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    Bond_tickHandler(cfg),
	}
}
