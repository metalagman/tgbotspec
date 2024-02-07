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
}

type Discriminator struct {
	PropertyName string            `yaml:"propertyName"`
	Mapping      map[string]string `yaml:"mapping,omitempty"`
}
