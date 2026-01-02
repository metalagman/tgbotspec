package openapi

type TypeSpec struct {
	Type                 string              `yaml:"type,omitempty"`
	Format               string              `yaml:"format,omitempty"`
	Description          string              `yaml:"description,omitempty"`
	Properties           map[string]TypeSpec `yaml:"properties,omitempty"`
	Required             []string            `yaml:"required,omitempty"`
	OneOf                []TypeSpec          `yaml:"oneOf,omitempty"`
	AnyOf                []TypeSpec          `yaml:"anyOf,omitempty"`
	AllOf                []TypeSpec          `yaml:"allOf,omitempty"`
	Example              string              `yaml:"example,omitempty"`
	Items                *TypeSpec           `yaml:"items,omitempty"`
	Enum                 []interface{}       `yaml:"enum,omitempty"`
	Default              interface{}         `yaml:"default,omitempty"`
	Minimum              *float64            `yaml:"minimum,omitempty"`
	Maximum              *float64            `yaml:"maximum,omitempty"`
	MinItems             *int                `yaml:"minItems,omitempty"`
	MaxItems             *int                `yaml:"maxItems,omitempty"`
	MinLength            *int                `yaml:"minLength,omitempty"`
	MaxLength            *int                `yaml:"maxLength,omitempty"`
	Pattern              string              `yaml:"pattern,omitempty"`
	UniqueItems          bool                `yaml:"uniqueItems,omitempty"`
	AdditionalProperties interface{}         `yaml:"additionalProperties,omitempty"` // Can be boolean or TypeSpec
	Nullable             bool                `yaml:"nullable,omitempty"`
	ReadOnly             bool                `yaml:"readOnly,omitempty"`
	WriteOnly            bool                `yaml:"writeOnly,omitempty"`
	Title                string              `yaml:"title,omitempty"`
	MultipleOf           *float64            `yaml:"multipleOf,omitempty"`
	MaxProperties        *int                `yaml:"maxProperties,omitempty"`
	MinProperties        *int                `yaml:"minProperties,omitempty"`
	Discriminator        *Discriminator      `yaml:"discriminator,omitempty"`
	Ref                  *TypeRef            `yaml:"$ref,omitempty"`
}

// WithDescription returns a copy of the TypeSpec with the provided description.
// If the TypeSpec uses a $ref, it wraps it in an allOf to avoid having $ref
// and description as siblings, which is forbidden in OpenAPI 3.0.
func (s *TypeSpec) WithDescription(desc string) *TypeSpec {
	if s == nil {
		return nil
	}

	res := *s
	res.Description = desc

	if res.Ref != nil {
		refOnly := TypeSpec{Ref: res.Ref}

		res.Ref = nil
		res.AllOf = []TypeSpec{refOnly}
	}

	return &res
}

type Discriminator struct {
	PropertyName string            `yaml:"propertyName"`
	Mapping      map[string]string `yaml:"mapping,omitempty"`
}

// TypeRef is a minimal schema that only references another component schema via $ref.
// It can be useful when a plain reference is desired instead of an inline TypeSpec.
// Note: current generators mostly use TypeSpec.Ref, but this struct is provided for flexibility.
type TypeRef struct {
	Name string `yaml:"-"`
}

// MarshalYAML ensures that when TypeRef is serialized as the value of a
// TypeSpec.Ref field (yaml:"$ref"), it produces the proper OpenAPI reference
// string while the struct itself only stores the plain schema name.
func (tr TypeRef) MarshalYAML() (interface{}, error) {
	if tr.Name == "" {
		return nil, nil
	}

	return "#/components/schemas/" + tr.Name, nil
}
