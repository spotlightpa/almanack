// Package paginate does the math to over-request pages by one
package paginate

import (
	"context"
	"fmt"

	"github.com/carlmjohnson/resperr"
)

type Paginator[I int | int32 | int64] struct {
	PageNum, PageSize, NextPage I
}

func PageNumber[I int | int32 | int64](pageNum I) *Paginator[I] {
	return &Paginator[I]{
		PageNum: pageNum,
	}
}

func (pg Paginator[I]) Offset() I {
	return pg.PageNum * pg.PageSize
}

func (pg Paginator[I]) Limit() I {
	return pg.PageSize + 1
}

func (pg Paginator[I]) HasMore() bool {
	return pg.NextPage > 0
}

type CtxFunc[Param, Result any] func(context.Context, Param) ([]Result, error)

func List[Param, Result any, I int | int32 | int64](
	pg *Paginator[I], ctx context.Context, fn CtxFunc[Param, Result], param Param,
) (results []Result, err error) {
	if pg.PageSize < 1 {
		panic(fmt.Sprint("bad pagination size", pg.PageSize))
	}
	if pg.PageNum < 0 {
		return nil, resperr.WithUserMessage(nil, "Invalid page number.")
	}

	results, err = fn(ctx, param)
	if err != nil {
		return nil, err
	}
	if len(results) > int(pg.PageSize) {
		pg.NextPage = pg.PageNum + 1
		results = results[:pg.PageSize]
	}
	return results, nil
}
