package w3req

import (
	"errors"
	"fmt"
	"sync"

	"github.com/algebrain/w3/w3sql"
)

type Logger interface {
	LogSQL(string, string, map[string]any)
	Logf(string, string, ...any)
}

type Conn interface {
	Select(any, string, ...any) ([]any, error)
	SelectInt(string, ...any) (int64, error)
}

type TotalGetter[T any] interface {
	Total(T) (int64, error)
}

type SelectConfig[T any] struct {
	FieldMap   map[string]string
	LowerCols  []string
	AllSQL     *w3sql.SQLString
	TotalSQL   *w3sql.SQLString
	SQLDialect string

	DumpRequests bool
	AutoTotal    bool

	TotalGetter TotalGetter[T]
	OnPanic     func()
}

type SelectOptions[T any] struct {
	And    *w3sql.Query
	Or     *w3sql.Query
	Logger Logger
	Conn   func() Conn
}

type SelectRequester[T any] interface {
	InitOnce(f func() *SelectOptions[T])
	Handle(q *w3sql.Query) ([]T, int64, error)
	SetDumpRequests(v bool)
}

type selectRequester[T any] struct {
	cfg       *SelectConfig[T]
	opt       *SelectOptions[T]
	lowerCols map[string]bool
	mut       sync.Mutex
	initOnce  sync.Once
	conn      Conn
}

func NewSelectRequester[T any](cfg *SelectConfig[T]) (SelectRequester[T], error) {
	if cfg.OnPanic == nil {
		return nil, errors.New("[w3req.SelectRequester.NewSelectRequester] OnPanic is mandatory")
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
	defer r.cfg.OnPanic()
	r.initOnce.Do(func() {
		if r.opt != nil {
			return
		}
		opt := f()
		if opt.Conn == nil {
			panic("[w3req.SelectRequester.NewSelectRequester] Conn is mandatory")
		}
		r.opt = opt
	})
}

func (r *selectRequester[T]) Handle(q *w3sql.Query) ([]T, int64, error) {
	defer r.cfg.OnPanic()

	if r.opt.And != nil {
		if r.opt.And.Search != nil {
			q.Search = w3sql.And(q.Search, r.opt.And.Search)
		}
		if r.opt.And.Sort != nil {
			q.Sort = append(q.Sort, r.opt.And.Sort...)
		}
	} else if r.opt.Or != nil {
		if r.opt.Or.Search != nil {
			q.Search = w3sql.Or(q.Search, r.opt.Or.Search)
		}
		if r.opt.Or.Sort != nil {
			q.Sort = append(q.Sort, r.opt.Or.Sort...)
		}
	}

	q.LowerSearchValues(r.lowerCols)

	sq, err := q.CompileSelect(r.cfg.SQLDialect, r.cfg.FieldMap)
	if err != nil {
		return nil, 0, err
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
			return nil, 0, err
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
			return nil, 0, err
		}

		if total == 0 {
			return []T{}, 0, nil
		}
	}

	t, err := sq.SQL(r.cfg.AllSQL)
	if err != nil {
		return nil, 0, err
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
		return nil, 0, err
	}

	//динамически
	if r.cfg.TotalSQL == nil && ret != nil && len(ret) > 0 {
		if r.cfg.TotalGetter != nil {
			total, err = r.cfg.TotalGetter.Total(ret[0])
			if err != nil {
				return nil, 0, err
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

	return ret, total, nil
}

func (r *selectRequester[T]) SetDumpRequests(v bool) {
	r.cfg.DumpRequests = v
}
