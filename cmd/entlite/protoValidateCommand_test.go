package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guntisdev/entlite/internal/util"
)

func TestProtoValidateCommand(t *testing.T) {

	tmpDir := t.TempDir()

	schemaDir := filepath.Join(tmpDir, "ent", "schema")
	logicDir := filepath.Join(tmpDir, "ent", "logic")
	pbDir := filepath.Join(tmpDir, "ent", "gen", "pb")

	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		t.Fatalf("Failed to create schema directory: %v", err)
	}

	if err := os.MkdirAll(logicDir, 0755); err != nil {
		t.Fatalf("Failed to create logic directory: %v", err)
	}

	if err := os.MkdirAll(pbDir, 0755); err != nil {
		t.Fatalf("Failed to create pb directory: %v", err)
	}

	// Write common test input files
	writeTestGoMod(t, tmpDir)
	writeTestUserSchema(t, schemaDir)
	writeTestLogic(t, logicDir)

	// Change to tmpDir/ent to simulate working in the project
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(filepath.Join(tmpDir, "ent")); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalDir)

	bufGenYamlContent := `version: v2
plugins:
  - remote: buf.build/protocolbuffers/go:v1.34.2
    out: gen/pb
    opt: paths=source_relative
  - remote: buf.build/connectrpc/go
    out: gen/pb
    opt:
      - paths=source_relative
      - package_suffix=
`
	bufGenYamlPath := filepath.Join(tmpDir, "ent", "buf.gen.yaml")
	if err := os.WriteFile(bufGenYamlPath, []byte(bufGenYamlContent), 0644); err != nil {
		t.Fatalf("Failed to write buf.gen.yaml file: %v", err)
	}

	protoValidate()

	validatePath := filepath.Join(pbDir, "proto_validate.go")
	if _, err := os.Stat(validatePath); os.IsNotExist(err) {
		t.Fatalf("Expected proto_validate.go was not created at %s", validatePath)
	}

	actualContent, err := os.ReadFile(validatePath)
	if err != nil {
		t.Fatalf("Failed to read generated proto_validate.go: %v", err)
	}

	expectedContent := `package pb

import (
	"context"
	"fmt"
	"connectrpc.com/connect"
	"github.com/guntisdev/entlite/examples/01-basic-entity/ent/logic"
)

func (r *CreateUserRequest) Validate() error {
	if !logic.StartsWithCapital(r.Name) {
		return fmt.Errorf("Validation failed for field name: Name")
	}
	return nil
}

func (r *UpdateUserRequest) Validate() error {
	if !logic.StartsWithCapital(r.Name) {
		return fmt.Errorf("Validation failed for field name: Name")
	}
	return nil
}

type validator interface {
	Validate() error
}

// ValidateInterceptor calls the generated Validate() method on any request message that implements it
type ValidateInterceptor struct{}

var _ connect.Interceptor = (*ValidateInterceptor)(nil)

func NewValidateInterceptor() *ValidateInterceptor {
	return &ValidateInterceptor{}
}

func validateMsg(msg any) error {
	v, ok := msg.(validator)
	if !ok {
		return nil
	}
	if err := v.Validate(); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	return nil
}

// WrapUnary implements connect.Interceptor.
func (i *ValidateInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if err := validateMsg(req.Any()); err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

// WrapStreamingClient implements connect.Interceptor.
func (i *ValidateInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler implements connect.Interceptor.
func (i *ValidateInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, &validatingHandlerConn{StreamingHandlerConn: conn})
	}
}

type validatingHandlerConn struct {
	connect.StreamingHandlerConn
}

func (c *validatingHandlerConn) Receive(msg any) error {
	if err := c.StreamingHandlerConn.Receive(msg); err != nil {
		return err
	}
	return validateMsg(msg)
}
`

	if d := util.Diff(expectedContent, string(actualContent)); d != "" {
		t.Errorf("proto_validate.go content mismatch (-expected +actual):\n%s", d)
	}
}
