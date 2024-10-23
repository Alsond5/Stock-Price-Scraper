package dotenv

import (
	"bufio"
	"os"
	"strings"
)

func LoadEnv() error {
	file, err := os.Open(".env")
	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		pair := strings.SplitN(line, "=", 2)
		if len(pair) == 2 {
			os.Setenv(pair[0], pair[1])
		}

		key := strings.TrimSpace(pair[0])
		value := strings.TrimSpace(pair[1])
		os.Setenv(key, value)
	}

	return nil
}
