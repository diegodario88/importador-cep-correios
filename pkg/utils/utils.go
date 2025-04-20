package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
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
	return strings.ReplaceAll(fmt.Sprintf("%d", n), "", ".")
}

func EmptyToNull(field string) any {
	if field == "" {
		return nil
	}
	return field
}
