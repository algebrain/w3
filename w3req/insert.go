package w3req

import (
	"errors"
	"fmt"
	"sync"

	"github.com/algebrain/w3/w3sql"
)

type InsertConfig struct {
	AllSQL       *w3sql.SQLString
	FieldMap     map[string]string
	SQLDialect   string
	DumpRequests bool
	OnPanic      func()
}

type InsertOptions struct {
	Logger    Logger
	DB        func() DB
	Transform w3sql.ValueTransform
}

type InsertRequester interface {
	InitOnce(f func() *InsertOptions)
	Handle(q *w3sql.Query) error
	SetDumpRequests(v bool)
}

type insertRequester struct {
	cfg      *InsertConfig
	opt      *InsertOptions
	mut      sync.Mutex
	initOnce sync.Once
	conn     DB
}

func NewInsertRequester(cfg *InsertConfig) (InsertRequester, error) {
	if cfg.OnPanic == nil {
		return nil, errors.New("[w3req.InsertRequester.NewInsertRequester] OnPanic is mandatory")
	}
	return &insertRequester{
		cfg: cfg,
		mut: sync.Mutex{},
	}, nil
}

func (r *insertRequester) InitOnce(f func() *InsertOptions) {
	defer r.cfg.OnPanic()
	r.initOnce.Do(func() {
		if r.opt != nil {
			return
		}
		opt := f()
		if opt.DB == nil {
			panic("[w3req.InsertRequester.NewInsertRequester] DB is mandatory")
		}
		r.opt = opt
	})
}

func (r *insertRequester) Handle(q *w3sql.Query) error {
	defer r.cfg.OnPanic()

	var tr []w3sql.ValueTransform
	if r.opt.Transform != nil {
		tr = []w3sql.ValueTransform{r.opt.Transform}
	}
	sq, err := q.CompileInsert(r.cfg.SQLDialect, r.cfg.FieldMap, tr...)
	if err != nil {
		return err
	}

	if sq == nil {
		panic("[w3req.InsertRequester.Handle]: no query")
	}

	func() {
		r.mut.Lock()
		defer r.mut.Unlock()
		if r.conn == nil {
			r.conn = r.opt.DB()
		}
	}()

	if r.conn == nil {
		panic("[w3req.InsertRequester.Handle]: DB is nil")
	}

	t, err := sq.SQL(r.cfg.AllSQL)
	if err != nil {
		return err
	}

	if r.cfg.DumpRequests && r.opt.Logger != nil {
		r.opt.Logger.LogSQL("Insert SQL:", t[0].Code, t[0].Params)
	}

	_, err = r.conn.Exec(t[0].Code, t[0].Params)
	if err != nil {
		err = fmt.Errorf(
			"Insert error: %s\nSQL: %s\nParams:%+v\n",
			err.Error(),
			t[0].Code, t[0].Params,
		)
		return err
	}

	return nil
}

func (r *insertRequester) SetDumpRequests(v bool) {
	r.cfg.DumpRequests = v
}
