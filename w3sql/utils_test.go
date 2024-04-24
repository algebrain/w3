package w3sql

import (
	"fmt"
	"testing"
)

func TestNormalizeSQLString(t *testing.T) {
	s := "abc\n\r\tabc  \t\r\n"
	if ss := normalizeSQLString(s); ss != "abc abc " {
		t.Fatal("wrong result:", ss)
	} else {
		fmt.Printf("in: <%s>\n", s)
		fmt.Printf("out: <%s>\n", ss)
	}
}

func TestRemoveRoundBracketsContents(t *testing.T) {
	s := `abc 'abc\' (sdfsd)' sdfsdf (sdfsd)sfw`
	if ss := removeRoundBracketsContents(s); ss != `abc 'abc\' (sdfsd)' sdfsdf sfw` {
		t.Fatal("wrong result:", ss)
	} else {
		fmt.Printf("in: <%s>\n", s)
		fmt.Printf("out: <%s>\n", ss)
	}
}

func TestEqualSQLStrings(t *testing.T) {
	s := "\n abc\n\r\tabc  \t\r\n"
	if !equalSQLStrings(s, "abc abc") {
		t.Fatal("equality expected")
	}
	if equalSQLStrings(s, "abc vabc") {
		t.Fatal("inequality expected")
	}
}
