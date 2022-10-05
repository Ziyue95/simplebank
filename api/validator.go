package api

import (
	"db.sqlc.dev/app/util"
	"github.com/go-playground/validator/v10"
)

// validator.Func: function that takes a validator.FieldLevel interface as input and return true when validation succeeds
// validator.FieldLevel: interface that contains all information and helper functions to validate a field.
var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	// call fieldLevel.Field().Interface() to get the value of the field as an interface{}
	// use .(string) to convert value to a string
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		return util.IsSupportedCurrency(currency)
	}
	return false
}
