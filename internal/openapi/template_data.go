package openapi

// TemplateData carries the information required to render the OpenAPI
// specification template. It intentionally avoids dependencies on the parser
// package so the renderer can live entirely within the openapi package.
type TemplateData struct {
	Title   string
	Version string
	Methods []Method
	Types   []Type
}

// Method captures the data needed to describe a Telegram Bot API method in the
// OpenAPI document.
type Method struct {
	Name        string
	Tags        []string
	Description []string
	Params      []MethodParam
	Return      *TypeSpec
}

// MethodParam describes a single parameter for a Telegram Bot API method.
type MethodParam struct {
	Name        string
	Description string
	Required    bool
	Schema      *TypeSpec
}

// Type models a Telegram Bot API object definition in the OpenAPI document.
type Type struct {
	Name        string
	Tag         string
	Description []string
	Fields      []TypeField
}

// TypeField represents a field within a Telegram Bot API object definition.
type TypeField struct {
	Name        string
	Description string
	Required    bool
	Schema      *TypeSpec
}
