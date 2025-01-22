package queries

import (
	"regexp"
	"strings"

	"jira-clone-api/common/configure"
	"jira-clone-api/common/logging"
	"jira-clone-api/common/request"
	"jira-clone-api/common/response"
	"jira-clone-api/database/mongo"
	"jira-clone-api/utilities/tool"

	respErr "jira-clone-api/common/response/error"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	logger      = logging.GetLogger()
	cfg         = configure.GetConfig()
	timeoutFunc = mongo.NewUtilityService().GetContextTimeout
)

const (
	QueryMethodEqual = iota
	QueryMethodNotEqual
	QueryMethodContains
	QueryMethodStartsWith
	QueryMethodEndsWith
	QueryMethodGreaterThan
	QueryMethodGreaterThanOrEqual
	QueryMethodLessThan
	QueryMethodLessThanOrEqual
	QueryMethodIn
	QueryMethodNotIn
)

const (
	SortTypeDesc = -1
	SortTypeAsc  = 1
)

var queryMethodMap = map[int]string{
	QueryMethodEqual:              "$eq",
	QueryMethodNotEqual:           "$ne",
	QueryMethodContains:           "$regex",
	QueryMethodStartsWith:         "$regex",
	QueryMethodEndsWith:           "$regex",
	QueryMethodGreaterThan:        "$gt",
	QueryMethodGreaterThanOrEqual: "$gte",
	QueryMethodLessThan:           "$lt",
	QueryMethodLessThanOrEqual:    "$lte",
	QueryMethodIn:                 "$in",
	QueryMethodNotIn:              "$nin",
}

const (
	ViewByDay   = "day"
	ViewByWeek  = "week"
	ViewByMonth = "month"
	ViewByYear  = "year"
)

type OptionsQuery interface {
	SetOnlyFields(fieldNames ...string)
	SetPagination(pagination *request.Pagination)
	QueryOnlyField() interface{}
	QueryPaginationLimit() *int64
	QueryPaginationPage() *int64
	QueryPaginationSkip() *int64
	QuerySort() bson.D
	ResetSort()
	AddSortKey(map[string]int)
}

type optionsQuery struct {
	pagination *request.Pagination
	sort       bson.D
	onlyFields []string
}

func NewOptions() OptionsQuery {
	return &optionsQuery{
		sort: make(bson.D, 0),
	}
}

func (o *optionsQuery) SetOnlyFields(fieldNames ...string) {
	o.onlyFields = fieldNames
}

func (o *optionsQuery) SetPagination(pagination *request.Pagination) {
	o.pagination = pagination
}

func (o *optionsQuery) QueryOnlyField() interface{} {
	if len(o.onlyFields) < 1 {
		return nil
	}
	result := make(bson.M)
	for _, fieldName := range o.onlyFields {
		result[fieldName] = 1
	}
	return result
}

func (o *optionsQuery) QueryPaginationLimit() *int64 {
	if o.pagination == nil {
		return nil
	}
	return &o.pagination.Limit
}

func (o *optionsQuery) QueryPaginationPage() *int64 {
	if o.pagination == nil {
		return nil
	}
	return &o.pagination.Page
}

func (o *optionsQuery) QueryPaginationSkip() *int64 {
	if o.pagination == nil {
		return nil
	}
	return &o.pagination.Skip
}

func (o *optionsQuery) AddSortKey(sorts map[string]int) {
	for sortBy, sortType := range sorts {
		if sortType != SortTypeAsc && sortType != SortTypeDesc {
			sortType = SortTypeDesc
		}
		o.sort = append(o.sort, bson.E{Key: sortBy, Value: sortType})
	}
}

func (o *optionsQuery) ResetSort() {
	o.sort = make(bson.D, 0)
}

func (o *optionsQuery) QuerySort() bson.D {
	return o.sort
}

type Filter struct {
	Value  interface{}
	By     string
	Method int
}

type optionsFilter struct {
	filters []Filter
}
type OptionsFilter interface {
	AddListFilter(filters []Filter) error
	BuildMongoFilterWithAndCondition() bson.M
	BuildMongoFilterWithOrCondition() bson.M
}

func NewOptionsFilter() OptionsFilter {
	return &optionsFilter{}
}

func (f *optionsFilter) AddListFilter(filters []Filter) error {
	for _, filter := range filters {
		if _, ok := queryMethodMap[filter.Method]; !ok {
			return response.NewError(fiber.StatusBadRequest, response.ErrorOptions{Data: respErr.ErrQueryMethodNotAllowed})
		}
		f.filters = append(f.filters, filter)
	}
	return nil
}

func (f *optionsFilter) BuildMongoFilterWithAndCondition() bson.M {
	if len(f.filters) == 0 {
		return make(bson.M)
	}
	condition := make([]bson.M, 0)
	for _, v := range f.filters {
		filter := make(bson.M)
		switch v.Method {
		case QueryMethodContains:
			filter[v.By] = bson.M{queryMethodMap[v.Method]: primitive.Regex{Pattern: f.buildVietnameseRegex(regexp.QuoteMeta(v.Value.(string))), Options: "i"}}
		case QueryMethodStartsWith:
			filter[v.By] = bson.M{queryMethodMap[v.Method]: primitive.Regex{Pattern: "^" + f.buildVietnameseRegex(v.Value.(string)), Options: "i"}}
		case QueryMethodEndsWith:
			filter[v.By] = bson.M{queryMethodMap[v.Method]: primitive.Regex{Pattern: f.buildVietnameseRegex(v.Value.(string)) + "$", Options: "i"}}
		default:
			id, ok := v.Value.(primitive.ObjectID)
			if ok {
				filter[v.By] = bson.M{queryMethodMap[v.Method]: id}
			} else {
				filter[v.By] = bson.M{queryMethodMap[v.Method]: v.Value}
			}
		}
		condition = append(condition, filter)
	}
	return bson.M{"$and": condition}
}

func (f *optionsFilter) BuildMongoFilterWithOrCondition() bson.M {
	if len(f.filters) == 0 {
		return make(bson.M)
	}
	condition := make([]bson.M, 0)
	for _, v := range f.filters {
		filter := make(bson.M)
		switch v.Method {
		case QueryMethodContains:
			filter[v.By] = bson.M{queryMethodMap[v.Method]: primitive.Regex{Pattern: f.buildVietnameseRegex(regexp.QuoteMeta(v.Value.(string))), Options: "i"}}
		case QueryMethodStartsWith:
			filter[v.By] = bson.M{queryMethodMap[v.Method]: primitive.Regex{Pattern: "^" + f.buildVietnameseRegex(v.Value.(string)), Options: "i"}}
		case QueryMethodEndsWith:
			filter[v.By] = bson.M{queryMethodMap[v.Method]: primitive.Regex{Pattern: f.buildVietnameseRegex(v.Value.(string)) + "$", Options: "i"}}
		default:
			id, ok := v.Value.(primitive.ObjectID)
			if ok {
				filter[v.By] = bson.M{queryMethodMap[v.Method]: id}
			} else {
				filter[v.By] = bson.M{queryMethodMap[v.Method]: v.Value}
			}
		}
		condition = append(condition, filter)
	}
	return bson.M{"$or": condition}
}

func (f *optionsFilter) ResetFilter() {
	f.filters = nil
}

func (f *optionsFilter) buildVietnameseRegex(value string) string {
	value = strings.ToLower(value)
	value = tool.New().DeaccentVietnameseString(value)
	replacer := strings.NewReplacer(
		"a", "[aàáãảạăằắẳẵặâầấẩẫậ]",
		"e", "[eèéẻẽẹêềếểễệ]",
		"u", "[uùúủũụưừứửữự]",
		"d", "[dđ]",
		"o", "[oòóỏõọôồốổỗộơờớởỡợ]",
		"i", "[iìíỉĩị]",
		"y", "[yýỳỹỵỷ]")
	return replacer.Replace(value)
}
