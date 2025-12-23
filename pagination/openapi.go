package pagination

// OpenAPISchema represents OpenAPI 3.0 schema components for pagination.
// This can be used to generate API documentation.
type OpenAPISchema struct {
	QueryParameters map[string]QueryParameterSchema `json:"query_parameters"`
	ResponseSchema  ResponseSchema                  `json:"response_schema"`
}

// QueryParameterSchema defines the schema for a query parameter.
type QueryParameterSchema struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Minimum     *int        `json:"minimum,omitempty"`
	Maximum     *int        `json:"maximum,omitempty"`
	Required    bool        `json:"required"`
}

// ResponseSchema defines the schema for a paginated response.
type ResponseSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

// GenerateOpenAPISchema generates OpenAPI 3.0 schema for pagination.
// itemSchema should be the JSON schema for the paginated item type.
func GenerateOpenAPISchema(cfg Config, itemSchema map[string]interface{}) OpenAPISchema {
	return OpenAPISchema{
		QueryParameters: map[string]QueryParameterSchema{
			"page": {
				Type:        "integer",
				Description: "Page number (1-indexed)",
				Default:     cfg.DefaultPage,
				Minimum:     intPtr(1),
				Required:    false,
			},
			"page_size": {
				Type:        "integer",
				Description: "Number of items per page",
				Default:     cfg.DefaultPageSize,
				Minimum:     intPtr(cfg.MinPageSize),
				Maximum:     intPtr(cfg.MaxPageSize),
				Required:    false,
			},
			"sort": {
				Type:        "string",
				Description: "Field name to sort by",
				Required:    false,
			},
			"order": {
				Type:        "string",
				Description: "Sort order: 'asc' or 'desc'",
				Default:     "asc",
				Required:    false,
			},
		},
		ResponseSchema: ResponseSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"data": map[string]interface{}{
					"type":  "array",
					"items": itemSchema,
				},
				"pagination": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"page": map[string]interface{}{
							"type":        "integer",
							"description": "Current page number",
						},
						"page_size": map[string]interface{}{
							"type":        "integer",
							"description": "Number of items per page",
						},
						"total_records": map[string]interface{}{
							"type":        "integer",
							"format":      "int64",
							"description": "Total number of records",
						},
						"total_pages": map[string]interface{}{
							"type":        "integer",
							"description": "Total number of pages",
						},
						"has_next": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether there is a next page",
						},
						"has_prev": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether there is a previous page",
						},
						"next_page": map[string]interface{}{
							"type":        "integer",
							"nullable":    true,
							"description": "Next page number, or null if no next page",
						},
						"prev_page": map[string]interface{}{
							"type":        "integer",
							"nullable":    true,
							"description": "Previous page number, or null if no previous page",
						},
					},
					"required": []string{
						"page", "page_size", "total_records", "total_pages",
						"has_next", "has_prev",
					},
				},
			},
			"required": []string{"data", "pagination"},
		},
	}
}

// intPtr returns a pointer to an integer.
func intPtr(i int) *int {
	return &i
}

