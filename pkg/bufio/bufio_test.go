// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bufio_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	. "github.com/Terry-Mao/goim/pkg/bufio"
)

// Reads from a reader and rot13s the result.
type rot13Reader struct {
	r io.Reader
}

func newRot13Reader(r io.Reader) *rot13Reader {
	r13 := new(rot13Reader)
	r13.r = r
	return r13
}

func (r13 *rot13Reader) Read(p []byte) (int, error) {
	n, err := r13.r.Read(p)
	for i := 0; i < n; i++ {
		c := p[i] | 0x20 // lowercase byte
		if 'a' <= c && c <= 'm' {
			p[i] += 13
		} else if 'n' <= c && c <= 'z' {
			p[i] -= 13
		}
	}
	return n, err
}

// Call ReadByte to accumulate the text of a file
func readBytes(buf *Reader) string {
	var b [1000]byte
	nb := 0
	for {
		c, err := buf.ReadByte()
		if err == io.EOF {
			break
		}
		if err == nil {
			b[nb] = c
			nb++
		} else if err != iotest.ErrTimeout {
			panic("Data: " + err.Error())
		}
	}
	return string(b[0:nb])
}

func TestReaderSimple(t *testing.T) {
	data := "hello world"
	b := NewReader(strings.NewReader(data))
	if s := readBytes(b); s != "hello world" {
		t.Errorf("simple hello world test failed: got %q", s)
	}

	b = NewReader(newRot13Reader(strings.NewReader(data)))
	if s := readBytes(b); s != "uryyb jbeyq" {
		t.Errorf("rot13 hello world test failed: got %q", s)
	}
}

type readMaker struct {
	name string
	fn   func(io.Reader) io.Reader
}

var readMakers = []readMaker{
	{"full", func(r io.Reader) io.Reader { return r }},
	{"byte", iotest.OneByteReader},
	{"half", iotest.HalfReader},
	{"data+err", iotest.DataErrReader},
	{"timeout", iotest.TimeoutReader},
}

// Call Read to accumulate the text of a file
func reads(buf *Reader, m int) string {
	var b [1000]byte
	nb := 0
	for {
		n, err := buf.Read(b[nb : nb+m])
		nb += n
		if err == io.EOF {
			break
		}
	}
	return string(b[0:nb])
}

type bufReader struct {
	name string
	fn   func(*Reader) string
}

var bufreaders = []bufReader{
	{"1", func(b *Reader) string { return reads(b, 1) }},
	{"2", func(b *Reader) string { return reads(b, 2) }},
	{"3", func(b *Reader) string { return reads(b, 3) }},
	{"4", func(b *Reader) string { return reads(b, 4) }},
	{"5", func(b *Reader) string { return reads(b, 5) }},
	{"7", func(b *Reader) string { return reads(b, 7) }},
	{"bytes", readBytes},
}

const minReadBufferSize = 16

var bufsizes = []int{
	0, minReadBufferSize, 23, 32, 46, 64, 93, 128, 1024, 4096,
}

func TestReader(t *testing.T) {
	var texts [31]string
	str := ""
	all := ""
	for i := 0; i < len(texts)-1; i++ {
		texts[i] = str + "\n"
		all += texts[i]
		str += fmt.Sprintf("%x", i%26+'a')
	}
	texts[len(texts)-1] = all

	for h := 0; h < len(texts); h++ {
		text := texts[h]
		for i := 0; i < len(readMakers); i++ {
			for j := 0; j < len(bufreaders); j++ {
				for k := 0; k < len(bufsizes); k++ {
					readmaker := readMakers[i]
					bufreader := bufreaders[j]
					bufsize := bufsizes[k]
					read := readmaker.fn(strings.NewReader(text))
					buf := NewReaderSize(read, bufsize)
					s := bufreader.fn(buf)
					if s != text {
						t.Errorf("reader=%s fn=%s bufsize=%d want=%q got=%q",
							readmaker.name, bufreader.name, bufsize, text, s)
					}
				}
			}
		}
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	return 0, nil
}

func TestZeroReader(t *testing.T) {
	var z zeroReader
	r := NewReader(z)

	c := make(chan error)
	go func() {
		_, err := r.ReadByte()
		c <- err
	}()

	select {
	case err := <-c:
		if err == nil {
			t.Error("error expected")
		} else if err != io.ErrNoProgress {
			t.Error("unexpected error:", err)
		}
	case <-time.After(time.Second):
		t.Error("test timed out (endless loop in ReadByte?)")
	}
}

func TestWriter(t *testing.T) {
	var data [8192]byte

	for i := 0; i < len(data); i++ {
		data[i] = byte(' ' + i%('~'-' '))
	}
	w := new(bytes.Buffer)
	for i := 0; i < len(bufsizes); i++ {
		for j := 0; j < len(bufsizes); j++ {
			nwrite := bufsizes[i]
			bs := bufsizes[j]

			// Write nwrite bytes using buffer size bs.
			// Check that the right amount makes it out
			// and that the data is correct.

			w.Reset()
			buf := NewWriterSize(w, bs)
			context := fmt.Sprintf("nwrite=%d bufsize=%d", nwrite, bs)
			n, e1 := buf.Write(data[0:nwrite])
			if e1 != nil || n != nwrite {
				t.Errorf("%s: buf.Write %d = %d, %v", context, nwrite, n, e1)
				continue
			}
			if e := buf.Flush(); e != nil {
				t.Errorf("%s: buf.Flush = %v", context, e)
			}

			written := w.Bytes()
			if len(written) != nwrite {
				t.Errorf("%s: %d bytes written", context, len(written))
			}
			for l := 0; l < len(written); l++ {
				if written[i] != data[i] {
					t.Errorf("wrong bytes written")
					t.Errorf("want=%q", data[0:len(written)])
					t.Errorf("have=%q", written)
				}
			}
		}
	}
}

// Check that write errors are returned properly.

type errorWriterTest struct {
	n, m   int
	err    error
	expect error
}

func (w errorWriterTest) Write(p []byte) (int, error) {
	return len(p) * w.n / w.m, w.err
}

var errorWriterTests = []errorWriterTest{
	{0, 1, nil, io.ErrShortWrite},
	{1, 2, nil, io.ErrShortWrite},
	{1, 1, nil, nil},
	{0, 1, io.ErrClosedPipe, io.ErrClosedPipe},
	{1, 2, io.ErrClosedPipe, io.ErrClosedPipe},
	{1, 1, io.ErrClosedPipe, io.ErrClosedPipe},
}

func TestWriteErrors(t *testing.T) {
	for _, w := range errorWriterTests {
		buf := NewWriter(w)
		_, e := buf.Write([]byte("hello world"))
		if e != nil {
			t.Errorf("Write hello to %v: %v", w, e)
			continue
		}
		// Two flushes, to verify the error is sticky.
		for i := 0; i < 2; i++ {
			e = buf.Flush()
			if e != w.expect {
				t.Errorf("Flush %d/2 %v: got %v, wanted %v", i+1, w, e, w.expect)
			}
		}
	}
}

func TestNewReaderSizeIdempotent(t *testing.T) {
	const BufSize = 1000
	b := NewReaderSize(strings.NewReader("hello world"), BufSize)
	// Does it recognize itself?
	b1 := NewReaderSize(b, BufSize)
	if b1 != b {
		t.Error("NewReaderSize did not detect underlying Reader")
	}
	// Does it wrap if existing buffer is too small?
	b2 := NewReaderSize(b, 2*BufSize)
	if b2 == b {
		t.Error("NewReaderSize did not enlarge buffer")
	}
}

func TestNewWriterSizeIdempotent(t *testing.T) {
	const BufSize = 1000
	b := NewWriterSize(new(bytes.Buffer), BufSize)
	// Does it recognize itself?
	b1 := NewWriterSize(b, BufSize)
	if b1 != b {
		t.Error("NewWriterSize did not detect underlying Writer")
	}
	// Does it wrap if existing buffer is too small?
	b2 := NewWriterSize(b, 2*BufSize)
	if b2 == b {
		t.Error("NewWriterSize did not enlarge buffer")
	}
}

func TestWriteString(t *testing.T) {
	const BufSize = 8
	buf := new(bytes.Buffer)
	b := NewWriterSize(buf, BufSize)
	_, _ = b.WriteString("0")                         // easy
	_, _ = b.WriteString("123456")                    // still easy
	_, _ = b.WriteString("7890")                      // easy after flush
	_, _ = b.WriteString("abcdefghijklmnopqrstuvwxy") // hard
	_, _ = b.WriteString("z")
	if err := b.Flush(); err != nil {
		t.Error("WriteString", err)
	}
	s := "01234567890abcdefghijklmnopqrstuvwxyz"
	if buf.String() != s {
		t.Errorf("WriteString wants %q gets %q", s, buf.String())
	}
}

func TestBufferFull(t *testing.T) {
	const longString = "And now, hello, world! It is the time for all good men to come to the aid of their party"
	buf := NewReaderSize(strings.NewReader(longString), minReadBufferSize)
	line, err := buf.ReadSlice('!')
	if string(line) != "And now, hello, " || err != ErrBufferFull {
		t.Errorf("first ReadSlice(,) = %q, %v", line, err)
	}
	line, err = buf.ReadSlice('!')
	if string(line) != "world!" || err != nil {
		t.Errorf("second ReadSlice(,) = %q, %v", line, err)
	}
}

func TestPeek(t *testing.T) {
	p := make([]byte, 10)
	// string is 16 (minReadBufferSize) long.
	buf := NewReaderSize(strings.NewReader("abcdefghijklmnop"), minReadBufferSize)
	if s, err := buf.Peek(1); string(s) != "a" || err != nil {
		t.Fatalf("want %q got %q, err=%v", "a", string(s), err)
	}
	if s, err := buf.Peek(4); string(s) != "abcd" || err != nil {
		t.Fatalf("want %q got %q, err=%v", "abcd", string(s), err)
	}
	if _, err := buf.Peek(-1); err != ErrNegativeCount {
		t.Fatalf("want ErrNegativeCount got %v", err)
	}
	if _, err := buf.Peek(32); err != ErrBufferFull {
		t.Fatalf("want ErrBufFull got %v", err)
	}
	if _, err := buf.Read(p[0:3]); string(p[0:3]) != "abc" || err != nil {
		t.Fatalf("want %q got %q, err=%v", "abc", string(p[0:3]), err)
	}
	if s, err := buf.Peek(1); string(s) != "d" || err != nil {
		t.Fatalf("want %q got %q, err=%v", "d", string(s), err)
	}
	if s, err := buf.Peek(2); string(s) != "de" || err != nil {
		t.Fatalf("want %q got %q, err=%v", "de", string(s), err)
	}
	if _, err := buf.Read(p[0:3]); string(p[0:3]) != "def" || err != nil {
		t.Fatalf("want %q got %q, err=%v", "def", string(p[0:3]), err)
	}
	if s, err := buf.Peek(4); string(s) != "ghij" || err != nil {
		t.Fatalf("want %q got %q, err=%v", "ghij", string(s), err)
	}
	if _, err := buf.Read(p[0:]); string(p[0:]) != "ghijklmnop" || err != nil {
		t.Fatalf("want %q got %q, err=%v", "ghijklmnop", string(p[0:minReadBufferSize]), err)
	}
	if s, err := buf.Peek(0); string(s) != "" || err != nil {
		t.Fatalf("want %q got %q, err=%v", "", string(s), err)
	}
	if _, err := buf.Peek(1); err != io.EOF {
		t.Fatalf("want EOF got %v", err)
	}

	// Test for issue 3022, not exposing a reader's error on a successful Peek.
	buf = NewReaderSize(dataAndEOFReader("abcd"), 32)
	if s, err := buf.Peek(2); string(s) != "ab" || err != nil {
		t.Errorf(`Peek(2) on "abcd", EOF = %q, %v; want "ab", nil`, string(s), err)
	}
	if s, err := buf.Peek(4); string(s) != "abcd" || err != nil {
		t.Errorf(`Peek(4) on "abcd", EOF = %q, %v; want "abcd", nil`, string(s), err)
	}
	if n, err := buf.Read(p[0:5]); string(p[0:n]) != "abcd" || err != nil {
		t.Fatalf("Read after peek = %q, %v; want abcd, EOF", p[0:n], err)
	}
	if n, err := buf.Read(p[0:1]); string(p[0:n]) != "" || err != io.EOF {
		t.Fatalf(`second Read after peek = %q, %v; want "", EOF`, p[0:n], err)
	}
}

type dataAndEOFReader string

func (r dataAndEOFReader) Read(p []byte) (int, error) {
	return copy(p, r), io.EOF
}

var testOutput = []byte("0123456789abcdefghijklmnopqrstuvwxy")
var testInput = []byte("012\n345\n678\n9ab\ncde\nfgh\nijk\nlmn\nopq\nrst\nuvw\nxy")
var testInputrn = []byte("012\r\n345\r\n678\r\n9ab\r\ncde\r\nfgh\r\nijk\r\nlmn\r\nopq\r\nrst\r\nuvw\r\nxy\r\n\n\r\n")

// TestReader wraps a []byte and returns reads of a specific length.
type testReader struct {
	data   []byte
	stride int
}

func (t *testReader) Read(buf []byte) (n int, err error) {
	n = t.stride
	if n > len(t.data) {
		n = len(t.data)
	}
	if n > len(buf) {
		n = len(buf)
	}
	copy(buf, t.data)
	t.data = t.data[n:]
	if len(t.data) == 0 {
		err = io.EOF
	}
	return
}

func testReadLine(t *testing.T, input []byte) {
	//for stride := 1; stride < len(input); stride++ {
	for stride := 1; stride < 2; stride++ {
		done := 0
		reader := testReader{input, stride}
		l := NewReaderSize(&reader, len(input)+1)
		for {
			line, isPrefix, err := l.ReadLine()
			if len(line) > 0 && err != nil {
				t.Errorf("ReadLine returned both data and error: %s", err)
			}
			if isPrefix {
				t.Errorf("ReadLine returned prefix")
			}
			if err != nil {
				if err != io.EOF {
					t.Fatalf("Got unknown error: %s", err)
				}
				break
			}
			if want := testOutput[done : done+len(line)]; !bytes.Equal(want, line) {
				t.Errorf("Bad line at stride %d: want: %x got: %x", stride, want, line)
			}
			done += len(line)
		}
		if done != len(testOutput) {
			t.Errorf("ReadLine didn't return everything: got: %d, want: %d (stride: %d)", done, len(testOutput), stride)
		}
	}
}

func TestReadLine(t *testing.T) {
	testReadLine(t, testInput)
	testReadLine(t, testInputrn)
}

func TestLineTooLong(t *testing.T) {
	data := make([]byte, 0)
	for i := 0; i < minReadBufferSize*5/2; i++ {
		data = append(data, '0'+byte(i%10))
	}
	buf := bytes.NewReader(data)
	l := NewReaderSize(buf, minReadBufferSize)
	line, isPrefix, err := l.ReadLine()
	if !isPrefix || !bytes.Equal(line, data[:minReadBufferSize]) || err != nil {
		t.Errorf("bad result for first line: got %q want %q %v", line, data[:minReadBufferSize], err)
	}
	data = data[len(line):]
	line, isPrefix, err = l.ReadLine()
	if !isPrefix || !bytes.Equal(line, data[:minReadBufferSize]) || err != nil {
		t.Errorf("bad result for second line: got %q want %q %v", line, data[:minReadBufferSize], err)
	}
	data = data[len(line):]
	line, isPrefix, err = l.ReadLine()
	if isPrefix || !bytes.Equal(line, data[:minReadBufferSize/2]) || err != nil {
		t.Errorf("bad result for third line: got %q want %q %v", line, data[:minReadBufferSize/2], err)
	}
	line, isPrefix, err = l.ReadLine()
	if isPrefix || err == nil {
		t.Errorf("expected no more lines: %x %s", line, err)
	}
}

func TestReadAfterLines(t *testing.T) {
	line1 := "this is line1"
	restData := "this is line2\nthis is line 3\n"
	inbuf := bytes.NewReader([]byte(line1 + "\n" + restData))
	outbuf := new(bytes.Buffer)
	maxLineLength := len(line1) + len(restData)/2
	l := NewReaderSize(inbuf, maxLineLength)
	line, isPrefix, err := l.ReadLine()
	if isPrefix || err != nil || string(line) != line1 {
		t.Errorf("bad result for first line: isPrefix=%v err=%v line=%q", isPrefix, err, string(line))
	}
	n, err := io.Copy(outbuf, l)
	if int(n) != len(restData) || err != nil {
		t.Errorf("bad result for Read: n=%d err=%v", n, err)
	}
	if outbuf.String() != restData {
		t.Errorf("bad result for Read: got %q; expected %q", outbuf.String(), restData)
	}
}

func TestReadEmptyBuffer(t *testing.T) {
	l := NewReaderSize(new(bytes.Buffer), minReadBufferSize)
	line, isPrefix, err := l.ReadLine()
	if err != io.EOF {
		t.Errorf("expected EOF from ReadLine, got '%s' %t %s", line, isPrefix, err)
	}
}

func TestLinesAfterRead(t *testing.T) {
	l := NewReaderSize(bytes.NewReader([]byte("foo")), minReadBufferSize)
	_, err := ioutil.ReadAll(l)
	if err != nil {
		t.Error(err)
		return
	}

	line, isPrefix, err := l.ReadLine()
	if err != io.EOF {
		t.Errorf("expected EOF from ReadLine, got '%s' %t %s", line, isPrefix, err)
	}
}

func TestReadLineNonNilLineOrError(t *testing.T) {
	r := NewReader(strings.NewReader("line 1\n"))
	for i := 0; i < 2; i++ {
		l, _, err := r.ReadLine()
		if l != nil && err != nil {
			t.Fatalf("on line %d/2; ReadLine=%#v, %v; want non-nil line or Error, but not both",
				i+1, l, err)
		}
	}
}

type readLineResult struct {
	line     []byte
	isPrefix bool
	err      error
}

var readLineNewlinesTests = []struct {
	input  string
	expect []readLineResult
}{
	{"012345678901234\r\n012345678901234\r\n", []readLineResult{
		{[]byte("012345678901234"), true, nil},
		{nil, false, nil},
		{[]byte("012345678901234"), true, nil},
		{nil, false, nil},
		{nil, false, io.EOF},
	}},
	{"0123456789012345\r012345678901234\r", []readLineResult{
		{[]byte("0123456789012345"), true, nil},
		{[]byte("\r012345678901234"), true, nil},
		{[]byte("\r"), false, nil},
		{nil, false, io.EOF},
	}},
}

func TestReadLineNewlines(t *testing.T) {
	for _, e := range readLineNewlinesTests {
		testReadLineNewlines(t, e.input, e.expect)
	}
}

func testReadLineNewlines(t *testing.T, input string, expect []readLineResult) {
	b := NewReaderSize(strings.NewReader(input), minReadBufferSize)
	for i, e := range expect {
		line, isPrefix, err := b.ReadLine()
		if !bytes.Equal(line, e.line) {
			t.Errorf("%q call %d, line == %q, want %q", input, i, line, e.line)
			return
		}
		if isPrefix != e.isPrefix {
			t.Errorf("%q call %d, isPrefix == %v, want %v", input, i, isPrefix, e.isPrefix)
			return
		}
		if err != e.err {
			t.Errorf("%q call %d, err == %v, want %v", input, i, err, e.err)
			return
		}
	}
}

// TestWriterReadFromCounts tests that using io.Copy to copy into a
// bufio.Writer does not prematurely flush the buffer. For example, when
// buffering writes to a network socket, excessive network writes should be
// avoided.
func TestWriterReadFromCounts(t *testing.T) {
	var w0 writeCountingDiscard
	b0 := NewWriterSize(&w0, 1234)
	_, _ = b0.WriteString(strings.Repeat("x", 1000))
	if w0 != 0 {
		t.Fatalf("write 1000 'x's: got %d writes, want 0", w0)
	}
	_, _ = b0.WriteString(strings.Repeat("x", 200))
	if w0 != 0 {
		t.Fatalf("write 1200 'x's: got %d writes, want 0", w0)
	}
	_, _ = io.Copy(b0, onlyReader{strings.NewReader(strings.Repeat("x", 30))})
	if w0 != 0 {
		t.Fatalf("write 1230 'x's: got %d writes, want 0", w0)
	}
	_, _ = io.Copy(b0, onlyReader{strings.NewReader(strings.Repeat("x", 9))})
	if w0 != 1 {
		t.Fatalf("write 1239 'x's: got %d writes, want 1", w0)
	}

	var w1 writeCountingDiscard
	b1 := NewWriterSize(&w1, 1234)
	_, _ = b1.WriteString(strings.Repeat("x", 1200))
	_ = b1.Flush()
	if w1 != 1 {
		t.Fatalf("flush 1200 'x's: got %d writes, want 1", w1)
	}
	_, _ = b1.WriteString(strings.Repeat("x", 89))
	if w1 != 1 {
		t.Fatalf("write 1200 + 89 'x's: got %d writes, want 1", w1)
	}
	_, _ = io.Copy(b1, onlyReader{strings.NewReader(strings.Repeat("x", 700))})
	if w1 != 1 {
		t.Fatalf("write 1200 + 789 'x's: got %d writes, want 1", w1)
	}
	_, _ = io.Copy(b1, onlyReader{strings.NewReader(strings.Repeat("x", 600))})
	if w1 != 2 {
		t.Fatalf("write 1200 + 1389 'x's: got %d writes, want 2", w1)
	}
	_ = b1.Flush()
	if w1 != 3 {
		t.Fatalf("flush 1200 + 1389 'x's: got %d writes, want 3", w1)
	}
}

// A writeCountingDiscard is like ioutil.Discard and counts the number of times
// Write is called on it.
type writeCountingDiscard int

func (w *writeCountingDiscard) Write(p []byte) (int, error) {
	*w++
	return len(p), nil
}

type negativeReader int

func (r *negativeReader) Read([]byte) (int, error) { return -1, nil }

func TestNegativeRead(t *testing.T) {
	// should panic with a description pointing at the reader, not at itself.
	// (should NOT panic with slice index error, for example.)
	b := NewReader(new(negativeReader))
	defer func() {
		switch err := recover().(type) {
		case nil:
			t.Fatal("read did not panic")
		case error:
			if !strings.Contains(err.Error(), "reader returned negative count from Read") {
				t.Fatalf("wrong panic: %v", err)
			}
		default:
			t.Fatalf("unexpected panic value: %T(%v)", err, err)
		}
	}()
	_, _ = b.Read(make([]byte, 100))
}

var errFake = errors.New("fake error")

type errorThenGoodReader struct {
	didErr bool
	nread  int
}

func (r *errorThenGoodReader) Read(p []byte) (int, error) {
	r.nread++
	if !r.didErr {
		r.didErr = true
		return 0, errFake
	}
	return len(p), nil
}

func TestReaderClearError(t *testing.T) {
	r := &errorThenGoodReader{}
	b := NewReader(r)
	buf := make([]byte, 1)
	if _, err := b.Read(nil); err != nil {
		t.Fatalf("1st nil Read = %v; want nil", err)
	}
	if _, err := b.Read(buf); err != errFake {
		t.Fatalf("1st Read = %v; want errFake", err)
	}
	if _, err := b.Read(nil); err != nil {
		t.Fatalf("2nd nil Read = %v; want nil", err)
	}
	if _, err := b.Read(buf); err != nil {
		t.Fatalf("3rd Read with buffer = %v; want nil", err)
	}
	if r.nread != 2 {
		t.Errorf("num reads = %d; want 2", r.nread)
	}
}

func TestReaderReset(t *testing.T) {
	r := NewReader(strings.NewReader("foo foo"))
	buf := make([]byte, 3)
	_, _ = r.Read(buf)
	if string(buf) != "foo" {
		t.Errorf("buf = %q; want foo", buf)
	}
	r.Reset(strings.NewReader("bar bar"))
	all, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(all) != "bar bar" {
		t.Errorf("ReadAll = %q; want bar bar", all)
	}
}

func TestWriterReset(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	w := NewWriter(&buf1)
	_, _ = w.WriteString("foo")
	w.Reset(&buf2) // and not flushed
	_, _ = w.WriteString("bar")
	_ = w.Flush()
	if buf1.String() != "" {
		t.Errorf("buf1 = %q; want empty", buf1.String())
	}
	if buf2.String() != "bar" {
		t.Errorf("buf2 = %q; want bar", buf2.String())
	}
}

func TestReaderDiscard(t *testing.T) {
	tests := []struct {
		name     string
		r        io.Reader
		bufSize  int // 0 means 16
		peekSize int

		n int // input to Discard

		want    int   // from Discard
		wantErr error // from Discard

		wantBuffered int
	}{
		{
			name:         "normal case",
			r:            strings.NewReader("abcdefghijklmnopqrstuvwxyz"),
			peekSize:     16,
			n:            6,
			want:         6,
			wantBuffered: 10,
		},
		{
			name:         "discard causing read",
			r:            strings.NewReader("abcdefghijklmnopqrstuvwxyz"),
			n:            6,
			want:         6,
			wantBuffered: 10,
		},
		{
			name:         "discard all without peek",
			r:            strings.NewReader("abcdefghijklmnopqrstuvwxyz"),
			n:            26,
			want:         26,
			wantBuffered: 0,
		},
		{
			name:         "discard more than end",
			r:            strings.NewReader("abcdefghijklmnopqrstuvwxyz"),
			n:            27,
			want:         26,
			wantErr:      io.EOF,
			wantBuffered: 0,
		},
		// Any error from filling shouldn't show up until we
		// get past the valid bytes. Here we return we return 5 valid bytes at the same time
		// as an error, but test that we don't see the error from Discard.
		{
			name: "fill error, discard less",
			r: newScriptedReader(func(p []byte) (n int, err error) {
				if len(p) < 5 {
					panic("unexpected small read")
				}
				return 5, errors.New("5-then-error")
			}),
			n:            4,
			want:         4,
			wantErr:      nil,
			wantBuffered: 1,
		},
		{
			name: "fill error, discard equal",
			r: newScriptedReader(func(p []byte) (n int, err error) {
				if len(p) < 5 {
					panic("unexpected small read")
				}
				return 5, errors.New("5-then-error")
			}),
			n:            5,
			want:         5,
			wantErr:      nil,
			wantBuffered: 0,
		},
		{
			name: "fill error, discard more",
			r: newScriptedReader(func(p []byte) (n int, err error) {
				if len(p) < 5 {
					panic("unexpected small read")
				}
				return 5, errors.New("5-then-error")
			}),
			n:            6,
			want:         5,
			wantErr:      errors.New("5-then-error"),
			wantBuffered: 0,
		},
		// Discard of 0 shouldn't cause a read:
		{
			name:         "discard zero",
			r:            newScriptedReader(), // will panic on Read
			n:            0,
			want:         0,
			wantErr:      nil,
			wantBuffered: 0,
		},
		{
			name:         "discard negative",
			r:            newScriptedReader(), // will panic on Read
			n:            -1,
			want:         0,
			wantErr:      ErrNegativeCount,
			wantBuffered: 0,
		},
	}
	for _, tt := range tests {
		br := NewReaderSize(tt.r, tt.bufSize)
		if tt.peekSize > 0 {
			peekBuf, err := br.Peek(tt.peekSize)
			if err != nil {
				t.Errorf("%s: Peek(%d): %v", tt.name, tt.peekSize, err)
				continue
			}
			if len(peekBuf) != tt.peekSize {
				t.Errorf("%s: len(Peek(%d)) = %v; want %v", tt.name, tt.peekSize, len(peekBuf), tt.peekSize)
				continue
			}
		}
		discarded, err := br.Discard(tt.n)
		if ge, we := fmt.Sprint(err), fmt.Sprint(tt.wantErr); discarded != tt.want || ge != we {
			t.Errorf("%s: Discard(%d) = (%v, %v); want (%v, %v)", tt.name, tt.n, discarded, ge, tt.want, we)
			continue
		}
		if bn := br.Buffered(); bn != tt.wantBuffered {
			t.Errorf("%s: after Discard, Buffered = %d; want %d", tt.name, bn, tt.wantBuffered)
		}
	}

}

// An onlyReader only implements io.Reader, no matter what other methods the underlying implementation may have.
type onlyReader struct {
	io.Reader
}

// An onlyWriter only implements io.Writer, no matter what other methods the underlying implementation may have.
type onlyWriter struct {
	io.Writer
}

// A scriptedReader is an io.Reader that executes its steps sequentially.
type scriptedReader []func(p []byte) (n int, err error)

func (sr *scriptedReader) Read(p []byte) (n int, err error) {
	if len(*sr) == 0 {
		panic("too many Read calls on scripted Reader. No steps remain.")
	}
	step := (*sr)[0]
	*sr = (*sr)[1:]
	return step(p)
}

func newScriptedReader(steps ...func(p []byte) (n int, err error)) io.Reader {
	sr := scriptedReader(steps)
	return &sr
}

func BenchmarkReaderCopyOptimal(b *testing.B) {
	// Optimal case is where the underlying reader implements io.WriterTo
	srcBuf := bytes.NewBuffer(make([]byte, 8192))
	src := NewReader(srcBuf)
	dstBuf := new(bytes.Buffer)
	dst := onlyWriter{dstBuf}
	for i := 0; i < b.N; i++ {
		srcBuf.Reset()
		src.Reset(srcBuf)
		dstBuf.Reset()
		_, _ = io.Copy(dst, src)
	}
}

func BenchmarkReaderCopyUnoptimal(b *testing.B) {
	// Unoptimal case is where the underlying reader doesn't implement io.WriterTo
	srcBuf := bytes.NewBuffer(make([]byte, 8192))
	src := NewReader(onlyReader{srcBuf})
	dstBuf := new(bytes.Buffer)
	dst := onlyWriter{dstBuf}
	for i := 0; i < b.N; i++ {
		srcBuf.Reset()
		src.Reset(onlyReader{srcBuf})
		dstBuf.Reset()
		_, _ = io.Copy(dst, src)
	}
}

func BenchmarkReaderCopyNoWriteTo(b *testing.B) {
	srcBuf := bytes.NewBuffer(make([]byte, 8192))
	srcReader := NewReader(srcBuf)
	src := onlyReader{srcReader}
	dstBuf := new(bytes.Buffer)
	dst := onlyWriter{dstBuf}
	for i := 0; i < b.N; i++ {
		srcBuf.Reset()
		srcReader.Reset(srcBuf)
		dstBuf.Reset()
		_, _ = io.Copy(dst, src)
	}
}

func BenchmarkWriterCopyOptimal(b *testing.B) {
	// Optimal case is where the underlying writer implements io.ReaderFrom
	srcBuf := bytes.NewBuffer(make([]byte, 8192))
	src := onlyReader{srcBuf}
	dstBuf := new(bytes.Buffer)
	dst := NewWriter(dstBuf)
	for i := 0; i < b.N; i++ {
		srcBuf.Reset()
		dstBuf.Reset()
		dst.Reset(dstBuf)
		_, _ = io.Copy(dst, src)
	}
}

func BenchmarkWriterCopyUnoptimal(b *testing.B) {
	srcBuf := bytes.NewBuffer(make([]byte, 8192))
	src := onlyReader{srcBuf}
	dstBuf := new(bytes.Buffer)
	dst := NewWriter(onlyWriter{dstBuf})
	for i := 0; i < b.N; i++ {
		srcBuf.Reset()
		dstBuf.Reset()
		dst.Reset(onlyWriter{dstBuf})
		_, _ = io.Copy(dst, src)
	}
}

func BenchmarkWriterCopyNoReadFrom(b *testing.B) {
	srcBuf := bytes.NewBuffer(make([]byte, 8192))
	src := onlyReader{srcBuf}
	dstBuf := new(bytes.Buffer)
	dstWriter := NewWriter(dstBuf)
	dst := onlyWriter{dstWriter}
	for i := 0; i < b.N; i++ {
		srcBuf.Reset()
		dstBuf.Reset()
		dstWriter.Reset(dstBuf)
		_, _ = io.Copy(dst, src)
	}
}

func BenchmarkReaderEmpty(b *testing.B) {
	b.ReportAllocs()
	str := strings.Repeat("x", 16<<10)
	for i := 0; i < b.N; i++ {
		br := NewReader(strings.NewReader(str))
		n, err := io.Copy(ioutil.Discard, br)
		if err != nil {
			b.Fatal(err)
		}
		if n != int64(len(str)) {
			b.Fatal("wrong length")
		}
	}
}

func BenchmarkWriterEmpty(b *testing.B) {
	b.ReportAllocs()
	str := strings.Repeat("x", 1<<10)
	bs := []byte(str)
	for i := 0; i < b.N; i++ {
		bw := NewWriter(ioutil.Discard)
		_ = bw.Flush()
		_ = bw.Flush()
		_, _ = bw.Write(bs)
		_ = bw.Flush()
		_, _ = bw.WriteString(str)
		_ = bw.Flush()
	}
}

func BenchmarkWriterFlush(b *testing.B) {
	b.ReportAllocs()
	bw := NewWriter(ioutil.Discard)
	str := strings.Repeat("x", 50)
	for i := 0; i < b.N; i++ {
		_, _ = bw.WriteString(str)
		_ = bw.Flush()
	}
}
