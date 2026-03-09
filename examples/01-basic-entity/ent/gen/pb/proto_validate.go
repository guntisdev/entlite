package pb

import (
	"fmt"
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

