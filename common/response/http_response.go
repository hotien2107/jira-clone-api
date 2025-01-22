package response

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"jira-clone-api/common/request"
	"jira-clone-api/utilities/local"
)

type Options struct {
	Extra      fiber.Map
	Data       interface{}
	Code       int
	ReturnCode int
}

func (opt *Options) addDefaultValue() {
	if opt.Code < fiber.StatusContinue {
		opt.Code = fiber.StatusOK
	}
	if opt.Data == nil {
		opt.Data = fiber.Map{}
	}
}

type Error struct {
	Data       interface{} `json:"message"`
	Code       int         `json:"code"`
	ReturnCode int         `json:"return_code"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v", e.Data)
}

type ErrorOptions struct {
	Data       interface{}
	ReturnCode int
}

type Service struct{}

func NewError(code int, opts ...ErrorOptions) *Error {
	e := &Error{Code: code}
	if len(opts) > 0 {
		opt := opts[0]
		e.Data = opt.Data
		e.ReturnCode = opt.ReturnCode
	} else {
		e.Data = utils.StatusMessage(code)
		e.ReturnCode = code
	}
	return e
}

func New(ctx *fiber.Ctx, opt Options) (err error) {
	opt.addDefaultValue()
	localStorage := local.New(ctx)
	localStorage.SetStatusCode(opt.Code)
	if fiber.StatusOK <= opt.Code && opt.Code < fiber.StatusMultipleChoices {
		if opt.Extra != nil {
			return ctx.Status(opt.Code).JSON(fiber.Map{
				"return_code": opt.ReturnCode,
				"status_code": opt.Code,
				"data":        opt.Data,
				"extra":       opt.Extra,
			})
		}
		return ctx.Status(opt.Code).JSON(fiber.Map{
			"return_code": opt.ReturnCode,
			"status_code": opt.Code,
			"data":        opt.Data,
		})
	}
	return ctx.Status(opt.Code).JSON(fiber.Map{
		"return_code": opt.ReturnCode,
		"status_code": opt.Code,
		"error":       opt.Data,
	})
}

func NewArrayWithPaginationSuccess(ctx *fiber.Ctx, data interface{}, pagination *request.Pagination) error {
	if pagination.Total != nil {
		return New(ctx, Options{Data: data, Extra: fiber.Map{"limit": pagination.Limit, "page": pagination.Page, "total": *pagination.Total}})
	}
	return New(ctx, Options{Data: data, Extra: fiber.Map{"limit": pagination.Limit, "page": pagination.Page}})
}

func NewArrayWithPaginationFailure(ctx *fiber.Ctx, pagination *request.Pagination) error {
	return New(ctx, Options{Data: []fiber.Map{}, Extra: fiber.Map{"limit": pagination.Limit, "page": pagination.Page, "total": pagination.Total}})
}

func NewArrayWithPagination(ctx *fiber.Ctx, data interface{}, pagination *request.Pagination) (err error) {
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		if !reflect.ValueOf(data).IsNil() {
			return NewArrayWithPaginationSuccess(ctx, data, pagination)
		}
		fallthrough
	default:
		return NewArrayWithPaginationFailure(ctx, pagination)
	}
}

func FiberErrorHandler(ctx *fiber.Ctx, err error) error {
	if e := new(Error); errors.As(err, &e) {
		return New(ctx, Options{Code: e.Code, Data: e.Data, ReturnCode: e.ReturnCode})
	} else if e := new(fiber.Error); errors.As(err, &e) {
		return New(ctx, Options{Code: e.Code, Data: e.Message})
	} else {
		return New(ctx, Options{Code: fiber.StatusInternalServerError, Data: err.Error()})
	}
}
