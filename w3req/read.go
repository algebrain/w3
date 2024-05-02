package w3req

import (
	"errors"
	"fmt"
	"github.com/algebrain/w3/w3sql"
	"sync"
)

type Logger interface {
	LogSQL(string, string, map[string]any)
	Logf(string, ...any)
}

type Conn interface {
	Select(any, string, map[string]any) (any, error)
	SelectInt(string, map[string]any) (int64, error)
}

type TotalGetter[T any] interface {
	Total(T) (int64, error)
}

type SelectConfig[T any] struct {
	Limit      int
	FieldMap   map[string]string
	LowerCols  []string
	AllSQL     *w3sql.SQLString
	TotalSQL   *w3sql.SQLString
	SQLDialect string

	DumpRequests bool
	AutoTotal    bool

	TotalGetter TotalGetter[T]
	onPanic     func()
}

type SelectOptions[T any] struct {
	Logger    Logger
	Conn      func() Conn
	Append    *w3sql.Query
	onSuccess func([]T, int64)
	onError   func(error)
}

type SelectRequester[T any] interface {
	InitOnce(f func() *SelectOptions[T])
	Handle(w3sql.Query)
}

type selectRequester[T any] struct {
	cfg       *SelectConfig[T]
	opt       *SelectOptions[T]
	lowerCols map[string]bool
	mut       sync.Mutex
	conn      Conn
}

func NewSelectRequester[T any](cfg *SelectConfig[T]) (SelectRequester[T], error) {
	if cfg.onPanic == nil {
		return nil, errors.New("onPanic is mandatory")
	}
	lowerCols := map[string]bool{}
	for _, c := range cfg.LowerCols {
		lowerCols[c] = true
	}
	return &selectRequester[T]{
		cfg:       cfg,
		lowerCols: lowerCols,
		mut:       sync.Mutex{},
	}, nil
}

func (r *selectRequester[T]) InitOnce(f func() *SelectOptions[T]) {
	if r.opt != nil {
		return
	}
	opt := f()
	if opt.Conn == nil {
		panic("[w3req.SelectRequester.InitOnce] Conn is mandatory")
	}
	if opt.onError == nil {
		panic("[w3req.SelectRequester.InitOnce] onError is mandatory")
	}
	if opt.onSuccess == nil {
		panic("[w3req.SelectRequester.InitOnce] onSuccess is mandatory")
	}
	r.opt = opt
}

func (r *selectRequester[T]) Handle(q w3sql.Query) {
	defer r.cfg.onPanic()

	if q.Limit == nil || *q.Limit > r.cfg.Limit || *q.Limit == 0 {
		q.Limit = &r.cfg.Limit
	}

	if r.opt.Append != nil {
		if r.opt.Append.Search != nil {
			q.Search = w3sql.And(q.Search, r.opt.Append.Search)
		}
		if r.opt.Append.Sort != nil {
			q.Sort = append(q.Sort, r.opt.Append.Sort...)
		}
	}

	q.LowerSearchValues(r.lowerCols)

	sq, err := q.CompileSelect(r.cfg.SQLDialect, r.cfg.FieldMap)
	if err != nil {
		r.opt.onError(err)
		return
	}

	if sq == nil {
		panic("[w3req.SelectRequester.Handle]: no query")
	}

	func() {
		r.mut.Lock()
		defer r.mut.Unlock()
		if r.conn == nil {
			r.conn = r.opt.Conn()
		}
	}()

	if r.conn == nil {
		panic("[w3req.SelectRequester.Handle]: Conn is nil")
	}

	var total int64

	if r.cfg.TotalSQL != nil {
		t, err := sq.NoLimitOffset().SQL(r.cfg.TotalSQL)
		if err != nil {
			r.opt.onError(err)
			return
		}

		if r.cfg.DumpRequests && r.opt.Logger != nil {
			r.opt.Logger.LogSQL("Total SQL:", t[0].Code, t[0].Params)
		}

		total, err = r.conn.SelectInt(t[0].Code, t[0].Params)
		if err != nil {
			err = fmt.Errorf(
				"SelectOne error: %s\nSQL: %s\nData:%+v\n",
				err.Error(),
				t[0].Code, t[0].Params,
			)
			r.opt.onError(err)
			return
		}

		if total == 0 {
			return
		}
	}

	t, err := sq.SQL(r.cfg.AllSQL)
	if err != nil {
		r.opt.onError(err)
		return
	}

	if r.cfg.DumpRequests && r.opt.Logger != nil {
		r.opt.Logger.LogSQL("Data SQL:", t[0].Code, t[0].Params)
	}

	var ret []T
	_, err = r.conn.Select(&ret, t[0].Code, t[0].Params)
	if err != nil {
		err = fmt.Errorf(
			"Select error: %s\nSQL: %s\nData:%+v\n",
			err.Error(),
			t[0].Code, t[0].Params,
		)
		r.opt.onError(err)
		return
	}

	//динамически
	if r.cfg.TotalSQL == nil && ret != nil && len(ret) > 0 {
		if r.cfg.TotalGetter != nil {
			total, err = r.cfg.TotalGetter.Total(ret[0])
			if err != nil {
				r.opt.onError(err)
				return
			}
		} else if r.cfg.AutoTotal {
			total = int64(len(ret))
			limit := *q.Limit
			if total >= int64(limit) && limit > 0 {
				total = total + 1 //101 чтобы работало листание
			}
			total = total + int64(*q.Offset) //= 100+97, исправление последней страницы
		}
	}

	r.opt.onSuccess(ret, total)
}
