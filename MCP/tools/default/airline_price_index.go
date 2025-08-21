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

func Airline_price_indexHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["airline"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("airline=%v", val))
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
		url := fmt.Sprintf("%s/airline/price-index%s", cfg.BaseURL, queryString)
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

func CreateAirline_price_indexTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_airline_price-index",
		mcp.WithDescription("Airline Price Index"),
		mcp.WithString("airline", mcp.Required(), mcp.Description("Filter data by airline. Accepted values: <code>united</code>,<code>delta</code>,<code>american_airlines</code>,<code>southwest</code>,<code>southern_airways_express</code>,<code>alaska_airlines</code>,<code>frontier_airlines</code>,<code>jetblue_airways</code>,<code>spirit_airlines</code>,<code>sun_country_airlines</code>,<code>breeze_airways</code>,<code>hawaiian_airlines</code>")),
		mcp.WithString("from", mcp.Required(), mcp.Description("From date <code>YYYY-MM-DD</code>.")),
		mcp.WithString("to", mcp.Required(), mcp.Description("To date <code>YYYY-MM-DD</code>.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    Airline_price_indexHandler(cfg),
	}
}
