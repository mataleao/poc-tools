package poctools

import (
	"fmt"
	"strings"
)

type FieldBuilder struct {
	tableAlias string
	fields     []string
	exclude    []string
}

func CreateFieldBuilder() *FieldBuilder {
	return &FieldBuilder{}
}

func (Builder *FieldBuilder) SetAlias(alias string) *FieldBuilder {
	Builder.tableAlias = alias
	return Builder
}

func (Builder *FieldBuilder) SetFields(fields []string) *FieldBuilder {
	Builder.fields = fields
	return Builder
}

func (Builder *FieldBuilder) ExcludeFields(fieldsToExclude []string) *FieldBuilder {
	Builder.exclude = fieldsToExclude
	return Builder
}

func (Builder *FieldBuilder) Build() string {

	finalFields := make([]string, 0)

	for _, fieldElement := range Builder.fields {
		foundFlag := false
		for _, excludeElement := range Builder.exclude {
			if excludeElement == fieldElement {
				foundFlag = true
			}
		}
		if !foundFlag {
			finalFields = append(finalFields, fieldElement)
		}
	}

	return strings.Join(finalFields, fmt.Sprintf(", %s.", Builder.tableAlias))
}
