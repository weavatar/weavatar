package transport

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/validator"
)

// Bind binds the query, then the body, then path parameters, each overriding
// the last, so a case-insensitive query match such as "?code=x" cannot replace
// a JSON "code"; then it validates.
func Bind[T any](c fiber.Ctx, v *validator.Validator) (*T, error) {
	req := new(T)

	if err := c.Bind().Query(req); err != nil {
		return nil, err
	}
	switch c.Method() {
	case fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch, fiber.MethodDelete:
		if c.Request().Header.ContentLength() > 0 {
			if err := c.Bind().Body(req); err != nil {
				return nil, err
			}
		}
	}
	if err := c.Bind().URI(req); err != nil {
		return nil, err
	}

	if hook, ok := any(req).(WithPrepare); ok {
		if err := hook.Prepare(c); err != nil {
			return nil, err
		}
	}

	vd, err := v.Struct(req)
	if err != nil {
		return nil, err
	}
	if hook, ok := any(req).(WithRules); ok {
		for field, expr := range hook.Rules(c) {
			if err := vd.AddRules(field, expr); err != nil {
				return nil, err
			}
		}
	}
	if hook, ok := any(req).(WithFilters); ok {
		for field, filters := range hook.Filters(c) {
			if err := vd.AddFilters(field, filters); err != nil {
				return nil, err
			}
		}
	}
	if hook, ok := any(req).(WithMessages); ok {
		if messages := hook.Messages(c); messages != nil {
			if err := vd.AddMessages(messages); err != nil {
				return nil, err
			}
		}
	}

	// ValidateAs validates and atomically writes filtered values back.
	if err := vd.ValidateAs(c.Context(), req); err != nil {
		if fields, ok := validator.AsErrors(err); ok {
			return nil, errors.New(fields.One())
		}
		return nil, err
	}

	return req, nil
}
