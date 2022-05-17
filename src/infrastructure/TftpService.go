package infrastructure

import (
	"fmt"
	"github.com/pin/tftp"
	"io"
	"os"
	"time"
)

type TFTPServer struct {
	Server  *tftp.Server
	Address string
	Port    string
}

func NewTFTPServer(address, port string) *TFTPServer {
	return &TFTPServer{
		Server:  tftp.NewServer(TFTPReadHandler, TFTPWriteHandler),
		Address: address,
		Port:    port,
	}
}

var (
	PathsRatio map[string]string
)

func (ts *TFTPServer) Start() {
	go func() {
		ts.Server.SetTimeout(5 * time.Second)
		err := ts.Server.ListenAndServe(fmt.Sprintf("%s:%s", ts.Address, ts.Port))
		if err != nil {
			fmt.Fprintf(os.Stdout, "server: %v\n", err)
			os.Exit(1)
		}
	}()
}

func AddNewPathToTFTPRatio(virtualPath, actualPath string) {
	if len(PathsRatio) == 0 {
		PathsRatio = make(map[string]string)
	}
	PathsRatio[virtualPath] = actualPath
}

// TFTPReadHandler is called when client starts file download from server
func TFTPReadHandler(filename string, rf io.ReaderFrom) error {
	actualPath := PathsRatio[filename]
	if actualPath == "" {
		err := fmt.Errorf("no such file or directory")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	file, err := os.Open(actualPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	n, err := rf.ReadFrom(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	fmt.Printf("%d bytes sent\n", n)
	return nil
}

// TFTPWriteHandler is called when client starts file upload to server
func TFTPWriteHandler(filename string, wt io.WriterTo) error {
	return fmt.Errorf("read-only available")
}
