package homework

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

var (
	ErrNotStruct                   = errors.New("wrong argument given, should be a struct")
	ErrInvalidValidatorSyntax      = errors.New("invalid validator syntax")
	ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")
	ErrLenValidationFailed         = errors.New("len validation failed")
	ErrInValidationFailed          = errors.New("in validation failed")
	ErrMaxValidationFailed         = errors.New("max validation failed")
	ErrMinValidationFailed         = errors.New("min validation failed")
)

type ValidationError struct {
	field string
	err   error
}

func NewValidationError(err error, field string) error {
	return &ValidationError{
		field: field,
		err:   err,
	}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.field, e.err)
}

func (e *ValidationError) Unwrap() error {
	return e.err
}

type Validator[t any, k any] struct {
	value    t
	template k
	outcome  error
}

func validateString(str, validator, template string) error {

	if validator == "len" {
		expected, err := strconv.Atoi(template)
		if err != nil {
			return ErrInvalidValidatorSyntax
		}
		if expected < 0 {
			return ErrInvalidValidatorSyntax
		}
		if len(str) != expected {
			return ErrLenValidationFailed
		}
	}

	if validator == "in" {
		constraints := strings.Split(template, ",")
		if slices.Index(constraints, str) < 0 {
			return ErrInValidationFailed
		}
	}

	if validator == "min" {
		expected, err := strconv.Atoi(template)
		if err != nil {
			return ErrInvalidValidatorSyntax
		}
		if len(str) < expected {
			return ErrMinValidationFailed
		}
	}

	if validator == "max" {
		expected, err := strconv.Atoi(template)
		if err != nil {
			return ErrInvalidValidatorSyntax
		}
		if len(str) > expected {
			return ErrMaxValidationFailed
		}
	}

	return nil
}

func validateInt(val int64, validator, template string) error {
	if validator == "in" {
		constraintsStr := strings.Split(template, ",")

		constraints := make([]int64, 0, len(constraintsStr))
		for _, v := range constraintsStr {
			intV, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return ErrInvalidValidatorSyntax
			}
			constraints = append(constraints, intV)
		}
		if slices.Index(constraints, val) < 0 {
			return ErrInValidationFailed
		}
	}

	if validator == "min" {
		expected, err := strconv.ParseInt(template, 10, 64)
		if err != nil {
			return ErrInvalidValidatorSyntax
		}
		if val < expected {
			return ErrMinValidationFailed
		}
	}

	if validator == "max" {
		expected, err := strconv.ParseInt(template, 10, 64)
		if err != nil {
			return ErrInvalidValidatorSyntax
		}
		if val > expected {
			return ErrMaxValidationFailed
		}
	}

	return nil
}

func validateField(fv reflect.Value, ft reflect.StructField) error {
	tag, ok := ft.Tag.Lookup("validate")
	if !ok {
		return nil
	}

	kv := strings.Split(tag, ":")
	if len(kv) != 2 || len(kv[0]) == 0 || len(kv[1]) == 0 {
		return NewValidationError(ErrInvalidValidatorSyntax, tag)
	}

	switch ft.Type.Kind() {
	case reflect.String:
		return validateString(fv.String(), kv[0], kv[1])
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		return validateInt(fv.Int(), kv[0], kv[1])
	case reflect.Struct:
		return Validate(fv.Interface())
	default:
		return nil
	}
}

func Validate(x any) error {
	valOfX := reflect.ValueOf(x)
	typeOfX := reflect.TypeOf(x)
	kindOfX := valOfX.Kind()

	if kindOfX != reflect.Struct {
		return ErrNotStruct
	}

	result := make([]error, 0)
	for i := range valOfX.NumField() {
		fv := valOfX.Field(i)
		ft := typeOfX.Field(i)

		if !ft.IsExported() {
			if _, ok := ft.Tag.Lookup("validate"); ok {
				result = append(result, NewValidationError(ErrValidateForUnexportedFields, ft.Name))
			}
			continue
		}

		if err := validateField(fv, ft); err != nil {
			result = append(result, NewValidationError(err, ft.Name))
		}
	}
	return errors.Join(result...)
}
