package w3ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/MasterDimmy/zipologger"
	"github.com/algebrain/w3/w3sql"
	"github.com/valyala/fasthttp"
)

type ErrorCodes map[string]int

var urlReplacer = strings.NewReplacer("/", "_", ".", "")

type TWebAnswer struct {
	Status  string
	ErrCode int         `json:",omitempty"` //код ошибки, если есть, описан в strings, =0 если все ОК
	Message string      `json:",omitempty"`
	Record  interface{} `json:",omitempty"`
}

type W2UIError struct {
	Status  string `json:"status"`
	ErrCode int    `json:"errcode,omitempty"`
	Message string `json:"message"`
}

func ReadCtxQuery(req any) (*w3sql.Query, error) {
	var rq w3sql.Query
	var decoder *json.Decoder

	switch t := req.(type) {
	case *http.Request:
		decoder = json.NewDecoder(t.Body)
	case *fasthttp.RequestCtx:
		decoder = json.NewDecoder(t.RequestBodyStream())
	}

	err := decoder.Decode(&rq)
	return &rq, err
}

func (codes ErrorCodes) Error(w http.ResponseWriter, msg string) string {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	code, ok := codes[msg]
	if !ok {
		code = 1
	}

	ret := W2UIError{
		Status:  "error",
		ErrCode: code,
		Message: msg,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", " ")
	err := encoder.Encode(ret)
	if err != nil {
		w.Write([]byte(`{
    "status"  : "error"
	"message" : "Can not marshal json data"
}`,
		))
	}
	return msg
}

//указаны в приоритете появления
/*
2021/09/29 17:39:46 167.99.230.206: GET / HTTP/1.0
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36 Edg/94.0.992$
Host: some.some
X-Realdsl-Ip: 167.99.230.206 << приоритет ниже
X-Real-Ip: 182.64.140.164    << приоритет выше
*/
var realIPHeaderNameSources = []string{
	"CF-Connecting-IP",
	"X-Real-IP",
	"X-Real-Ip",
	"X-Real-TRAPPER-IP",
	"X-Realdsl-Ip",
	"X-TRAPPER-IP",
	"X-Client-IP",
}

func GetRealIP(ctx *fasthttp.RequestCtx) string {
	for _, v := range realIPHeaderNameSources {
		cf := ctx.Request.Header.Peek(v)
		if len(cf) > 0 {
			return string(cf)
		}
	}

	return ctx.RemoteIP().To4().String()
}

func (codes ErrorCodes) ctxRetErrorCommon(ctx *fasthttp.RequestCtx, text string) string {
	code, ok := codes[text]
	if !ok {
		code = 1
	}

	buf, _ := json.MarshalIndent(&TWebAnswer{
		Status:  "error",
		ErrCode: code,
		Message: text,
	}, "", " ")
	ctx.Success("application/json", buf)
	return string(buf) //+ string(ctx.PostBody())
}

func (codes ErrorCodes) CtxRetError(ctx *fasthttp.RequestCtx, text string) string {
	if runtime.GOOS == "windows" {
		_, f, l, _ := runtime.Caller(1)
		pos := fmt.Sprintf("%s:%d: ", f, l)
		os.Stdout.WriteString("\n-->> " + pos + text + "\n\n")
	}

	if ctx == nil {
		return text
	}
	ipv4 := GetRealIP(ctx)

	st := ""
	if !ctx.IsPut() {
		m := len(ctx.PostBody())
		if m > 1000 {
			m = 1000
		}
		st = string(ctx.PostBody()[:m])
	}

	requrl := string(ctx.Request.URI().Path())
	upd_url, ok := ctx.UserValue("URL").(string)
	if ok {
		requrl = upd_url
	}

	ret := ""
	if len(st) > 0 {
		ret = fmt.Sprintf(
			"[%s]=>[%s] => %s\n-----\n%s\n-----",
			ipv4,
			ctx.Request.RequestURI(),
			codes.ctxRetErrorCommon(ctx, text),
			st,
		)
	} else {
		ret = fmt.Sprintf(
			"[%s]=>[%s] => %s\n",
			ipv4,
			ctx.Request.RequestURI(),
			codes.ctxRetErrorCommon(ctx, text),
		)
	}
	logged := zipologger.GetLoggerBySuffix(
		urlReplacer.Replace(requrl),
		"./logs/answers_error_",
		3, 3, 3, false,
	).Print(ret)
	return ipv4 + ":[" + string(ctx.Request.RequestURI()) +
		"] " + text + "\n" + logged
}
