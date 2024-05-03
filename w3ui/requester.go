package w3ui

import (
	"encoding/json"
	"net/http"

	"github.com/MasterDimmy/zipologger"
	"github.com/algebrain/w3/w3req"
	"github.com/algebrain/w3/w3sql"
	"github.com/valyala/fasthttp"
)

type Logger struct {
	errorLog            *zipologger.Logger
	debugLog            *zipologger.Logger
	outputOriginalError bool
}

type DataRequester[T any] struct {
	sel          w3req.SelectRequester[T]
	formatFields func([]T)
	logger       *Logger
	errorCodes   ErrorCodes
}

func newLogger(errorLog *zipologger.Logger) *Logger {
	return &Logger{
		errorLog: errorLog,
	}
}

func (log *Logger) setDebugLogger(z *zipologger.Logger) {
	log.debugLog = z
}

func (log *Logger) LogSQL(string, string, map[string]any) {
	//TODO
}

func (log *Logger) Logf(string, ...any) {
	//TODO
}

func (log *Logger) LogError(prefix string, err error) {
	//TODO
}

// не требуется select count(*) as total - считается динамически:
// total = count(ret)+1 от текущего запроса если count(ret) = limit, иначе count(ret)
func NewDataRequester3[T any](
	allSQL *w3sql.SQLString, //запрос
	compileMap map[string]string, //карта соответствия фронт аргумент -> sql
	lowerEm []string, //значения поискового запроса фронта будут to_lower
	errorCodes map[string]int, //коды ошибок
) *DataRequester[T] {
	opt := w3req.SelectConfig[T]{
		AllSQL:    allSQL,
		FieldMap:  compileMap,
		LowerCols: lowerEm,
		OnPanic:   zipologger.HandlePanic,
	}

	req, err := w3req.NewSelectRequester[T](&opt)
	if err != nil {
		panic(err)
	}

	return &DataRequester[T]{
		sel:        req,
		errorCodes: ErrorCodes(errorCodes),
	}
}

type RequesterOptions[T any] struct {
	And                 *w3sql.Query
	Or                  *w3sql.Query
	GetDatabaseProvider func() w3req.Conn
	ErrorLog            *zipologger.Logger
	FormatFields        func([]T) //для всех записей ответа обработка полей
}

func (d *DataRequester[T]) InitOnce(f func() RequesterOptions[T]) *DataRequester[T] {
	d.sel.InitOnce(func() *w3req.SelectOptions[T] {
		opt := f()
		logger := newLogger(opt.ErrorLog)
		d.formatFields = opt.FormatFields
		d.logger = logger
		return &w3req.SelectOptions[T]{
			Logger: logger,
			Conn:   opt.GetDatabaseProvider,
			And:    opt.And,
			Or:     opt.Or,
		}
	})
	return d
}

// если включен, то пишет дамп запроса и SQL с параметрами
func (d *DataRequester[T]) DumpRequests() *DataRequester[T] {
	d.sel.SetDumpRequests(true)
	return d
}

// если указан, то будет журналировать все запросы
func (d *DataRequester[T]) SetDebugLog(log *zipologger.Logger) *DataRequester[T] {
	d.logger.setDebugLogger(log)
	return d
}

// если включен, то вместо "Invalid Parameters" будет возвращать настоящую ошибку
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
	defer zipologger.HandlePanic()

	q, err := ReadCtxQuery(req)
	if err != nil {
		d.logger.LogError("Invalid parameters", err)
		return
	}

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

	rr := allTableW2UI{}

	if q.Search != nil {
		if q.Limit == nil || *q.Limit > limit || *q.Limit == 0 {
			q.Limit = &limit
		}

		records, total, err := d.sel.Handle(q)
		if err != nil {
			errout("System error") // TODO
			d.logger.LogError("System error", err)
			return
		}

		if err == nil {
			rr.Status = "success"
		} else {
			errout("System error") // TODO
			d.logger.LogError("RequestDatabase:", err)
			return
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
