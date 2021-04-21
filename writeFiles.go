package main

import (
	"os"
)

type WriteFileStruct struct {
	Data  string
	Where string
}

func Write2File(data WriteFileStruct) error {
	var (
		f   *os.File
		err error
	)
	if f, err = os.OpenFile(data.Where, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.WriteString(data.Data + "\n"); err != nil {
		return err
	}
	return nil
}
