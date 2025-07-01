package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/itsadijmbt/simple_bank/db/util"
)

// ^ fl is an interface that contains all info
var validCurrency validator.Func = func(feildlevel validator.FieldLevel) bool {

	// ! this returs true
	if curr, ok := feildlevel.Field().Interface().(string); ok {
		//check if supported
		return util.IsSupportedCurrency(curr)
	}
	return false
}
