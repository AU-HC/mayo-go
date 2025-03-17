package mayo

import (
	"mayo-go/field"
)

type Mayo struct {
	field *field.Field
}

// InitMayo initializes mayo with the correct parameters according to the specification. Note that
// mayo has 4 levels: 1, 2, 3, and 5.
func InitMayo() *Mayo {
	return &Mayo{
		field: field.InitField(),
	}
}
