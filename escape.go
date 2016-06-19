package xssjson

import (
	"io"
	"strconv"
	"strings"
)

const (
	stateTop = iota
	stateQuote
	stateBackslash
	stateUnicode1
	stateUnicode2
	stateUnicode3
	stateUnicode4
)

const (
	escapedGT          = "&gt;"
	escapedLT          = "&lt;"
	escapedAmp         = "&amp;"
	escapedSingleQuote = "&#x27;"
	escapedDoubleQuote = "&quot;"
)

// Encoder the state of the XssJSON encoder
type Encoder struct {
	io.Writer
	state int
	uchar []byte
}

// NewEncoder creates a new XssJSON encoder
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		Writer: w,
		state:  stateTop,
		uchar:  make([]byte, 6),
	}
}

// IsHTMLEscaped a raw string HTML escaped or not?
// Work in Progress -- reall just for testing
func IsHTMLEscaped(s string) bool {
	// easy case -- if does contain any
	//  * greater than
	//   * less than
	//  * single quote
	//  * double quote
	// then not escaped
	if strings.ContainsAny(s, `<>'"`) {
		return false
	}
	// contains no <,>,',"
	// does it contain ampersands?
	// if not then definitely safe
	if !strings.Contains(s, "&") {
		return true
	}

	// we have &... but is it an entity? or a raw &?
	// have to pick out & and see what comes after it and if it looks valid or not
	return true
}

// Write is the standard io.Writer interface
func (w *Encoder) Write(s []byte) (int, error) {
	last := 0
	for pos, c := range s {
		switch w.state {
		case stateTop:
			if c == '"' {
				w.state = stateQuote
			}
		case stateQuote:
			switch c {
			case '"':
				w.state = stateTop
			case '\\':
				// backslash is a barrier... write all chars before it
				// we don't want a partial read here where "\u" is one write chunk
				// and the 0123 is another
				w.Writer.Write(s[last:pos])
				w.state = stateBackslash
			case '&':
				w.Writer.Write(s[last:pos])
				w.Writer.Write([]byte(escapedAmp))
				last = pos + 1
			case '<':
				w.Writer.Write(s[last:pos])
				w.Writer.Write([]byte(escapedLT))
				last = pos + 1
			case '>':
				w.Writer.Write(s[last:pos])
				w.Writer.Write([]byte(escapedGT))
				last = pos + 1
			case '\'':
				w.Writer.Write(s[last:pos])
				w.Writer.Write([]byte(escapedSingleQuote))
				last = pos + 1
				//			case '/':
				//				w.Writer.Write(s[last:pos])
				//				w.Writer.Write([]byte("&#x2F;"))
				//				last = pos + 1
			}
		case stateBackslash:
			{
				switch c {
				case '"':
					w.Writer.Write([]byte(escapedDoubleQuote))
					last = pos + 1
					w.state = stateQuote
				case 'u':
					w.uchar[0] = '\\'
					w.uchar[1] = 'u'
					w.state = stateUnicode1
				default:
					w.uchar[0] = '\\'
					w.uchar[1] = c
					// its some other escape sequence.. just emit whatever it is
					w.state = stateQuote
					w.Writer.Write(w.uchar[0:2])
					last = pos + 1
				}
			}
		case stateUnicode1:
			w.uchar[2] = c
			w.state = stateUnicode2
		case stateUnicode2:
			w.uchar[3] = c
			w.state = stateUnicode3
		case stateUnicode3:
			w.uchar[4] = c
			w.state = stateUnicode4
		case stateUnicode4:
			w.uchar[5] = c
			// no matter happens we advance
			last = pos + 1
			w.state = stateQuote

			// parse 4 hexs as integer.  Since we only want ascii we can
			// request that is fits in 8 bytes
			val, err := strconv.ParseUint(string(w.uchar[2:6]), 16, 8)
			if err != nil {
				val = 0
			}
			switch uint8(val) {
			default:
				w.Writer.Write(w.uchar)
			case '"':
				w.Writer.Write([]byte(escapedDoubleQuote))
			case '&':
				w.Writer.Write([]byte(escapedAmp))
			case '<':
				w.Writer.Write([]byte(escapedLT))
			case '>':
				w.Writer.Write([]byte(escapedGT))
			case '\'':
				w.Writer.Write([]byte(escapedSingleQuote))
				//			case '/':
				//				w.Writer.Write([]byte("&#x2F;"))
			}
		}
	}

	total := len(s)
	if last < total && (w.state == stateTop || w.state == stateQuote) {
		w.Writer.Write(s[last:])
	}
	return total, nil
}
