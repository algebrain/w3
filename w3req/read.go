package w3req

import (
	"errors"
	"github.com/algebrain/w3/w3sql"
)

type Logger interface {
	LogError(error)
	LogSQL(*w3sql.SQLQuery)
	Logf(string, ...any)
}

type Conn interface {
	Select(any, string, map[string]any) (any, error)
}

type TotalGetter[T any] interface {
	Total(T) (int64, error)
}

type SelectConfig[T any] struct {
	Limit     int
	FieldMap  map[string]string
	LowerCols []string
	AllSQL    string
	TotalSQL  string
	AutoTotal bool

	TotalGetter  TotalGetter[T]
	PanicHandler func()
}

type SelectOptions struct {
	Logger Logger
	Conn   Conn
}

type SelectRequester interface {
	SetOptions(SelectOptions) error
	Handle(w3sql.Query)
}

type selectRequester[T any] struct {
	cfg SelectConfig[T]
	opt SelectOptions
}

func NewSelectRequester[T any](cfg SelectConfig[T]) (SelectRequester, error) {
	if cfg.PanicHandler == nil {
		return nil, errors.New("PanicHandler is mandatory")
	}
	return &selectRequester[T]{
		cfg: cfg,
	}, nil
}

func (r *selectRequester[T]) SetOptions(opt SelectOptions) error {
	if opt.Conn == nil {
		return errors.New("Conn is mandatory")
	}
	r.opt = opt
	return nil
}

func (r *selectRequester[T]) Handle(q w3sql.Query) {
	defer r.cfg.PanicHandler()
	// TODO:
}
