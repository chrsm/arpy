package arpy

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/MacIt/pickle"
)

type Version int

const (
	RPA32 Version = 4
	RPA3  Version = 3
	RPA2  Version = 2
)

var (
	mRPA32 = []byte("RPA-3.2 ")
	mRPA3  = []byte("RPA-3.0 ")
	mRPA2  = []byte("RPA-2.0 ")

	ErrUnsupportedVersion = errors.New("unsupported RPA version")
	ErrInvalidMetadata    = errors.New("archive missing metadata")
)

type RPA struct {
	r io.ReadSeeker

	Version Version
	Key     int64

	Indexes []Index
	files   [][]byte
}

type Index struct {
	Name         string
	Offset, Size int64
}

func (r *RPA) FileAt(idx Index) ([]byte, error) {
	r.r.Seek(idx.Offset, io.SeekStart)

	buf := make([]byte, idx.Size)
	r.r.Read(buf)

	return buf, nil
}

func New(key int64) *RPA {
	return &RPA{
		Key: key,
	}
}

func (r *RPA) AddFile(name string, src []byte) {
	// padding? what is this, a bed?
	r.Indexes = append(r.Indexes, Index{
		Name:   name,
		Offset: 0, // fixup at write
		Size:   int64(len(src)) ^ r.Key,
	})

	r.files = append(r.files, src)
}

func (r *RPA) WriteTo(w *os.File) (int, error) {
	// i don't know the 3.2 format well enough because i'm LAZY
	// "SUE ME" - jk
	// actually please don't - i don't have enough money for that
	// also, as part of using this library you agree to not sue me nor hold me
	// accountable for any damages while using this software blah blah blah ok?
	// this includes, but is not limited to: hair loss, developing negative feelings
	// towards python, throwing your laptop as well as fleeing the country
	w.Write(mRPA3)

	if len(r.Indexes) == 0 {
		panic("RPA with no content being written. and i took offense to that.")
	}

	w.WriteString("0000000000000000 ")
	w.WriteString(fmt.Sprintf("%08x", r.Key))
	w.Write([]byte{'\n'})

	for i := range r.Indexes {
		src := r.files[i]

		r.Indexes[i].Offset, _ = w.Seek(0, io.SeekCurrent)
		r.Indexes[i].Offset ^= r.Key
		w.Write(src)
	}

	ofsidx, _ := w.Seek(0, io.SeekCurrent)
	w.Seek(int64(len(mRPA3)), io.SeekStart)
	w.WriteString(fmt.Sprintf("%016x ", ofsidx))
	w.Seek(ofsidx, io.SeekStart) // am i dumb? (probably)

	// w/pickle -> zlib -> file, r/file -> zlib -> pickle
	b := new(bytes.Buffer)
	cucumber := pickle.NewEncoder(b)

	dat := make(map[interface{}]interface{})
	for i := range r.Indexes {
		cur := r.Indexes[i]
		dat[cur.Name] = []interface{}{
			pickle.Tuple{
				big.NewInt(cur.Offset),
				big.NewInt(cur.Size),
			},
		}
	}

	//        haha, i'm so clever `/s`
	if err := cucumber.Encode(dat); err != nil {
		panic(err)
	}

	z := zlib.NewWriter(w)
	z.Write(b.Bytes())
	z.Close()

	return 0, nil
}

func Decode(r *os.File) (*RPA, error) {
	// i don't check errors just like cool guys don't look at explosions
	r.Seek(0, io.SeekStart)

	buf := new(bytes.Buffer)
	rbuf := make([]byte, 8)
	r.Read(rbuf)

	pak := &RPA{r: r}
	switch {
	case bytes.Contains(rbuf, mRPA32):
		pak.Version = RPA32
	case bytes.Contains(rbuf, mRPA3):
		pak.Version = RPA3
	case bytes.Contains(rbuf, mRPA2):
		pak.Version = RPA2
	default:
		return nil, ErrUnsupportedVersion
	}

	nl := []byte{0x0A}
	idx := -1
	for {
		n, err := r.Read(rbuf)
		buf.Write(rbuf[:n])

		idx = bytes.Index(buf.Bytes(), nl)
		if idx > -1 {
			break
		}

		if err == io.EOF {
			break
		}
	}

	if idx <= -1 {
		return nil, ErrInvalidMetadata
	}

	md := strings.Split(string(buf.Bytes()[:idx]), " ")
	ofs, _ := strconv.ParseInt(md[0], 16, 64)

	switch pak.Version {
	case RPA3:
		pak.Key = 0

		pak.Key, _ = strconv.ParseInt(md[1], 16, 64)
	case RPA32:
		pak.Key, _ = strconv.ParseInt(md[2], 16, 64)
	}

	r.Seek(ofs, io.SeekStart)

	zr, err := zlib.NewReader(r)
	if err != nil {
		panic("zlib err: " + err.Error()) // your problem
	}

	cbuf := new(bytes.Buffer)
	cbuf.ReadFrom(zr)
	zr.Close()

	pdec := pickle.NewDecoder(cbuf)
	pidx, err := pdec.Decode()
	if err != nil {
		return nil, err // your problem
	}

	// expecting these to be very particular which might turn out to be wrong later
	iidx := pidx.(map[interface{}]interface{})
	for k, v := range iidx {
		// pickle.Tuple is an []interface{} in a fancy hat
		val := (v.([]interface{})[0]).(pickle.Tuple)

		var ofs, sz int64

		switch nv := val[0].(type) {
		case *big.Int:
			ofs = nv.Int64()
		case int64:
			ofs = nv	
		}

		switch nv := val[1].(type) {
		case *big.Int:
			sz = nv.Int64()
		case int64:
			sz = nv
		}

		pak.Indexes = append(pak.Indexes, Index{
			Name:   k.(string),
			Offset: ofs ^ int64(pak.Key), // lol
			Size:   sz ^ int64(pak.Key),  // lol
		})
	}

	return pak, nil
}
