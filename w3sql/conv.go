package w3sql

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

var AllowNaN = false

// иногда база и внутренняя логика портится, если f = NaN
func fixNaN(f float64) float64 {
	if AllowNaN {
		return f
	}
	if math.IsNaN(f) {
		return 0
	}
	return f
}

func getString(v any) string {
	s, ok := v.(string)
	if !ok {
		s = fmt.Sprint(v)
	}
	return s
}

func getFloat(v any) (float64, error) {
	switch vt := v.(type) {
	case float64:
		return vt, nil
	case float32:
		return float64(vt), nil
	case int64:
		return float64(vt), nil
	case int:
		return float64(vt), nil
	case string:
		return strconv.ParseFloat(vt, 64)
	}
	return 0, errors.New("unknown type of floating point parameter")
}

func convNumber(t any) (float64, error) {
	v, err := getFloat(t)
	if err != nil {
		return 0, err
	}
	return fixNaN(v), nil
}

func dateFmt(d any) (string, error) {
	date := getString(d)
	var r time.Time
	var err error
	for _, fmt := range []string{"2006/1/2", "2/1/2006", "02.01.2006", "2.1.2006", "2006-Jan-02"} {
		if r, err = time.Parse(fmt, date); err == nil {
			return r.Format("2006-01-02"), nil
		}
	}
	return "", err
}

func dateTimeFmt(d any) (string, error) {
	date := getString(d)
	var r time.Time
	var err error
	for _, fmt := range []string{"2006/1/2 15:04:05", "2/1/2006 15:04:05", "02.01.2006 15:04:05", "2.1.2006 15:04:05", "2006-Jan-02 15:04:05"} {
		if r, err = time.Parse(fmt, date); err == nil {
			return r.Format("2006-01-02 15:04:05"), nil
		}
	}
	return "", err
}

func convRange(t []any, tp string) (rng struct {
	from any
	to   any
}, err error) {
	switch tp {
	case "date":
		rng.from, err = dateFmt(t[0])
		rng.to, err = dateFmt(t[1])
	case "datetime":
		rng.from, err = dateTimeFmt(t[0])
		rng.to, err = dateTimeFmt(t[1])
	case "numeric":
		rng.from, err = getFloat(t[0])
		rng.to, err = getFloat(t[1])
	default:
		rng.from = t[0]
		rng.to = t[1]
	}
	return
}

func convList(t []any) (l any, err error) {
	if len(t) > 0 {
		return t[0], nil
	}
	return "", nil
}

func convValueElem(t any, tp string) (any, error) {
	switch tp {
	case "text", "textis", "list", "string":
		return fmt.Sprint(t), nil
	case "number", "int", "float":
		return convNumber(t)
	case "date":
		return dateFmt(t)
	case "datetime":
		return dateTimeFmt(t)
	default:
		return nil, errors.New("w3sql: '" + tp + "' is not supported")
	}
}

func convValue_(ts []any, tp string) (any, error) {
	if len(ts) < 1 {
		return nil, errors.New("no value")
	}
	switch tp {
	case "text", "string", "number", "int", "float", "date", "datetime", "textis":
		return convValueElem(ts[0], tp)
	case "list", "bool":
		return convList(ts)
	case "enum":
		return nil, errors.New("w3sql: 'enum' is not supported yet")
	default:
		return nil, errors.New("w3sql: '" + tp + "' is not supported")
	}
}

func convValue(ts any, tp string) (any, error) {
	if vt, ok := ts.([]any); ok {
		return convValue_(vt, tp)
	}
	return convValue_([]any{ts}, tp)
}
