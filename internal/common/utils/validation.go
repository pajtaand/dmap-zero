package utils

import (
	"fmt"
	"reflect"
)

func CheckStringNotEmpty(s interface{}, field string) error {
	r := reflect.ValueOf(s)
	v := reflect.Indirect(r).FieldByName(field)
	if v.String() == "" {
		return fmt.Errorf("field '%s' is empty", field)
	}
	return nil
}

func CheckNotNil(s interface{}, field string) error {
	r := reflect.ValueOf(s)
	v := reflect.Indirect(r).FieldByName(field)
	if v.IsNil() {
		return fmt.Errorf("field '%s' not set", field)
	}
	return nil
}
