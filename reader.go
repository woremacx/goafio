package afio

import (
	"bytes"
	"errors"
	"io"
	"strconv"
)

type Reader struct {
	rd          io.Reader
	remaining   int
	Pos         int64
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		rd: r,
	}
}

var (
	ErrInvalidHeader = errors.New("Did not find valid magic number")
)

func (r *Reader) Next() (*Header, error) {
	if r.remaining > 0 {
		if err := r.skip(r.remaining); err != nil {
			return nil, err
		}
	}

	h, err := readHeader(r.rd)
	if err != nil {
		return nil, err
	}

	r.remaining = int(h.Size)
	h.Offset = r.Pos
	h.AllSize = h.Consumed + h.Size
	r.Pos += h.Consumed

	return h, nil
}

func (r *Reader) Read(b []byte) (int, error) {
	if r.remaining == 0 {
		return 0, io.EOF
	}

	if len(b) > r.remaining {
		b = b[:r.remaining]
	}

	n, err := r.rd.Read(b)

	if n > 0 {
		r.remaining -= n
		r.Pos += int64(n)
	}

	return n, err
}

func readHeader(rd io.Reader) (*Header, error) {
	magic := make([]byte, 6)
	if _, err := io.ReadFull(rd, magic); err != nil {
		return nil, err
	}

	if bytes.Equal(magic, []byte("070707")) {
		hdr, err := readHeaderAfioOld(rd)
		return hdr, err
	}
/*
	} else if bytes.Equal(magic, []byte("070717")) {
		hdr, err := readHeaderAfioExtended(rd)
		return hdr, err
	} else if bytes.Equal(magic, []byte("070727")) {
		hdr, err := readHeaderAfioLarge(rd)
		return hdr, err
	}
*/

	return nil, ErrInvalidHeader
}

// #define PH_SCAN "%6lo%6lo%6lo%6lo%6lo%6lo%6lo%11lo%6o%11lo"
// &pasb.PSt_dev, &pasb.PSt_ino, &pasb.PSt_mode, &pasb.PSt_uid, &pasb.PSt_gid, &pasb.PSt_nlink, &pasb.PSt_rdev, &pasb.PSt_mtime, &namelen, &pasb.PSt_size

func readHeaderAfioOld(rd io.Reader) (*Header, error) {
	headerLen := 70
	b := make([]byte, headerLen)
	if _, err := io.ReadFull(rd, b); err != nil {
		return nil, err
	}

	h := Header{}
	h.Consumed = int64(headerLen + 6)

	mode, err := strconv.ParseInt(string(b[12:18]), 8, 64)
	if err != nil {
		return nil, err
	}
	h.Mode = mode

	uid, err := strconv.ParseInt(string(b[18:24]), 8, 64)
	if err != nil {
		return nil, err
	}
	h.Uid = int(uid)

	gid, err := strconv.ParseInt(string(b[24:30]), 8, 64)
	if err != nil {
		return nil, err
	}
	h.Gid = int(gid)

	mtime, err := strconv.ParseInt(string(b[42:52]), 8, 64)
	if err != nil {
		return nil, err
	}
	h.Mtime = mtime

	size, err := strconv.ParseInt(string(b[59:70]), 8, 64)
	if err != nil {
		return nil, err
	}
	h.Size = size

	l, err := strconv.ParseInt(string(b[53:59]), 8, 64)
	if err != nil {
		return nil, err
	}

	name := make([]byte, l)
	if _, err := io.ReadFull(rd, name); err != nil {
		return nil, err
	}
	// skip the trailing "0" (?)
	h.Name = string(name[:len(name)-1])

	h.Consumed += int64(l)

	return &h, nil
}

func (r *Reader) skip(n int) error {
	c := 0
	for c < n {
		buf := make([]byte, n-c)

		nr, err := r.rd.Read(buf)
		if err != nil {
			return err
		}

		c += nr
		r.Pos += int64(nr)
	}

	return nil
}

