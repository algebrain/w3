package w3ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"

	"github.com/algebrain/w3/w3req"
	"github.com/algebrain/w3/w3sql"
	"github.com/valyala/fasthttp"
)

const (
	SYSTEM_ERROR       = "System error"
	INVALID_PARAMETERS = "Invalid parameters"
)

type ExtLogger interface {
	Print(string)
	Printf(string, any, ...any)
}

type Logger struct {
	errorLog            ExtLogger
	debugLog            ExtLogger
	outputOriginalError bool
}

type DataRequester[T any] struct {
	sel          w3req.SelectRequester[T]
	formatFields func([]T)
	logger       *Logger
	errorCodes   ErrorCodes
	onPanic      func()
}

func (log *Logger) setErrorLogger(z ExtLogger) {
	log.errorLog = z
}

func (log *Logger) setDebugLogger(z ExtLogger) {
	log.debugLog = z
}

func (log *Logger) LogSQL(prefix string, sql string, params map[string]any) {
	if log.debugLog != nil {
		log.debugLog.Printf("%s\n<<%s>>\nParameters:\n%+v\n", prefix, sql, getJSON(params))
	}
}

func (log *Logger) Logf(format string, arg string, args ...any) {
	if log.debugLog != nil {
		log.debugLog.Printf(format, arg, args...)
	}
}

func (log *Logger) LogError(extError string, err error, errout func(string)) {
	if runtime.GOOS == "windows" {
		fmt.Printf("%s: %s\n", extError, err.Error())
		debug.PrintStack()
	}

	if log.errorLog != nil {
		log.errorLog.Print(extError + ": " + err.Error())
	}

	if errout == nil {
		panic("[w3ui.requester.LogError] error: errout is nil")
		return
	}

	if log.outputOriginalError {
		errout(extError + ": " + err.Error())
	} else {
		errout(extError)
	}
}

func NewDataRequester3[T any](
	allSQL *w3sql.SQLString, //запрос
	compileMap map[string]string, //карта соответствия фронт аргумент -> sql
	lowerEm []string, //значения поискового запроса фронта будут to_lower
	errorCodes map[string]int, //коды ошибок
	onPanic func(),
) *DataRequester[T] {
	if onPanic == nil {
		panic("[w3ui.NewDataRequester] ERROR: onPanic should not be nil")
	}

	opt := w3req.SelectConfig[T]{
		AllSQL:    allSQL,
		FieldMap:  compileMap,
		LowerCols: lowerEm,
		OnPanic:   onPanic,
		AutoTotal: true,
	}

	req, err := w3req.NewSelectRequester[T](&opt)
	if err != nil {
		panic(err)
	}

	return &DataRequester[T]{
		sel:        req,
		errorCodes: ErrorCodes(errorCodes),
		onPanic:    onPanic,
		logger:     &Logger{},
	}
}

type RequesterOptions[T any] struct {
	GetDatabaseProvider func() w3req.Conn
	ErrorLog            ExtLogger
	FormatFields        func([]T) //для всех записей ответа обработка полей
}

func (d *DataRequester[T]) InitOnce(f func() RequesterOptions[T]) *DataRequester[T] {
	d.sel.InitOnce(func() *w3req.SelectOptions[T] {
		opt := f()
		d.formatFields = opt.FormatFields
		d.logger.setErrorLogger(opt.ErrorLog)
		return &w3req.SelectOptions[T]{
			Logger: d.logger,
			Conn:   opt.GetDatabaseProvider,
		}
	})
	return d
}

// если включен, то пишет дамп запроса и SQL с параметрами
// вызывать внутри InitOnce
func (d *DataRequester[T]) DumpRequests() *DataRequester[T] {
	d.sel.SetDumpRequests(true)
	return d
}

// если указан, то будет журналировать все запросы
// вызывать внутри InitOnce
func (d *DataRequester[T]) SetDebugLog(log ExtLogger) *DataRequester[T] {
	d.logger.setDebugLogger(log)
	return d
}

// если включен, то вместо "Invalid Parameters" будет возвращать настоящую ошибку
// вызывать внутри InitOnce
func (d *DataRequester[T]) OutputOriginalErrorText() *DataRequester[T] {
	d.logger.outputOriginalError = true
	return d
}

type allTableW2UI struct {
	Status  string `json:"status"`
	Total   int64  `json:"total"`
	Records any    `json:"records"`
}

func (d *DataRequester[T]) GetFasthttpRequestHandlerInner(
	w http.ResponseWriter,
	req any,
	limit int,
	appendQuery *Query,
) {
	defer d.onPanic()

	errout := func(t string) {}
	successout := func(b []byte) {}

	switch t := req.(type) {
	case *http.Request:
		errout = func(text string) {
			d.errorCodes.Error(w, text)
		}
		successout = func(b []byte) {
			w.WriteHeader(200)
			w.Header().Set("Content-Type", "application/json")
			w.Write(b)
		}

	case *fasthttp.RequestCtx:
		errout = func(text string) {
			d.errorCodes.CtxRetError(t, text)
		}
		successout = func(b []byte) {
			t.Success("application/json", b)
		}
	}

	q, err := ReadCtxQuery(req)
	if err != nil {
		d.logger.LogError(INVALID_PARAMETERS, err, errout)
		return
	}

	if appendQuery.Sort != nil {
		q.Sort = append(q.Sort, appendQuery.Sort...)
	}

	if appendQuery.Search != nil {
		switch v := appendQuery.Search.(type) {
		case *w3sql.AtomaryCondition:
			q.Search = w3sql.And(q.Search, v)
		case *w3sql.CompoundCondition:
			q.Search = &w3sql.CompoundCondition{
				Op:    v.Op,
				Query: append(v.Query, q.Search),
			}
		}
	}

	rr := allTableW2UI{}

	if q.Search != nil {
		if q.Limit == nil || *q.Limit > limit || *q.Limit == 0 {
			q.Limit = &limit
		}

		records, total, err := d.sel.Handle((*w3sql.Query)(q))
		if err != nil {
			d.logger.LogError(SYSTEM_ERROR, err, errout)
			return
		} else {
			rr.Status = "success"
		}

		rr.Total = total
		rr.Records = records

		if d.formatFields != nil && total != 0 {
			d.formatFields(records)
		}
	}

	buf, _ := json.Marshal(&rr)
	successout(buf)
}

// fasthttp
func (d *DataRequester[T]) GetFasthttpRequestHandler(limit int, appendQuery *Query) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		d.GetFasthttpRequestHandlerInner(nil, ctx, limit, appendQuery)
	})
}

// net/http
func (d *DataRequester[T]) GetHttpRequestHandler(limit int, appendQuery *Query) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.GetFasthttpRequestHandlerInner(w, r, limit, appendQuery)
	})
}
