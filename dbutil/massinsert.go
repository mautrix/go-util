// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil

import (
	"fmt"
	"regexp"
	"strings"
)

type Array interface {
	[1]any | [2]any | [3]any | [4]any | [5]any | [6]any | [7]any | [8]any | [9]any | [10]any | [11]any | [12]any | [13]any | [14]any | [15]any | [16]any | [17]any | [18]any | [19]any | [20]any
}

type MassInsertable[T Array] interface {
	GetMassInsertValues() T
}

type MassInsertBuilder[Item MassInsertable[DynamicParams], StaticParams Array, DynamicParams Array] struct {
	queryTemplate       string
	placeholderTemplate string
}

func NewMassInsertBuilder[Item MassInsertable[DynamicParams], StaticParams Array, DynamicParams Array](
	singleInsertQuery, placeholderTemplate string,
) *MassInsertBuilder[Item, StaticParams, DynamicParams] {
	var dyn DynamicParams
	var stat StaticParams
	totalParams := len(dyn) + len(stat)
	mainQueryVariablePlaceholderParts := make([]string, totalParams)
	for i := 0; i < totalParams; i++ {
		mainQueryVariablePlaceholderParts[i] = fmt.Sprintf(`\$%d`, i+1)
	}
	mainQueryVariablePlaceholderRegex := regexp.MustCompile(fmt.Sprintf(`\(\s*%s\s*\)`, strings.Join(mainQueryVariablePlaceholderParts, `\s*,\s*`)))
	queryPlaceholders := mainQueryVariablePlaceholderRegex.FindAllString(singleInsertQuery, -1)
	if len(queryPlaceholders) == 0 {
		panic(fmt.Errorf("invalid insert query: placeholders not found"))
	} else if len(queryPlaceholders) > 1 {
		panic(fmt.Errorf("invalid insert query: multiple placeholders found"))
	}
	for i := 0; i < len(stat); i++ {
		if !strings.Contains(placeholderTemplate, fmt.Sprintf("$%d", i+1)) {
			panic(fmt.Errorf("invalid placeholder template: static placeholder $%d not found", i+1))
		}
	}
	if strings.Contains(placeholderTemplate, fmt.Sprintf("$%d", len(stat)+1)) {
		panic(fmt.Errorf("invalid placeholder template: non-static placeholder $%d found", len(stat)+1))
	}
	fmtParams := make([]any, len(dyn))
	for i := 0; i < len(dyn); i++ {
		fmtParams[i] = fmt.Sprintf("$%d", len(stat)+i+1)
	}
	formattedPlaceholder := fmt.Sprintf(placeholderTemplate, fmtParams...)
	if strings.Contains(formattedPlaceholder, "!(EXTRA string=") {
		panic(fmt.Errorf("invalid placeholder template: extra string found"))
	}
	for i := 0; i < len(dyn); i++ {
		if !strings.Contains(formattedPlaceholder, fmt.Sprintf("$%d", len(stat)+i+1)) {
			panic(fmt.Errorf("invalid placeholder template: dynamic placeholder $%d not found", len(stat)+i+1))
		}
	}
	return &MassInsertBuilder[Item, StaticParams, DynamicParams]{
		queryTemplate:       strings.Replace(singleInsertQuery, queryPlaceholders[0], "%s", 1),
		placeholderTemplate: placeholderTemplate,
	}
}

func (mib *MassInsertBuilder[Item, StaticParams, DynamicParams]) Build(static StaticParams, data []Item) (query string, params []any) {
	var itemValues DynamicParams
	params = make([]any, len(static)+len(itemValues)*len(data))
	placeholders := make([]string, len(data))
	for i := 0; i < len(static); i++ {
		params[i] = static[i]
	}
	fmtParams := make([]any, len(itemValues))
	for i, item := range data {
		baseIndex := len(static) + len(itemValues)*i
		itemValues = item.GetMassInsertValues()
		for j := 0; j < len(itemValues); j++ {
			params[baseIndex+j] = itemValues[j]
			fmtParams[j] = baseIndex + j + 1
		}
		placeholders[i] = fmt.Sprintf(mib.placeholderTemplate, fmtParams...)
	}
	query = fmt.Sprintf(mib.queryTemplate, strings.Join(placeholders, ", "))
	return
}
