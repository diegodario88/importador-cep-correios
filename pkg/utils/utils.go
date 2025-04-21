package utils

import (
	"bufio"
	"log"
	"os"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func CountLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}

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
