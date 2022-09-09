// The task is to implement a command that is going to be poor-man-clone of the wget command.
// It should take 1 argument - the url of a file to download, and save it to local disk.

// Each 1 second it should log current progress so far in bytes.

// Example usage:

// $ gowget http://releases.ubuntu.com/18.04.3/ubuntu-18.04.3-desktop-amd64.iso
// Downloaded 123 bytes ...
// Downloaded 456 bytes ...
// Downloaded 789 bytes ...

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

func main() {
	resp, err := http.Get("http://releases.ubuntu.com/18.04.3/ubuntu-18.04.3-desktop-amd64.iso")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		panic(resp.StatusCode)
	}
	f, err := os.Create("result.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = io.Copy(f, NewLoggingReader(resp.Body))
	if err != nil {
		panic(err)
	}
	fmt.Println("done")
}

type LoggingReader struct {
	r      io.Reader
	amount int64
}

func NewLoggingReader(r io.Reader) io.Reader {
	lr := &LoggingReader{
		r: r,
	}
	go lr.printLoop()
	return lr
}

func (r *LoggingReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	atomic.AddInt64(&r.amount, int64(n))
	return n, err
}

func (r *LoggingReader) printLoop() {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for range t.C {
		fmt.Printf("Downloaded %d bytes ...\n", atomic.LoadInt64(&r.amount))
	}
}
