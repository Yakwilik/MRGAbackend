package model

type DocumentPage struct {
	Content string `json:"content"`
}

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
	Pages        []DocumentPage        `json:"pages"`
	FieldsBlocks []DocumentFieldsBlock `json:"fields_blocks"`
}
