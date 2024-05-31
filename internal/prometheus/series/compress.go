package series

import (
	"fmt"
	"io"
	"math"
	"time"
)

type compressingSeriesWriter struct {
	w     io.Writer
	base  int
	count int
	v0    int
	v1    int
}

func newCompressingSeriesWriter(w io.Writer) *compressingSeriesWriter {
	return &compressingSeriesWriter{
		w:     w,
		base:  0,
		count: -3,
		v0:    0,
		v1:    math.MaxInt,
	}
}

func (w *compressingSeriesWriter) writerFunc() func(int, time.Time) error {
	return func(v int, _ time.Time) error {
		if w.v1-w.v0 == v-w.v1 {
			// 1  5  9 => 1+4x2
			// v0 v1 v
			if w.count < 1 {
				w.base = w.v0
				w.count = 1
			} else {
				w.count++
			}
		} else {
			switch {
			case w.count < -1:
				// count == -3  count == -2
				// x  x  1      x  1  4
				// v0 v1 v      v0 v1 v
				w.count++
			case w.count == -1 || w.count == 0:
				// count == 0
				// 1  5  8
				// v0 v1 v
				// --> print "1" (v0)
				if err := w.emitSingle(w.v0); err != nil {
					return err
				}
			case w.count > 0:
				// count == 1
				// 1  5  9  2
				//    v0 v1 v
				// base
				// --> print "1+4x2"
				if err := w.emitExpanding(); err != nil {
					return err
				}
				w.count = -2
				w.v1 = math.MaxInt
			}
		}
		w.v0 = w.v1
		w.v1 = v
		return nil
	}
}

func (w *compressingSeriesWriter) emitSingle(v int) error {
	_, err := fmt.Fprintf(w.w, "%d ", v)
	return err
}

func (w *compressingSeriesWriter) emitExpanding() error {
	d := w.v1 - w.v0
	var err error
	if d == 0 {
		_, err = fmt.Fprintf(w.w, "%dx%d ", w.base, w.count+1)
	} else {
		_, err = fmt.Fprintf(w.w, "%d%+dx%d ", w.base, w.v1-w.v0, w.count+1)
	}
	return err
}

func (w *compressingSeriesWriter) Close() error {
	t := time.UnixMicro(0)
	f := w.writerFunc()
	f(math.MinInt, t)
	f(0, t)
	return nil
}
