package forwarder

import (
	"bufio"
	"os"
	"strings"

	"github.com/mdxabu/bridge/internal/config"
	"github.com/mdxabu/bridge/internal/logger"
)


func ReadIPv4Addresses() ([]string, error) {
	cfg, err := config.ParseConfig()
	if err != nil {
		return nil, err
	}

	filePath, err := cfg.GetDestIPPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ipAddresses []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}

		ipAddresses = append(ipAddresses, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	logger.Info("Read %d IPv4 addresses from %s", len(ipAddresses), filePath)
	return ipAddresses, nil
}
