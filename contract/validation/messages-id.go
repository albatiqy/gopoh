package validation

import (
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/albatiqy/gopoh/contract"
)

type messageTranslator struct {
	fieldLabel contract.FieldLabel
	messages   map[string]string
}

func (t messageTranslator) translate(fieldError validator.FieldError) string {
	message, ok := t.messages[fieldError.ActualTag()]
	if !ok {
		return fieldError.Error()
	}
	//escreplacer := strings.NewReplacer("{{", "#ld", "}}", "#rd")
	replacer := strings.NewReplacer(
		"{:field}",
		t.getLabel(fieldError),
		"{:param}",
		fieldError.Param(),
	)
	return replacer.Replace(message)
}

func (t messageTranslator) getLabel(fieldError validator.FieldError) string {
	if t.fieldLabel == nil {
		return fieldError.Field()
	}
	return t.fieldLabel.GetLabel(fieldError.StructField())
}

func newTranslator(fieldLabel contract.FieldLabel) *messageTranslator {
	return &messageTranslator{
		fieldLabel: fieldLabel,
		messages: map[string]string{
			"required": "{:field} wajib diisi",
			"email":    "{:field} harus berupa alamat email yang valid",
			"url":      "{:field} harus berupa URL yang valid",
			"nik":      "format {:field} tidak valid",
			"nip":      "format {:field} tidak valid",
			"npwp":     "format {:field} tidak valid",
			"nocell":   "nomor hape tidak valid",
		},
	}
}
