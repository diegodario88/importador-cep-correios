package utils

import (
	"log"
	"os"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func GetCWD() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return wd
}

func FormatNumber(n int) string {
	p := message.NewPrinter(language.BrazilianPortuguese)
	return p.Sprintf("%d", n)
}

func HandleEmpty(field string, fileName string) any {
	if fileName == "ECT_PAIS.TXT" && field == "" {
		return field
	}

	if field == "" {
		return nil
	}

	return strings.TrimSpace(field)
}
