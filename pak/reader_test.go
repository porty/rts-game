package pak

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListFilesEmpty(t *testing.T) {
	r := NewReader(bytes.NewReader([]byte{}))

	records, err := r.ListFiles()

	require.EqualError(t, err, "failed to read file offset: EOF")
	require.Nil(t, records)
}

func TestListFilesNoFiles(t *testing.T) {
	r := NewReader(bytes.NewReader([]byte{0, 0, 0, 0}))

	records, err := r.ListFiles()

	require.NoError(t, err)
	require.Equal(t, 0, len(records))
}

func TestListFiles(t *testing.T) {
	contents := []byte{
		0xAA, 0x02, 0, 0, 'H', 'e', 'l', 'l', 'o', '.', 'V', 'O', 'C', 0,
		0x8C, 0x4E, 0, 0, 'T', 'H', 'I', 'N', 'G', '.', 'V', 'O', 'C', 0,
		0, 0, 0, 0,
	}
	r := NewReader(bytes.NewReader(contents))

	records, err := r.ListFiles()

	require.NoError(t, err)
	require.Equal(t, []Record{
		{
			Filename: "Hello.VOC",
			Offset:   0x02AA,
		},
		{
			Filename: "THING.VOC",
			Offset:   0x4E8C,
		},
	}, records)
}

func TestListFilesOutOfOrder(t *testing.T) {
	contents := []byte{
		0x8C, 0x4E, 0, 0, 'T', 'H', 'I', 'N', 'G', '.', 'V', 'O', 'C', 0,
		0xAA, 0x02, 0, 0, 'H', 'e', 'l', 'l', 'o', '.', 'V', 'O', 'C', 0,
		0, 0, 0, 0,
	}
	r := NewReader(bytes.NewReader(contents))

	records, err := r.ListFiles()

	require.EqualError(t, err, "files seem out of order: last offset = 00004e8c, this offset = 000002aa")
	require.Nil(t, records)
}

func TestExtract(t *testing.T) {
	dir, err := ioutil.TempDir("", "paktest")
	if err != nil {
		panic("failed to get temporary directory for testing: " + err.Error())
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			log.Printf("There might still be garbage in %q: %s", dir, err.Error())
		}
	}()
	contents := []byte{
		32, 0, 0, 0, 'T', 'H', 'I', 'N', 'G', '.', 'T', 'X', 'T', 0,
		39, 0, 0, 0, 'H', 'e', 'l', 'l', 'o', '.', 'V', 'O', 'C', 0,
		0, 0, 0, 0,
		't', 'h', 'i', 'n', 'g', '!', '\n',
		'h', 'e', 'l', 'l', 'o', '!', '\n',
	}
	r := NewReader(bytes.NewReader(contents))

	err = r.ExtractFiles(dir)

	require.NoError(t, err)
	infos, err := ioutil.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, infos, 2)
	b, err := ioutil.ReadFile(path.Join(dir, "THING.TXT"))
	require.NoError(t, err)
	require.Equal(t, []byte("thing!\n"), b)
	b, err = ioutil.ReadFile(path.Join(dir, "Hello.VOC"))
	require.NoError(t, err)
	require.Equal(t, []byte("hello!\n"), b)
}
