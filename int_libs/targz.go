package int_libs

import (
	"compress/gzip"
	"log"
	"fmt"
	"os"
	"io/ioutil"
	"archive/tar"
	"io"
	"path/filepath"
)

func check(e error) {
	if e != nil {
		log.Panicln("[DEBUG]", "error=", e)
	}
}

// extract  .tgz files
func untgz(src, dst_dir string) error {
	fi, err := os.Open(src)
	check(err)
	defer fi.Close()

	zr, err := gzip.NewReader(fi)
	check(err)
	zr.Multistream(false)
	fmt.Printf("Name: %s\nComment: %s\nModTime: %s\n\n", zr.Header.Name, zr.Comment, zr.ModTime.UTC())
	defer zr.Close()

	tr := tar.NewReader(zr)

	// Iterate through the files in the tar archive.
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		check(err)
		path := filepath.Join(dst_dir, hdr.Name)
		info := hdr.FileInfo()
		log.Println("[DEBUG]", "get from archive", hdr.Name)
		if info.IsDir() {
			err = os.MkdirAll(path, info.Mode())
			check(err)
			continue
		}

		fileContents, err := ioutil.ReadAll(tr)
		check(err)
		err = ioutil.WriteFile(dst_dir + hdr.Name, fileContents, 0644)
		check(err)
	}
	return error(nil)
}

// wrapper around func untgz(src, dst_dir string)
// just with functionality to delete src file
func Untgz(src, dst_dir string, del_src bool) {
	err := untgz(src, dst_dir)
	check(err)
	if del_src {
		err := os.Remove(src)
		check(err)
		log.Println("[INFO]", "deleted file=", src)
	}
}


