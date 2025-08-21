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

func Global_filings_search_filterHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["field"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("field=%v", val))
		}
		if val, ok := args["source"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("source=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/global-filings/filter%s", cfg.BaseURL, queryString)
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

func CreateGlobal_filings_search_filterTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_global-filings_filter",
		mcp.WithDescription("Search Filter"),
		mcp.WithString("field", mcp.Required(), mcp.Description("Field to get available filters. Available filters are \"countries\", \"exchanges\", \"exhibits\", \"forms\", \"gics\", \"naics\", \"caps\", \"acts\", and \"sort\".")),
		mcp.WithString("source", mcp.Description("Get available forms for each source.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    Global_filings_search_filterHandler(cfg),
	}
}
