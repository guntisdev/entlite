package pb

import (
	"context"
	"fmt"
	"connectrpc.com/connect"
	"github.com/guntisdev/entlite/examples/01-sqlite-entity/ent/logic"
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
