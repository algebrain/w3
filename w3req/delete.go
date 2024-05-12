package w3req

import (
	"errors"
	"fmt"
	"sync"

	"github.com/algebrain/w3/w3sql"
)

type DeleteConfig struct {
	AllSQL       *w3sql.SQLString
	Tables       []*w3sql.DeletePair
	SQLDialect   string
	DumpRequests bool
	OnPanic      func()
}

type DeleteOptions struct {
	Logger    Logger
	DB        func() DB
	Transform w3sql.DeleteTransform
}

type DeleteRequester interface {
	InitOnce(f func() *DeleteOptions)
	Handle(q *w3sql.Query) error
	SetDumpRequests(v bool)
}

type deleteRequester struct {
	cfg      *DeleteConfig
	opt      *DeleteOptions
	mut      sync.Mutex
	initOnce sync.Once
	conn     DB
}

func NewDeleteRequester(cfg *DeleteConfig) (DeleteRequester, error) {
	if cfg.OnPanic == nil {
		return nil, errors.New("[w3req.DeleteRequester.NewDeleteRequester] OnPanic is mandatory")
	}
	return &deleteRequester{
		cfg: cfg,
		mut: sync.Mutex{},
	}, nil
}

func (r *deleteRequester) InitOnce(f func() *DeleteOptions) {
	defer r.cfg.OnPanic()
	r.initOnce.Do(func() {
		if r.opt != nil {
			return
		}
		opt := f()
		if opt.DB == nil {
			panic("[w3req.DeleteRequester.NewDeleteRequester] DB is mandatory")
		}
		r.opt = opt
	})
}

func (r *deleteRequester) Handle(q *w3sql.Query) error {
	defer r.cfg.OnPanic()

	var tr []w3sql.DeleteTransform
	if r.opt.Transform != nil {
		tr = []w3sql.DeleteTransform{r.opt.Transform}
	}
	sq, err := q.CompileDelete(r.cfg.SQLDialect, r.cfg.Tables, tr...)
	if err != nil {
		return err
	}

	if sq == nil {
		panic("[w3req.DeleteRequester.Handle]: no query")
	}

	func() {
		r.mut.Lock()
		defer r.mut.Unlock()
		if r.conn == nil {
			r.conn = r.opt.DB()
		}
	}()

	if r.conn == nil {
		panic("[w3req.DeleteRequester.Handle]: DB is nil")
	}

	t, err := sq.SQL(r.cfg.AllSQL)
	if err != nil {
		return err
	}

	if r.cfg.DumpRequests && r.opt.Logger != nil {
		r.opt.Logger.LogSQL("Delete SQL:", t[0].Code, t[0].Params)
	}

	for _, tt := range t {
		_, err = r.conn.Exec(tt.Code, tt.Params)
		if err != nil {
			err = fmt.Errorf(
				"Delete error: %s\nSQL: %s\nParams:%+v\n",
				err.Error(),
				tt.Code, tt.Params,
			)
			return err
		}
	}

	return nil
}

func (r *deleteRequester) SetDumpRequests(v bool) {
	r.cfg.DumpRequests = v
}
