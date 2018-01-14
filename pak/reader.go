package pak

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

type Reader struct {
	r io.ReadSeeker
}

func NewReader(r io.ReadSeeker) *Reader {
	return &Reader{
		r: r,
	}
}

type Record struct {
	Offset   uint32
	Filename string
}

func (r *Reader) ListFiles() ([]Record, error) {
	var records []Record
	var lastOffset uint32
	br := bufio.NewReader(r.r)
	for {
		var offset uint32
		if err := binary.Read(br, binary.LittleEndian, &offset); err != nil {
			return nil, errors.New("failed to read file offset: " + err.Error())
		}
		if offset == 0 {
			break
		}
		if offset < lastOffset {
			return nil, fmt.Errorf("files seem out of order: last offset = %08x, this offset = %08x", lastOffset, offset)
		}
		lastOffset = offset
		var chars []byte
		for {
			if b, err := br.ReadByte(); err != nil {
				return nil, errors.New("failed to read file name: " + err.Error())
			} else if b != 0 {
				// DOS 8.3 filenames only
				if len(chars) >= 12 {
					return nil, errors.New("filename too long")
				}
				chars = append(chars, b)
				continue
			}
			records = append(records, Record{
				Filename: string(chars),
				Offset:   offset,
			})
			break
		}
	}
	return records, nil
}

func (r *Reader) ExtractFiles(dir string) error {
	if fi, err := os.Stat(dir); err != nil {
		return errors.New("failed to get output directory information: " + err.Error())
	} else if !fi.IsDir() {
		return errors.New("output path is not a directory")
	}

	records, err := r.ListFiles()
	if err != nil {
		return err
	}

	end, err := r.r.Seek(0, os.SEEK_END)
	if err != nil {
		return errors.New("failed to get length of file: " + err.Error())
	}

	for i, rec := range records {
		if _, err := r.r.Seek(int64(rec.Offset), os.SEEK_SET); err != nil {
			return errors.New("failed to seek to file offset: " + err.Error())
		}
		var size uint32
		if i+1 < len(records) {
			size = records[i+1].Offset - rec.Offset
		} else {
			size = uint32(end) - rec.Offset
		}
		extractPath := path.Join(dir, rec.Filename)
		if err := saveToFile(r.r, int64(size), extractPath); err != nil {
			return fmt.Errorf("error extracting to %q: %s", extractPath, err.Error())
		}
	}
	return nil
}

func saveToFile(r io.Reader, size int64, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, io.LimitReader(r, size))
	return err
}
