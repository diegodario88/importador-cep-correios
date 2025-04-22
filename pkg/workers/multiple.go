package workers

import (
	"fmt"
	"path/filepath"
	"sync"
)

func Multiple(filePattern string, tools JobTools) {
	var wg sync.WaitGroup

	counter := Counter{
		Increment: 1,
		Error:     nil,
	}

	matches, err := filepath.Glob(filepath.Join(tools.BasePath, filePattern))
	if err != nil {
		counter.Error = fmt.Errorf("erro ao buscar arquivos: %w", err)
		tools.CounterChan <- counter
		return
	}

	ufs := len(matches)
	if ufs == 0 {
		counter.Error = fmt.Errorf("padrão %s não encontrou arquivos", filePattern)
		tools.CounterChan <- counter
		return
	}

	wg.Add(ufs)

	for _, filePath := range matches {
		fileName := filepath.Base(filePath)
		go func(fileName string) {
			defer wg.Done()
			Single(fileName, tools)
		}(fileName)
	}

	wg.Wait()
}
