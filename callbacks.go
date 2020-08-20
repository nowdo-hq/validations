package validations

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"gorm.io/gorm"
)

var skipValidations = "validations:skip_validations"

func validate(scope *gorm.DB) {
	if _, ok := scope.Get("gorm:update_column"); !ok {
		if result, ok := scope.Get(skipValidations); !(ok && result.(bool)) {
			if scope.Error == nil {
				var runValidation = func(model interface{}) {
					if v, ok := model.(Validator); ok {
						v.Validate(scope)
					}
					if v, ok := model.(ValidatorWithError); ok {
						if err := v.Validate(scope); err != nil {
							scope.AddError(err)
						}
					}
					if ok, err := govalidator.ValidateStruct(model); !ok {
						if errors, ok := err.(govalidator.Errors); ok {
							for _, err := range flatValidatorErrors(errors) {
								scope.AddError(formattedError(err, model))
							}
						} else {
							scope.AddError(err)
						}
					}
				}
				switch scope.Statement.ReflectValue.Kind() {
				case reflect.Slice, reflect.Array:
					for i := 0; i < scope.Statement.ReflectValue.Len(); i++ {
						runValidation(scope.Statement.ReflectValue.Index(i).Interface())
					}
				case reflect.Struct:
					runValidation(scope.Statement.Model)
				}
			}
		}
	}
}

func flatValidatorErrors(validatorErrors govalidator.Errors) []govalidator.Error {
	resultErrors := []govalidator.Error{}
	for _, validatorError := range validatorErrors.Errors() {
		if errors, ok := validatorError.(govalidator.Errors); ok {
			for _, e := range errors {
				resultErrors = append(resultErrors, e.(govalidator.Error))
			}
		}
		if e, ok := validatorError.(govalidator.Error); ok {
			resultErrors = append(resultErrors, e)
		}
	}
	return resultErrors
}

func formattedError(err govalidator.Error, resource interface{}) error {
	message := err.Error()
	attrName := err.Name
	if strings.Index(message, "non zero value required") >= 0 {
		message = fmt.Sprintf("%v can't be blank", attrName)
	} else if strings.Index(message, "as length") >= 0 {
		reg, _ := regexp.Compile(`\(([0-9]+)\|([0-9]+)\)`)
		submatch := reg.FindSubmatch([]byte(err.Error()))
		message = fmt.Sprintf("%v is the wrong length (should be %v~%v characters)", attrName, string(submatch[1]), string(submatch[2]))
	} else if strings.Index(message, "as numeric") >= 0 {
		message = fmt.Sprintf("%v is not a number", attrName)
	} else if strings.Index(message, "as email") >= 0 {
		message = fmt.Sprintf("%v is not a valid email address", attrName)
	}
	return NewError(resource, attrName, message)

}

// RegisterCallbacks register callback into GORM DB
func RegisterCallbacks(db *gorm.DB) {
	callback := db.Callback()
	if callback.Create().Get("validations:validate") == nil {
		callback.Create().Before("gorm:before_create").Register("validations:validate", validate)
	}
	if callback.Update().Get("validations:validate") == nil {
		callback.Update().Before("gorm:before_update").Register("validations:validate", validate)
	}
}
