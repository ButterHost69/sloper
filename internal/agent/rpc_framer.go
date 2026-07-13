package agent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
)

// frameWriter writes JSON objects followed by '\n' to an io.Writer.
// It intentionally uses json.Encoder (which appends '\n') rather than
// building manual byte slices, to avoid escaping mistakes.
type frameWriter struct {
	enc *json.Encoder
	w   io.Writer // kept for Close check
}

func newFrameWriter(w io.Writer) *frameWriter {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return &frameWriter{enc: enc, w: w}
}

// Write marshals v as JSON and writes it followed by a newline.
func (fw *frameWriter) Write(v any) error { return fw.enc.Encode(v) }

// frameReader reads JSONL frames from an io.Reader.
//
// IMPORTANT: we do NOT use bufio.Scanner because it splits on U+2028 and
// U+2029, which are valid inside JSON strings.  The pi RPC docs explicitly
// warn against this.  Instead we read raw bytes up to the next '\n'.
type frameReader struct {
	rd *bufio.Reader
}

func newFrameReader(r io.Reader) *frameReader {
	return &frameReader{rd: bufio.NewReader(r)}
}

// ReadLine returns the next full line (without trailing \r or \n).
// Empty lines are skipped automatically.
func (fr *frameReader) ReadLine() ([]byte, error) {
	for {
		line, err := fr.rd.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		line = bytes.TrimRight(line, "\r\n")
		if len(line) > 0 {
			return line, nil
		}
		// empty line → skip and try next
	}
}

// Decode reads one JSONL line and unmarshals it into dst.
func (fr *frameReader) Decode(dst any) error {
	line, err := fr.ReadLine()
	if err != nil {
		return err
	}
	return json.Unmarshal(line, dst)
}
