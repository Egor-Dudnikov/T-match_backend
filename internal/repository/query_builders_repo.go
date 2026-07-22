package repository

import (
	"strconv"
	"strings"
)

type querySelectBuilder struct {
	index        int
	query        strings.Builder
	existWhere   bool
	existOrderBy bool
	values       []interface{}
}

func newQuerySelectBuilder(selectQuery string) *querySelectBuilder {
	var selectBuilder strings.Builder
	selectBuilder.WriteString(selectQuery)
	qb := &querySelectBuilder{
		query: selectBuilder,
		index: 1,
	}
	return qb
}

func (qb *querySelectBuilder) addWhereWithIndex(preIndex string, postIndex string, value any) {
	if value == nil {
		return
	}

	if qb.existWhere {
		qb.query.WriteString(" AND ")
	} else {
		qb.query.WriteString(" WHERE ")
		qb.existWhere = true
	}

	qb.query.WriteString(preIndex)
	qb.query.WriteString(strconv.Itoa(qb.index))
	qb.index++
	qb.query.WriteString(postIndex)
	qb.values = append(qb.values, value)
}

func (qb *querySelectBuilder) addWhere(where string) {
	if qb.existWhere {
		qb.query.WriteString(" AND ")
	} else {
		qb.query.WriteString(" WHERE ")
		qb.existWhere = true
	}
	qb.query.WriteString(where)
}

func (qb *querySelectBuilder) addOrderBy(sort string, orderBy string) {
	if qb.existOrderBy {
		qb.query.WriteString(", ")
	} else {
		qb.query.WriteString(" ORDER BY ")
		qb.existOrderBy = true
	}
	qb.query.WriteString(sort + " " + orderBy)
}

func (qb *querySelectBuilder) addLimit(limit *int) {
	if limit == nil {
		return
	}
	qb.query.WriteString(" LIMIT $")
	qb.query.WriteString(strconv.Itoa(qb.index))
	qb.index++
	qb.values = append(qb.values, *limit)
}

func (qb *querySelectBuilder) addOffset(offset *int) {
	if offset == nil {
		return
	}
	qb.query.WriteString(" OFFSET $")
	qb.query.WriteString(strconv.Itoa(qb.index))
	qb.index++
	qb.values = append(qb.values, *offset)
}

func (qb *querySelectBuilder) parseBuilder() (string, []interface{}) {
	return qb.query.String(), qb.values
}

type queryUpdateBuilder struct {
	index       int
	existUpdate bool
	query       strings.Builder
	values      []interface{}
}

func newUpdateQuery(updateQuery string, id int) *queryUpdateBuilder {
	var selectBuilder strings.Builder
	selectBuilder.WriteString(updateQuery)
	qb := &queryUpdateBuilder{
		query: selectBuilder,
		index: 2,
	}
	qb.values = append(qb.values, id)
	return qb
}

func (qb *queryUpdateBuilder) addFilled(name string, value any) {
	if value == nil {
		return
	}

	if qb.existUpdate {
		qb.query.WriteString(", ")
	}
	qb.existUpdate = true

	qb.query.WriteString(name)
	qb.query.WriteString(" = $")
	qb.query.WriteString(strconv.Itoa(qb.index))

	qb.values = append(qb.values, value)
	qb.index++
}

func (qb *queryUpdateBuilder) parseBuilder(where string) (string, []interface{}) {
	qb.query.WriteString(" " + where)
	return qb.query.String(), qb.values
}
