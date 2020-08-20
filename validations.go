package validations

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

// NewError generate a new error for a model's field
func NewError(resource interface{}, column, err string) error {
	return &Error{Resource: resource, Column: column, Message: err}
}

// Error is a validation error struct that hold model, column and error message
type Error struct {
	Resource interface{}
	Column   string
	Message  string
}

// Label is a label including model type, primary key and column name
func (err Error) Label() string {
	stmt := gorm.Statement{}
	stmt.Parse(err.Resource)
	var vars = []string{}
	for _, field := range stmt.Schema.PrimaryFields {
		v, _ := field.ValueOf(reflect.ValueOf(err.Resource))
		vars = append(vars, fmt.Sprint(v))
	}
	return fmt.Sprintf("%v_%v_%v", stmt.Schema.ModelType.Name(), strings.Join(vars, "::"), err.Column)
}

// Error show error message
func (err Error) Error() string {
	return fmt.Sprintf("%v", err.Message)
}

// Validator ensures struct has validate functionality.
type Validator interface {
	Validate(db *gorm.DB)
}

// ValidatorWithError ensures struct has Validate functionality returns error as validation failure.
type ValidatorWithError interface {
	Validate(db *gorm.DB) error
}
