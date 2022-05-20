package infrastructure

import (
	"fmt"
	"github.com/pin/tftp"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

type TFTPServer struct {
	Server     *tftp.Server
	Address    string
	Port       string
	PathsRatio map[string]string
	Logger     *logrus.Logger
}

func NewTFTPServer(address, port string, log *logrus.Logger) *TFTPServer {
	pathsRatio := make(map[string]string)
	return &TFTPServer{
		Address:    address,
		Port:       port,
		PathsRatio: pathsRatio,
		Logger:     log,
	}
}

func (t *TFTPServer) Start() {
	t.Server = tftp.NewServer(
		func(filename string, rf io.ReaderFrom) error {
			actualPath := t.PathsRatio[filename]
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
		}, nil)
	go func() {
		t.Server.SetTimeout(5 * time.Second)
		err := t.Server.ListenAndServe(fmt.Sprintf("%s:%s", t.Address, t.Port))
		if err != nil {
			fmt.Fprintf(os.Stdout, "server: %v\n", err)
			os.Exit(1)
		}
	}()
}

func (t *TFTPServer) AddNewPathToTFTPRatio(virtualPath, actualPath string) {
	t.PathsRatio[virtualPath] = actualPath
}
