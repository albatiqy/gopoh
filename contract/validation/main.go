package validation

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/albatiqy/gopoh/contract"
	"github.com/albatiqy/gopoh/pkg/lib/null"
	isValid "github.com/albatiqy/gopoh/pkg/lib/validator"
)

type ValidationError interface {
	error
	ValidationMessages() map[string][]string
}

type inputError struct {
	validatorError validator.ValidationErrors
	message        string
	translator     *messageTranslator
}

func (o inputError) Error() string {
	return o.message
}

func (o inputError) ValidationMessages() map[string][]string {
	var fieldsMessages map[string][]string
	if len(o.validatorError) > 0 {
		fieldsMessages = map[string][]string{}
		for _, fieldError := range o.validatorError {
			attr := fieldError.Field()
			fieldsMessages[attr] = append(fieldsMessages[attr], o.translator.translate(fieldError))
		}
	}
	return fieldsMessages
}

type InputValidator struct {
	validator   *validator.Validate
	inputStruct interface{}
	translator  *messageTranslator
	err         error
}

func (v *InputValidator) CustomFieldRule(structField string, fn func(val interface{}) (bool, map[string]string)) { // validator.Func
	if v.err != nil {
		return
	}

	validationTagName := "custom" + structField
	rv := reflect.ValueOf(v.inputStruct)
	/*
		if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
			return nil, newError("v must be pointer to struct")
		}
		rv = rv.Elem()
	*/

	if rv.Kind() != reflect.Struct {
		v.err = fmt.Errorf("v must be struct")
		return
	}

	t := rv.Type()
	sf := make([]reflect.StructField, 0)
	for i := 0; i < t.NumField(); i++ {
		sf = append(sf, t.Field(i))
		if t.Field(i).Name == structField {
			validateTag, ok := sf[i].Tag.Lookup("validate")
			tagString := string(sf[i].Tag)

			var newTag string
			if ok {
				validatePos := strings.Index(tagString, "validate:\"")
				if validatePos >= 0 {
					depan := tagString[:validatePos]
					belakang := tagString[validatePos+len(validateTag)+11:]
					if validateTag != "" {
						newTag = depan + "validate:\"" + validateTag + "," + validationTagName + "\"" + belakang
					} else {
						newTag = depan + "validate:\"" + validationTagName + "\"" + belakang
					}
				} else {
					v.err = fmt.Errorf("error constructing validate tag")
					return
				}
			} else {
				if tagString != "" {
					newTag = tagString + " validate:\"" + validationTagName + "\""
				} else {
					newTag = "validate:\"" + validationTagName + "\""
				}
			}
			sf[i].Tag = reflect.StructTag(newTag)
		}
	}
	newType := reflect.StructOf(sf)
	newValue := rv.Convert(newType)

	v.validator.RegisterValidation(validationTagName, func(fl validator.FieldLevel) bool {
		ruleIsValid, ruleMessages := fn(fl.Field().Interface())
		if ruleMessages != nil {
			if msg, ok := ruleMessages["ID"]; ok { // lc ??
				v.translator.messages[validationTagName] = msg
			}
		}
		return ruleIsValid
	})

	v.inputStruct = newValue.Interface()
}

func (v *InputValidator) SetMessage(lc, tag, message string) {
	v.translator.messages[tag] = message
}

func (v InputValidator) Validate() error {
	if v.err != nil {
		return v.err
	}
	err := v.validator.Struct(v.inputStruct)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}
		return inputError{
			validatorError: err.(validator.ValidationErrors),
			message:        "",
			translator:     v.translator,
		}
	}
	return nil
}

func NewInputValidator(input interface{}, fieldsLabel contract.FieldsLabel, tags ...string) InputValidator {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(sf reflect.StructField) string {
		name := strings.SplitN(sf.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			name = sf.Name
		}
		return name
	})
	validate.RegisterCustomTypeFunc(func(rv reflect.Value) interface{} {
		if valuer, ok := rv.Interface().(driver.Valuer); ok {

			val, err := valuer.Value()
			if err == nil {
				return val
			}
			// handle the error how you want
		}
		return nil
	}, null.String{})

	for _, tag := range tags {
		switch tag {
		case "nik":
			validate.RegisterValidation("nik", func(fl validator.FieldLevel) bool {
				return isValid.NIK(fl.Field().String())
			})
		}
	}

	return InputValidator{
		validator:   validate,
		inputStruct: input,
		translator:  newTranslator(fieldsLabel),
	}
}
