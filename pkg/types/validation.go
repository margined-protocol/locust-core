package types

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ValidateConfig checks if the configuration is valid.
func ValidateConfig(cfg Config) error {
	if cfg.Chain.Fees != nil && (cfg.Chain.GasAdjustment != nil || cfg.Chain.Gas != nil || cfg.Chain.GasPrices != nil) {
		return errors.New("if 'fees' is provided, 'gas', 'gas_adjustment', and 'gas_prices' must not be provided")
	}

	if cfg.Chain.Fees == nil && (cfg.Chain.GasAdjustment == nil || cfg.Chain.Gas == nil || cfg.Chain.GasPrices == nil) {
		return errors.New("'gas', 'gas_adjustment', and 'gas_prices' must all be provided together if 'fees' is not used")
	}

	return nil
}

func validateStructField(field reflect.Value) error {
	if field.Kind() == reflect.Struct {
		// Recursively check nested structs
		return validateStructFields(field)
	}
	return nil
}

// validateStructFields recursively checks if any field in the struct is zero-valued or invalid.
func validateStructFields(v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanInterface() {
			continue // Skip unexported fields
		}

		fieldType := v.Type().Field(i)

		// Check if the field is a struct
		if err := validateStructField(field); err != nil {
			return fmt.Errorf("%s: %w", fieldType.Name, err)
		}

		// Check if the field is zero-valued or invalid
		if reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
			return fmt.Errorf("field %s is missing or invalid", fieldType.Name)
		}

		// Additional validation for string fields
		if field.Kind() == reflect.String && strings.TrimSpace(field.String()) == "" {
			return fmt.Errorf("field %s is empty", fieldType.Name)
		}

		// Additional validation for int and float fields
		if field.Kind() == reflect.Int64 && field.Int() <= 0 {
			return fmt.Errorf("field %s should be greater than zero", fieldType.Name)
		}
		if field.Kind() == reflect.Float64 && field.Float() <= 0 {
			return fmt.Errorf("field %s should be greater than zero", fieldType.Name)
		}
	}
	return nil
}
