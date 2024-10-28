package serverPlugin

import (
	"compress/gzip"
	"context"
	"hammy/cacheFunction"
	"hammy/htaccessFunction"
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

// LoadConfig loads the server configuration from a config.yaml file
func LoadConfig() Config {
	file, err := os.ReadFile("./config.yaml")
	if err != nil {
		log.Fatalf("Error reading config.yaml file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		log.Fatalf("Error parsing config.yaml file: %v", err)
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
		Handler:        securityHeadersMiddleware(gzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handleRequest(w, r, htaccess) }))),
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		IdleTimeout:    config.IdleTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}
}

// securityHeadersMiddleware sets security headers for each request
func securityHeadersMiddleware(next http.Handler) http.Handler {
	headers := map[string]string{
		"Strict-Transport-Security": "max-age=63072000; includeSubDomains",
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"Content-Security-Policy":   "default-src 'self'; script-src 'self' 'unsafe-eval'; object-src 'none'; style-src 'self' 'unsafe-inline'; style-src-elem 'self' 'unsafe-inline';",
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		next.ServeHTTP(w, r)
	})
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
			serve404(w, r, r.URL.Path)
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

	if !fileExists(filePath) {
		if isEmptyDir("/var/www/html") {
			log.Printf("No content found, and /var/www/html is empty, serving Hammy index\n")
			http.ServeFile(w, r, "serverPlugin/pages/hammy-index.html")
			return
		}
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		serve404(w, r, filePath)
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

func fileExists(filePath string) bool {
	extensions := []string{".php", ".html", ".htmlx", ".jpg", ".png", ".zip", ".css", ".js"}
	for _, ext := range extensions {
		if _, err := os.Stat(filePath + ext); !os.IsNotExist(err) {
			return true
		}
	}
	return false
}

func isEmptyDir(dir string) bool {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Error accessing %s directory: %v", dir, err)
		return false
	}
	for _, file := range files {
		if !file.IsDir() {
			return false
		}
	}
	return true
}

func serve404(w http.ResponseWriter, r *http.Request, filePath string) {
	log.Printf("File not found: Path=%s, URL=%s\n", filePath, r.URL.Path)

	customErrorPage, err := os.ReadFile("/var/www/html/404.html")
	if err != nil {
		log.Printf("404 page not found! Serving default Hammy 404")
		customErrorPage, err = os.ReadFile("serverPlugin/pages/hammy-404.html")
		if err != nil {
			log.Printf("Default Hammy 404 page not found! Serving basic 404 message.")
			customErrorPage = []byte("<html><body><h1>404 - File Not Found</h1><p>The requested file could not be found.</p></body></html>")
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write(customErrorPage)
}

func executePHP(w http.ResponseWriter, filePath string) {
	cmd := exec.Command("php", filePath)
	output, err := cmd.Output()
	if err != nil {
		handlePHPError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(output)
}

func handlePHPError(w http.ResponseWriter, err error) {
	if strings.Contains(err.Error(), "exec: \"php\": executable file not found in $PATH") {
		log.Printf("PHP execution error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error executing PHP file! PHP is not installed!"))
		return
	}
	log.Printf("Error executing PHP file: %v", err)
	w.WriteHeader(http.StatusInternalServerError)
	serve500(w)
}

func serve500(w http.ResponseWriter) {
	hammy500Page, err := os.ReadFile("/var/www/html/500.html")
	if err != nil {
		log.Printf("500 page not found! Serving default Hammy 500")
		hammy500Page, err = os.ReadFile("serverPlugin/pages/hammy-500.html")
		if err != nil {
			log.Printf("Default Hammy 500 page not found! Serving basic 500 message.")
			hammy500Page = []byte("<html><body><h1>500 - Internal Server Error</h1><p>An unexpected condition was encountered.</p></body></html>")
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(hammy500Page)
}

func serve505(w http.ResponseWriter) {
	hammy505Page, err := os.ReadFile("/var/www/html/505.html")
	if err != nil {
		log.Printf("505 page not found! Serving default Hammy 505")
		hammy505Page, err = os.ReadFile("serverPlugin/pages/hammy-505.html")
		if err != nil {
			log.Printf("Default Hammy 505 page not found! Serving basic 505 message.")
			hammy505Page = []byte("<html><body><h1>505 - HTTP Version Not Supported</h1><p>The server does not support the HTTP protocol version used in the request.</p></body></html>")
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusHTTPVersionNotSupported)
	if _, err := w.Write(hammy505Page); err != nil {
		log.Printf("Error writing 505 response: %v", err)
	}
}

// setContentType sets the Content-Type header based on the file extension
func setContentType(w http.ResponseWriter, filePath string) {
	ext := strings.ToLower(filepath.Ext(filePath))
	contentTypes := map[string]string{
		".html":  "text/html; charset=utf-8",
		".htmlx": "text/html; charset=utf-8",
		".js":    "application/javascript",
		".jpg":   "image/jpeg",
		".jpeg":  "image/jpeg",
		".png":   "image/png",
		".zip":   "application/zip",
		".css":   "text/css",
	}
	if contentType, found := contentTypes[ext]; found {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
}

// shutdownServer attempts a graceful shutdown of the server
func shutdownServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
