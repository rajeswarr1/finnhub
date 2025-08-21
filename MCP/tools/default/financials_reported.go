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

func Financials_reportedHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
		if val, ok := args["freq"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("freq=%v", val))
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
		url := fmt.Sprintf("%s/stock/financials-reported%s", cfg.BaseURL, queryString)
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

func CreateFinancials_reportedTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_stock_financials-reported",
		mcp.WithDescription("Financials As Reported"),
		mcp.WithString("symbol", mcp.Description("Symbol.")),
		mcp.WithString("cik", mcp.Description("CIK.")),
		mcp.WithString("accessNumber", mcp.Description("Access number of a specific report you want to retrieve financials from.")),
		mcp.WithString("freq", mcp.Description("Frequency. Can be either <code>annual</code> or <code>quarterly</code>. Default to <code>annual</code>.")),
		mcp.WithString("from", mcp.Description("From date <code>YYYY-MM-DD</code>. Filter for endDate.")),
		mcp.WithString("to", mcp.Description("To date <code>YYYY-MM-DD</code>. Filter for endDate.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    Financials_reportedHandler(cfg),
	}
}
