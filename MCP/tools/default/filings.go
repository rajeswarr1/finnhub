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

func FilingsHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["symbol"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("symbol=%v", val))
		}
		if val, ok := args["cik"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("cik=%v", val))
		}
		if val, ok := args["accessNumber"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("accessNumber=%v", val))
		}
		if val, ok := args["form"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("form=%v", val))
		}
		if val, ok := args["from"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("from=%v", val))
		}
		if val, ok := args["to"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("to=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/stock/filings%s", cfg.BaseURL, queryString)
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

func CreateFilingsTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_stock_filings",
		mcp.WithDescription("SEC Filings"),
		mcp.WithString("symbol", mcp.Description("Symbol. Leave <code>symbol</code>,<code>cik</code> and <code>accessNumber</code> empty to list latest filings.")),
		mcp.WithString("cik", mcp.Description("CIK.")),
		mcp.WithString("accessNumber", mcp.Description("Access number of a specific report you want to retrieve data from.")),
		mcp.WithString("form", mcp.Description("Filter by form. You can use this value <code>NT 10-K</code> to find non-timely filings for a company.")),
		mcp.WithString("from", mcp.Description("From date: 2023-03-15.")),
		mcp.WithString("to", mcp.Description("To date: 2023-03-16.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    FilingsHandler(cfg),
	}
}
