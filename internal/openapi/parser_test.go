package openapi

import (
	"testing"
)

func TestParser_ParseBytes(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantOps     int
		wantErr     bool
		validateDoc func(*testing.T, *Document)
	}{
		{
			name: "parses minimal spec with single endpoint",
			// given
			yaml: `
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: listUsers
      summary: List users
      responses:
        '200':
          description: OK
`,
			wantOps: 1,
			wantErr: false,
			// then
			validateDoc: func(t *testing.T, doc *Document) {
				if doc.Title != "Test API" {
					t.Errorf("Title = %q, want %q", doc.Title, "Test API")
				}
				if doc.Version != "1.0.0" {
					t.Errorf("Version = %q, want %q", doc.Version, "1.0.0")
				}

				op, ok := doc.Operations["GET:/users"]
				if !ok {
					t.Fatal("missing operation GET:/users")
				}
				if op.OperationID != "listUsers" {
					t.Errorf("OperationID = %q, want %q", op.OperationID, "listUsers")
				}
				if op.Summary != "List users" {
					t.Errorf("Summary = %q, want %q", op.Summary, "List users")
				}
			},
		},
		{
			name: "parses spec with multiple methods on same path",
			// given
			yaml: `
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: listUsers
      responses:
        '200':
          description: OK
    post:
      operationId: createUser
      responses:
        '201':
          description: Created
`,
			wantOps: 2,
			wantErr: false,
			// then
			validateDoc: func(t *testing.T, doc *Document) {
				if _, ok := doc.Operations["GET:/users"]; !ok {
					t.Error("missing operation GET:/users")
				}
				if _, ok := doc.Operations["POST:/users"]; !ok {
					t.Error("missing operation POST:/users")
				}
			},
		},
		{
			name: "parses spec with path parameters",
			// given
			yaml: `
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{id}:
    get:
      operationId: getUser
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
`,
			wantOps: 1,
			wantErr: false,
			// then
			validateDoc: func(t *testing.T, doc *Document) {
				op := doc.Operations["GET:/users/{id}"]
				if op == nil {
					t.Fatal("missing operation GET:/users/{id}")
				}
				if len(op.Parameters) != 1 {
					t.Fatalf("Parameters count = %d, want 1", len(op.Parameters))
				}
				param := op.Parameters[0]
				if param.Name != "id" {
					t.Errorf("Parameter.Name = %q, want %q", param.Name, "id")
				}
				if param.In != "path" {
					t.Errorf("Parameter.In = %q, want %q", param.In, "path")
				}
				if !param.Required {
					t.Error("Parameter.Required = false, want true")
				}
			},
		},
		{
			name: "parses spec with request body",
			// given
			yaml: `
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      operationId: createUser
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                email:
                  type: string
                password:
                  type: string
      responses:
        '201':
          description: Created
`,
			wantOps: 1,
			wantErr: false,
			// then
			validateDoc: func(t *testing.T, doc *Document) {
				op := doc.Operations["POST:/users"]
				if op == nil {
					t.Fatal("missing operation POST:/users")
				}
				if op.RequestBody == nil {
					t.Fatal("RequestBody is nil")
				}
				if !op.RequestBody.Required {
					t.Error("RequestBody.Required = false, want true")
				}
				content, ok := op.RequestBody.Content["application/json"]
				if !ok {
					t.Fatal("missing content type application/json")
				}
				if content.Schema == nil {
					t.Fatal("Schema is nil")
				}
				if content.Schema.Type != "object" {
					t.Errorf("Schema.Type = %q, want %q", content.Schema.Type, "object")
				}
				if len(content.Schema.Properties) != 2 {
					t.Errorf("Properties count = %d, want 2", len(content.Schema.Properties))
				}
			},
		},
		{
			name: "parses spec with response schema",
			// given
			yaml: `
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{id}:
    get:
      operationId: getUser
      responses:
        '200':
          description: User found
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  email:
                    type: string
        '404':
          description: User not found
`,
			wantOps: 1,
			wantErr: false,
			// then
			validateDoc: func(t *testing.T, doc *Document) {
				op := doc.Operations["GET:/users/{id}"]
				if op == nil {
					t.Fatal("missing operation GET:/users/{id}")
				}
				resp200, ok := op.Responses["200"]
				if !ok {
					t.Fatal("missing response 200")
				}
				if resp200.Description != "User found" {
					t.Errorf("Response.Description = %q, want %q", resp200.Description, "User found")
				}
				content, ok := resp200.Content["application/json"]
				if !ok {
					t.Fatal("missing content type application/json in response")
				}
				if content.Schema == nil || content.Schema.Type != "object" {
					t.Error("response schema should be object type")
				}

				resp404, ok := op.Responses["404"]
				if !ok {
					t.Fatal("missing response 404")
				}
				if resp404.Description != "User not found" {
					t.Errorf("Response.Description = %q, want %q", resp404.Description, "User not found")
				}
			},
		},
		{
			name: "parses spec with $ref schemas",
			// given
			yaml: `
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      operationId: createUser
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    CreateUserRequest:
      type: object
      properties:
        email:
          type: string
    User:
      type: object
      properties:
        id:
          type: string
        email:
          type: string
`,
			wantOps: 1,
			wantErr: false,
			// then
			validateDoc: func(t *testing.T, doc *Document) {
				op := doc.Operations["POST:/users"]
				if op == nil {
					t.Fatal("missing operation POST:/users")
				}
				content := op.RequestBody.Content["application/json"]
				if content.Schema == nil {
					t.Fatal("RequestBody schema is nil")
				}
				if !content.Schema.IsRef() {
					t.Error("RequestBody schema should be a $ref")
				}
				if content.Schema.RefName() != "CreateUserRequest" {
					t.Errorf("RefName = %q, want %q", content.Schema.RefName(), "CreateUserRequest")
				}

				respContent := op.Responses["201"].Content["application/json"]
				if respContent.Schema.RefName() != "User" {
					t.Errorf("Response RefName = %q, want %q", respContent.Schema.RefName(), "User")
				}
			},
		},
		{
			name:    "fails on invalid YAML",
			yaml:    `invalid: yaml: syntax`,
			wantErr: true,
		},
		{
			name: "handles response without description",
			// given
			yaml: `
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /health:
    get:
      operationId: healthCheck
      responses:
        '200': {}
`,
			wantOps: 1,
			wantErr: false,
			// then
			validateDoc: func(t *testing.T, doc *Document) {
				op := doc.Operations["GET:/health"]
				if op == nil {
					t.Fatal("missing operation GET:/health")
				}
				resp, ok := op.Responses["200"]
				if !ok {
					t.Fatal("missing response 200")
				}
				// Description should be empty string, not cause panic
				if resp.Description != "" {
					t.Errorf("Description = %q, want empty string", resp.Description)
				}
			},
		},
		{
			name: "handles schema with only $ref (no explicit type)",
			// given
			yaml: `
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: listUsers
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserList'
components:
  schemas:
    UserList:
      type: array
      items:
        $ref: '#/components/schemas/User'
    User:
      type: object
`,
			wantOps: 1,
			wantErr: false,
			// then
			validateDoc: func(t *testing.T, doc *Document) {
				op := doc.Operations["GET:/users"]
				content := op.Responses["200"].Content["application/json"]
				// Schema with $ref should not panic, should have empty Type
				if content.Schema == nil {
					t.Fatal("schema is nil")
				}
				if !content.Schema.IsRef() {
					t.Error("schema should be a ref")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			parser := NewParser("")
			doc, err := parser.ParseBytes([]byte(tt.yaml))

			// then
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(doc.Operations) != tt.wantOps {
				t.Errorf("Operations count = %d, want %d", len(doc.Operations), tt.wantOps)
			}

			if tt.validateDoc != nil {
				tt.validateDoc(t, doc)
			}
		})
	}
}

func TestParseBinding(t *testing.T) {
	tests := []struct {
		name       string
		bindsTo    string
		wantServer string
		wantMethod string
		wantPath   string
		wantErr    bool
	}{
		{
			name: "parses valid binding",
			// given
			bindsTo:    "http.server.api:POST:/users",
			wantServer: "http.server.api",
			wantMethod: "POST",
			wantPath:   "/users",
			wantErr:    false,
		},
		{
			name: "parses binding with path parameter",
			// given
			bindsTo:    "http.server.api:GET:/users/{id}",
			wantServer: "http.server.api",
			wantMethod: "GET",
			wantPath:   "/users/{id}",
			wantErr:    false,
		},
		{
			name: "parses binding with nested path",
			// given
			bindsTo:    "api.main:DELETE:/users/{userId}/posts/{postId}",
			wantServer: "api.main",
			wantMethod: "DELETE",
			wantPath:   "/users/{userId}/posts/{postId}",
			wantErr:    false,
		},
		{
			name:    "fails on empty binding",
			bindsTo: "",
			wantErr: true,
		},
		{
			name:    "fails on missing method",
			bindsTo: "http.server.api:/users",
			wantErr: true,
		},
		{
			name:    "fails on invalid method",
			bindsTo: "http.server.api:INVALID:/users",
			wantErr: true,
		},
		{
			name:    "fails on path without leading slash",
			bindsTo: "http.server.api:GET:users",
			wantErr: true,
		},
		{
			name:    "fails on missing path",
			bindsTo: "http.server.api:GET:",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			serverID, method, path, err := ParseBinding(tt.bindsTo)

			// then
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBinding() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if serverID != tt.wantServer {
				t.Errorf("serverID = %q, want %q", serverID, tt.wantServer)
			}
			if method != tt.wantMethod {
				t.Errorf("method = %q, want %q", method, tt.wantMethod)
			}
			if path != tt.wantPath {
				t.Errorf("path = %q, want %q", path, tt.wantPath)
			}
		})
	}
}

func TestOperationKey(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   string
	}{
		{"GET", "/users", "GET:/users"},
		{"POST", "/users", "POST:/users"},
		{"GET", "/users/{id}", "GET:/users/{id}"},
		{"DELETE", "/users/{id}/posts/{postId}", "DELETE:/users/{id}/posts/{postId}"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			// given method and path
			// when
			got := OperationKey(tt.method, tt.path)
			// then
			if got != tt.want {
				t.Errorf("OperationKey(%q, %q) = %q, want %q", tt.method, tt.path, got, tt.want)
			}
		})
	}
}

func TestSchema_RefName(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want string
	}{
		{
			name: "extracts name from components schema ref",
			// given
			ref:  "#/components/schemas/User",
			want: "User",
		},
		{
			name: "extracts name from nested ref",
			// given
			ref:  "#/components/schemas/api/UserRequest",
			want: "UserRequest",
		},
		{
			name: "handles empty ref",
			// given
			ref:  "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			schema := &Schema{Ref: tt.ref}
			got := schema.RefName()

			// then
			if got != tt.want {
				t.Errorf("RefName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOperation_OperationKey(t *testing.T) {
	// given
	op := &Operation{
		Method: "POST",
		Path:   "/users",
	}

	// when
	key := op.OperationKey()

	// then
	if key != "POST:/users" {
		t.Errorf("OperationKey() = %q, want %q", key, "POST:/users")
	}
}
