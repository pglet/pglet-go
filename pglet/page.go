package pglet

import (
	"fmt"
	"os"
)

type Page struct {
	Name string
}

func NewPage(name string) *Page {
	err := install()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return &Page{
		Name: name,
	}
}
