// Package server assembles the HTTP layer from the modules' route contributions.
package server

import (
	"mime/multipart"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/validator"
	"github.com/libtnb/validator/contrib/openapi"

	"github.com/weavatar/weavatar/internal/shared/registry"
)

var pathParams = regexp.MustCompile(`:([A-Za-z0-9_]+)`)

// Version is the build version, injected by main; the OpenAPI document carries it.
type Version string

// NewVersion normalizes the build-time version for generated documentation.
func NewVersion(version string) Version {
	if version == "" {
		return "dev"
	}
	return Version(version)
}

// HTTP registers every route contribution on r, in contribution order; an
// endpoint's middlewares run before its handler.
func HTTP(groups registry.Routes, r fiber.Router) {
	for _, endpoints := range groups {
		for _, e := range endpoints {
			handlers := make([]any, 0, len(e.Middlewares)+1)
			for _, m := range e.Middlewares {
				handlers = append(handlers, m)
			}
			handlers = append(handlers, e.Handler)
			r.Add([]string{e.Method}, e.Path, handlers[0], handlers[1:]...)
		}
	}
}

// SpecJSON assembles the OpenAPI 3.1 document from every documented endpoint.
func SpecJSON(title string, version Version, validate *validator.Validator, groups registry.Routes) ([]byte, error) {
	g, err := openapi.New(title, string(version),
		openapi.WithValidator(validate),
		openapi.WithSchema[time.Time](&openapi.Schema{Type: "string", Format: "date-time"}),
		// Uploads bind from multipart forms; document the file, not the header struct.
		openapi.WithSchemaTransform[multipart.FileHeader](func(s *openapi.Schema) error {
			*s = openapi.Schema{Type: "string", Format: "binary"}
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}
	for _, endpoints := range groups {
		for _, e := range endpoints {
			if e.Document == nil {
				continue
			}
			if err := e.Document(g, e.Method, pathParams.ReplaceAllString(e.Path, "{$1}"), e.Summary, e.Tags); err != nil {
				return nil, err
			}
		}
	}

	return g.JSON()
}
