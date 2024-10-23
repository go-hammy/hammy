package htaccessFunction

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"
)

// HtaccessRule represents a single .htaccess rule
type HtaccessRule struct {
	Pattern string
	Action  string
}

// HtaccessPlugin manages .htaccess rules
type HtaccessPlugin struct {
	Rules []HtaccessRule
}

// LoadHtaccess loads .htaccess rules from a file
func (p *HtaccessPlugin) LoadHtaccess(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			log.Printf("Invalid htaccess rule: %s", line)
			continue
		}

		p.Rules = append(p.Rules, HtaccessRule{
			Pattern: parts[0],
			Action:  parts[1],
		})
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// ApplyHtaccess applies the loaded .htaccess rules to an HTTP request
func (p *HtaccessPlugin) ApplyHtaccess(w http.ResponseWriter, r *http.Request) {
	for _, rule := range p.Rules {
		if strings.Contains(r.URL.Path, rule.Pattern) {
			switch rule.Action {
			case "deny":
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("403 - Forbidden"))
				return
			case "allow":
				// Allow the request to proceed
				return
			default:
				log.Printf("Unknown htaccess action: %s", rule.Action)
			}
		}
	}
}

// NewHtaccessPlugin creates a new instance of HtaccessPlugin
func NewHtaccessPlugin() *HtaccessPlugin {
	return &HtaccessPlugin{
		Rules: []HtaccessRule{},
	}
}
