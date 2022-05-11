package db

import (
	"context"
	"fmt"
)

type Paginator[I int | int32 | int64] struct {
	PageNum, PageSize, NextPage I
}

func PageNumSize[I int | int32 | int64](pageNum, pageSize I) *Paginator[I] {
	if pageNum < 0 || pageSize < 1 {
		panic(fmt.Sprint("bad pagination options", pageNum, pageSize))
	}
	return &Paginator[I]{
		PageNum:  pageNum,
		PageSize: pageSize,
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

func Paginate[Param, Result any, I int | int32 | int64](
	pg *Paginator[I], ctx context.Context, fn CtxFunc[Param, Result], param Param,
) (results []Result, err error) {
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
