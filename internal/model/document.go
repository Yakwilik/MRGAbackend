package model

type DocumentField struct {
	Descr       string   `json:"descr"`
	IsPrefilled bool     `json:"is_prefilled"`
	Value       string   `json:"value"`
	MaxLength   uint     `json:"max_length,omitempty"`
	IsRequired  bool     `json:"is_required"`
	Tag         string   `json:"tag"`
	Typ         string   `json:"type"`
	Options     []string `json:"options,omitempty"`
}

type DocumentFieldsBlock struct {
	Title  string          `json:"title"`
	Fields []DocumentField `json:"fields"`
}

type Document struct {
	FieldsBlocks []DocumentFieldsBlock `json:"fields_blocks"`
}

type DocumentInfo struct {
	CategoryName string `json:"category_name"`
	TypeName     string `json:"type_name"`
	Key          string `json:"key"`
	Name         string `json:"name"`
}

type DocumentCategory struct {
	CategoryName  string         `json:"category_name"`
	DocumentTypes []DocumentType `json:"document_types"`
}

type CreateTypesRequest struct {
	CategoryName  string   `json:"category_name"`
	DocumentTypes []string `json:"document_types"`
}

type CreateVariantsRequest struct {
	TypeName         string            `json:"type_name"`
	DocumentVariants []DocumentVariant `json:"variants"`
}

type DocumentType struct {
	TypeName string            `json:"type_name"`
	Variants []DocumentVariant `json:"variants"`
}

type DocumentVariant struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}
