package xssjson

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestIsHTMLEncoded(t *testing.T) {
	tests := []struct {
		input string
		html  bool
	}{
		{`abcdef`, true},
		{"", true},
		{`foo <`, false},
		{`bar >`, false},
		{`"quotes"`, false},
		{`'quotes'`, false},
		{`&`, true},
	}

	for pos, tt := range tests {
		if IsHTMLEscaped(tt.input) != tt.html {
			t.Errorf("test %d: %s is %v encoded??", pos, tt.input, tt.html)
		}
	}
}

// this test encoding whole complete json document
func TestEncodingWhole(t *testing.T) {
	tests := []struct {
		raw      string
		expected string
	}{
		{`{}`, `{}`},
		{`{"key":"plain"}`, `{"key":"plain"}`},
		{`{"key":"greater than > "}`, `{"key":"greater than &gt; "}`},
		{`{"key":"greater than \u003e "}`, `{"key":"greater than &gt; "}`},
		{`{"key":"less than < "}`, `{"key":"less than &lt; "}`},
		{`{"key":"less than \u003c "}`, `{"key":"less than &lt; "}`},
		{`{"key":"ampersand & "}`, `{"key":"ampersand &amp; "}`},
		{`{"key":"ampersand \u0026 "}`, `{"key":"ampersand &amp; "}`},
		{`{"key":"apos ' "}`, `{"key":"apos &#x27; "}`},
		{`{"key":"apos \u0027 "}`, `{"key":"apos &#x27; "}`},
		//		{`{"key":"slash / "}`, `{"key":"slash &#x2F; "}`},
		//		{`{"key":"slash \u002F "}`, `{"key":"slash &#x2F; "}`},
		{`{"key":"quote \" "}`, `{"key":"quote &quot; "}`},
		{`{"key":"quote \u0022 "}`, `{"key":"quote &quot; "}`},
		{`{"key":"newline \n "}`, `{"key":"newline \n "}`},
		{`{"key":"Hello, \u2318"}`, `{"key":"Hello, \u2318"}`},
	}

	for pos, tt := range tests {
		buf := bytes.Buffer{}
		e := NewEncoder(&buf)
		e.Write([]byte(tt.raw))

		// unmarshal to make sure output is valid
		tmp := make(map[string]interface{})
		err := json.Unmarshal(buf.Bytes(), &tmp)
		if err != nil {
			t.Errorf("Output was not json: %s", buf.String())
		}

		actual := buf.String()
		if tt.expected != actual {
			t.Errorf("Test %d with %q: expected %s got %s", pos, tt.raw, tt.expected, actual)
		}

		// Ok now we are going to split the input into parts and feed in separately
		for i := 1; i < len(tt.raw)-1; i++ {
			buf.Reset()
			e := NewEncoder(&buf)
			e.Write([]byte(tt.raw[0:i]))
			e.Write([]byte(tt.raw[i:]))

			tmp := make(map[string]interface{})
			err := json.Unmarshal(buf.Bytes(), &tmp)
			if err != nil {
				t.Errorf("Output was not json: %s", buf.String())
			}

			actual := buf.String()
			if tt.expected != actual {
				t.Errorf("Test %d:%d : got %s", pos, i, actual)
			}
		}
	}
}

// This tests using the native golang JSON marshalers
func TestLiveEncoding(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"", `{"key":""}`},
		{"plain", `{"key":"plain"}`},
		{"slash /", `{"key":"slash /"}`},
		{"greater than > ", `{"key":"greater than &gt; "}`},
		{"less than < ", `{"key":"less than &lt; "}`},
		{"ampersand & ", `{"key":"ampersand &amp; "}`},
		{"apos ' ", `{"key":"apos &#x27; "}`},
		{"quote \" ", `{"key":"quote &quot; "}`},
	}

	for pos, tt := range tests {
		amap := map[string]string{
			"key": tt.key,
		}
		raw, err := json.Marshal(amap)
		if err != nil {
			panic(err)
		}
		buf := bytes.Buffer{}
		e := NewEncoder(&buf)
		e.Write(raw)

		// unmarshal to make sure out put is valid
		tmp := make(map[string]interface{})
		err = json.Unmarshal(buf.Bytes(), &tmp)
		if err != nil {
			t.Errorf("Output was not json: %s", buf.String())
		}

		actual := buf.String()
		if tt.expected != actual {
			t.Errorf("Test %d with %q: got %s", pos, tt.key, actual)
		}
	}
}

// Test a struct that needs some fields escaped but not others
func TestMixedSafety(t *testing.T) {
	t.Skip()

	type mixSafety struct {
		Safe   string
		Unsafe string
	}

	safe := "http://www.test.com/?abc=123&def=456"
	unsafe := ">&<"
	obj := mixSafety{
		Safe:   safe,
		Unsafe: unsafe,
	}

	// Test default json
	buf := bytes.Buffer{}
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(obj)

	if err != nil {
		t.Errorf("Unable to marshal!: %s", err)
	}

	expected := `{"Safe":"http://www.test.com/?abc=123\u0026def=456","Unsafe":"\u003e\u0026\u003c"}`
	actual := strings.TrimSpace(buf.String())
	if actual != expected {
		t.Errorf("JSON failed actual: '%s', expected: '%s'", actual, expected)
	}

	// Test xxs
	buf.Reset()
	encoder = json.NewEncoder(NewEncoder(&buf))
	err = encoder.Encode(obj)

	if err != nil {
		t.Errorf("Unable to marshal!: %s", err)
	}

	expected = `{"Safe":"http://www.test.com/?abc=123\u0026def=456","Unsafe":"&gt;&amp;&lt;"}`
	actual = strings.TrimSpace(buf.String())
	if actual != expected {
		t.Errorf("XSS failed actual: '%s', expected: '%s'", actual, expected)
	}
}
