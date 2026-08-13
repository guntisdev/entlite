package pb

import (
	"context"
	"encoding/json"
	"fmt"
	"connectrpc.com/connect"
	"github.com/guntisdev/entlite/examples/03-optional/sqlite/ent/logic"
)

func (r *CreateArticleRequest) Validate() error {
	if r.Metadata != nil && !json.Valid([]byte(*r.Metadata)) {
		return fmt.Errorf("Invalid json for field name: Metadata")
	}
	if !logic.NotBlank(r.Title) {
		return fmt.Errorf("Validation failed for field name: Title")
	}
	return nil
}

func (r *UpdateArticleRequest) Validate() error {
	if r.Metadata != nil && !json.Valid([]byte(*r.Metadata)) {
		return fmt.Errorf("Invalid json for field name: Metadata")
	}
	if !logic.NotBlank(r.Title) {
		return fmt.Errorf("Validation failed for field name: Title")
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
