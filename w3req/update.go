package w3req

import (
	"errors"
	"fmt"
	"sync"

	"github.com/algebrain/w3/w3sql"
)

type UpdateConfig struct {
	AllSQL       *w3sql.SQLString
	IDFieldName  string
	FieldMap     map[string]string
	SQLDialect   string
	DumpRequests bool
	OnPanic      func()
}

type UpdateOptions struct {
	Logger    Logger
	DB        func() DB
	Transform w3sql.ValueTransform
}

type UpdateRequester interface {
	InitOnce(f func() *UpdateOptions)
	Handle(q *w3sql.Query) error
	SetDumpRequests(v bool)
}

type updateRequester struct {
	cfg      *UpdateConfig
	opt      *UpdateOptions
	mut      sync.Mutex
	initOnce sync.Once
	conn     DB
}

func NewUpdateRequester(cfg *UpdateConfig) (UpdateRequester, error) {
	if cfg.OnPanic == nil {
		return nil, errors.New("[w3req.UpdateRequester.NewUpdateRequester] OnPanic is mandatory")
	}
	return &updateRequester{
		cfg: cfg,
		mut: sync.Mutex{},
	}, nil
}

func (r *updateRequester) InitOnce(f func() *UpdateOptions) {
	defer r.cfg.OnPanic()
	r.initOnce.Do(func() {
		if r.opt != nil {
			return
		}
		opt := f()
		if opt.DB == nil {
			panic("[w3req.UpdateRequester.NewUpdateRequester] DB is mandatory")
		}
		r.opt = opt
	})
}

func (r *updateRequester) Handle(q *w3sql.Query) error {
	defer r.cfg.OnPanic()

	var tr []w3sql.ValueTransform
	if r.opt.Transform != nil {
		tr = []w3sql.ValueTransform{r.opt.Transform}
	}
	sq, err := q.CompileUpdate(r.cfg.SQLDialect, r.cfg.FieldMap, r.cfg.IDFieldName, tr...)
	if err != nil {
		return err
	}

	if sq == nil {
		panic("[w3req.UpdateRequester.Handle]: no query")
	}

	func() {
		r.mut.Lock()
		defer r.mut.Unlock()
		if r.conn == nil {
			r.conn = r.opt.DB()
		}
	}()

	if r.conn == nil {
		panic("[w3req.UpdateRequester.Handle]: DB is nil")
	}

	t, err := sq.SQL(r.cfg.AllSQL)
	if err != nil {
		return err
	}

	if r.cfg.DumpRequests && r.opt.Logger != nil {
		r.opt.Logger.LogSQL("Update SQL:", t[0].Code, t[0].Params)
	}

	_, err = r.conn.Exec(t[0].Code, t[0].Params)
	if err != nil {
		err = fmt.Errorf(
			"Update error: %s\nSQL: %s\nParams:%+v\n",
			err.Error(),
			t[0].Code, t[0].Params,
		)
		return err
	}

	return nil
}

func (r *updateRequester) SetDumpRequests(v bool) {
	r.cfg.DumpRequests = v
}
