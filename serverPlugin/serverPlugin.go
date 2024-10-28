package serverPlugin

import (
	"compress/gzip"
	"context"
	"hammy/cacheFunction"
	"hammy/htaccessFunction"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Config holds server configuration settings
type Config struct {
	Port           string        `yaml:"serverPort"`
	HammyVersion   string        `yaml:"hammyVersion"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	MaxHeaderBytes int           `yaml:"max_header_bytes"`
}

// Predefined content types for file extensions
var contentTypes = map[string]string{
	".html": "text/html; charset=utf-8",
	".css":  "text/css",
	".js":   "application/javascript",
	".json": "application/json",
	".xml":  "application/xml",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".svg":  "image/svg+xml",
	".txt":  "text/plain; charset=utf-8",
	".zip":  "application/zip",
}

// LoadConfig loads the server configuration from config.yaml or environment variables
func LoadConfig() Config {
	file, err := os.ReadFile("./config.yaml")
	if err != nil {
		log.Fatalf("Error reading config.yaml file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		log.Fatalf("Error parsing config.yaml file: %v", err)
	}

	if port := os.Getenv("SERVER_PORT"); port != "" {
		config.Port = port
	}
	return config
}

// StartServer initializes and starts the HTTP server
func StartServer() {
	config := LoadConfig()

	log.Println("Webserver is starting...")
	log.Printf("Listening on port :%s", config.Port)
	log.Println("Use ctrl + c to shutdown the server")

	htaccess := htaccessFunction.NewHtaccessPlugin()
	if err := htaccess.LoadHtaccess("/var/www/html/.htaccess"); err != nil {
		log.Printf("Error loading .htaccess file: %v", err)
	}

	server := createServer(config, htaccess)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", server.Addr, err)
		}
	}()

	<-stop
	log.Println("Shutting down the server...")
	shutdownServer(server)
	log.Println("Server exiting")
}

// createServer initializes and returns a new HTTP server
func createServer(config Config, htaccess *htaccessFunction.HtaccessPlugin) *http.Server {
	return &http.Server{
		Addr:           ":" + config.Port,
		Handler:        applyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handleRequest(w, r, htaccess) })),
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		IdleTimeout:    config.IdleTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}
}

// applyMiddleware applies all middlewares in sequence
func applyMiddleware(handler http.Handler) http.Handler {
	return securityHeadersMiddleware(gzipMiddleware(handler))
}

// gzipMiddleware applies gzip compression to the response
func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzr, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// securityHeadersMiddleware sets security headers for each request
func securityHeadersMiddleware(next http.Handler) http.Handler {
	headers := map[string]string{
		"Strict-Transport-Security": "max-age=63072000; includeSubDomains",
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"Content-Security-Policy":   "default-src 'self'; script-src 'self' 'unsafe-eval'; object-src 'none'; style-src 'self' 'unsafe-inline';",
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		next.ServeHTTP(w, r)
	})
}

// handleRequest handles incoming HTTP requests and responds with appropriate content
func handleRequest(w http.ResponseWriter, r *http.Request, htaccess *htaccessFunction.HtaccessPlugin) {
	w.Header().Set("Server", "HAMMY "+LoadConfig().HammyVersion)

	htaccess.ApplyHtaccess(w, r)
	if w.Header().Get("Content-Type") == "text/html; charset=utf-8" {
		return
	}

	if r.URL.Path == "/" {
		redirectToIndex(w, r)
		return
	}

	if !strings.Contains(r.URL.Path, ".") {
		if !redirectToFile(w, r) {
			serveCustomErrorPage(w, http.StatusNotFound, "/var/www/html/404.html", "serverPlugin/pages/hammy-404.html", "404 - File Not Found")
		}
		return
	}

	if cachedResponse, found := cacheFunction.GetFromCache(r.URL.Path); found {
		setContentType(w, r.URL.Path)
		w.Write(cachedResponse)
		return
	}

	serveFile(w, r)
}

func setContentType(w http.ResponseWriter, path string) {
	ext := filepath.Ext(path)
	if contentType, exists := contentTypes[ext]; exists {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
}

func redirectToIndex(w http.ResponseWriter, r *http.Request) {
	extensions := []string{".php", ".html", ".htmlx"}
	for _, ext := range extensions {
		if _, err := os.Stat("/var/www/html/index" + ext); !os.IsNotExist(err) {
			http.Redirect(w, r, "/index"+ext, http.StatusMovedPermanently)
			return
		}
	}
}

func redirectToFile(w http.ResponseWriter, r *http.Request) bool {
	extensions := []string{".php", ".html", ".htmlx", ".jpg", ".png", ".zip", ".css", ".js"}
	for _, ext := range extensions {
		if _, err := os.Stat("/var/www/html" + r.URL.Path + ext); !os.IsNotExist(err) {
			http.Redirect(w, r, r.URL.Path+ext, http.StatusMovedPermanently)
			return true
		}
	}
	return false
}

func serveFile(w http.ResponseWriter, r *http.Request) {
	filePath := "/var/www/html" + r.URL.Path
	if filePath == "/var/www/html/" {
		filePath = "/var/www/html/index.html"
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if isEmptyDir("/var/www/html") {
			log.Println("No content found, and /var/www/html is empty, serving Hammy index")
			http.ServeFile(w, r, "serverPlugin/pages/hammy-index.html")
		} else {
			serveCustomErrorPage(w, http.StatusNotFound, "/var/www/html/404.html", "serverPlugin/pages/hammy-404.html", "404 - File Not Found")
		}
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		serveCustomErrorPage(w, http.StatusNotFound, "/var/www/html/404.html", "serverPlugin/pages/hammy-404.html", "404 - File Not Found")
		return
	}

	if strings.HasSuffix(filePath, ".php") {
		executePHP(w, filePath)
		return
	}

	setContentType(w, filePath)
	w.Write(content)
	cacheFunction.AddToCache(r.URL.Path, content)
}

func serveCustomErrorPage(w http.ResponseWriter, statusCode int, errorPagePath, defaultErrorPage, defaultMsg string) {
	customErrorPage, err := os.ReadFile(errorPagePath)
	if err != nil {
		log.Printf("%d page not found! Serving default Hammy %d", statusCode, statusCode)
		customErrorPage, err = os.ReadFile(defaultErrorPage)
		if err != nil {
			log.Printf("Default Hammy %d page not found! Serving basic message.", statusCode)
			customErrorPage = []byte("<html><body><h1>" + defaultMsg + "</h1><p>An error occurred while processing the request.</p></body></html>")
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(customErrorPage)
}

func executePHP(w http.ResponseWriter, filePath string) {
	cmd := exec.Command("php", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error executing PHP script %s: %v", filePath, err)
		http.Error(w, "Error executing PHP script", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(output)
}

func isEmptyDir(name string) bool {
	dir, err := os.Open(name)
	if err != nil {
		return false
	}
	defer dir.Close()

	_, err = dir.Readdir(1)
	return err == io.EOF
}
func shutdownServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
