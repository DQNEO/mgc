package ioutil

import "os"

const MYBUFSIZ = 65536 * 2

func ReadFile(filenameAsString string) ([]byte, error) {
	var buf []byte
	buf = makeSlice(MYBUFSIZ, MYBUFSIZ, 24)
	var err error

	// Currently, there is no way to declare type of other package, so Let it infer
	f := &os.File{}

	f, err = os.Open(filenameAsString)

	var n int = MYBUFSIZ

	var nread int
	nread = read(f.innerFile.id, buf, n)
	var buf2 []byte
	buf2 = buf[0:nread:nread]
	return buf2,nil
}
