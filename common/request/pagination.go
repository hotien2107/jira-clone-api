package request

import (
	"github.com/gofiber/fiber/v2"
	"jira-clone-api/common/configure"
)

var (
	cfg = configure.GetConfig()
)

// Pagination contains pagination request
type Pagination struct {
	Total *int64 `json:"-"`
	Limit int64  `form:"limit" json:"limit"`
	Page  int64  `form:"page" json:"page"`
	Skip  int64  `json:"-"`
}

func (p *Pagination) SetTotal(value int64) {
	p.Total = &value
}

func (p *Pagination) Format() {
	var maxLimit = cfg.PaginationMaxItem
	if p.Limit > maxLimit {
		p.Limit = maxLimit
	} else if p.Limit < 1 {
		p.Limit = cfg.PaginationMaxItem
	}
	if p.Page < 1 {
		p.Page = 1
	}
	p.Skip = p.Limit * (p.Page - 1)
	p.SetTotal(0)
}

func NewPagination(limit, page int64) *Pagination {
	p := Pagination{Limit: limit, Page: page}
	p.Format()
	return &p
}

func NewPaginationWithCtx(ctx *fiber.Ctx) (*Pagination, error) {
	var p Pagination
	if err := ctx.BodyParser(&p); err != nil {
		return nil, err
	}
	p.Format()
	return &p, nil
}
