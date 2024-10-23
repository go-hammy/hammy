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
	Port           string        `yaml:"server_port"`
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
	err = yaml.Unmarshal(file, &config)
	if err != nil {
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

	// Load htaccess rules once at startup
	htaccess := htaccessFunction.NewHtaccessPlugin()
	err := htaccess.LoadHtaccess("/var/www/html/.htaccess")
	if err != nil {
		log.Printf("Error loading .htaccess file: %v", err)
	}

	server := createServer(config, htaccess)

	// Channel to listen for interrupt or terminate signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Run server in a goroutine
	go func() {
		startServer(server)
	}()

	// Block until a signal is received
	<-stop

	log.Println("Shutting down the server...")

	// Attempt graceful shutdown
	shutdownServer(server)

	log.Println("Server exiting")
}

// createServer initializes and returns a new HTTP server
func createServer(config Config, htaccess *htaccessFunction.HtaccessPlugin) *http.Server {
	return &http.Server{
		Addr:           ":" + config.Port,
		Handler:        securityHeadersMiddleware(gzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleRequest(w, r, htaccess)
		}))),
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		IdleTimeout:    config.IdleTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}
}

// securityHeadersMiddleware sets security headers for each request
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-eval'; object-src 'none'; style-src 'self' 'unsafe-inline'; style-src-elem 'self' 'unsafe-inline';")
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
	w.Header().Set("Server", "HAMMY v1.0")

	// Apply htaccess rules
	htaccess.ApplyHtaccess(w, r)
	if w.Header().Get("Content-Type") == "text/html; charset=utf-8" {
		return
	}

	// Redirect to the appropriate index file if the root is accessed
	if r.URL.Path == "/" {
		extensions := []string{".php", ".html", ".htmlx"}
		for _, ext := range extensions {
			if _, err := os.Stat("/var/www/html/index" + ext); !os.IsNotExist(err) {
				http.Redirect(w, r, "/index"+ext, http.StatusMovedPermanently)
				return
			}
		}
	}

	// Redirect to corresponding file if a path without extension is accessed
	if !strings.Contains(r.URL.Path, ".") {
		extensions := []string{".php", ".html", ".htmlx", ".jpg", ".png", ".zip", ".css", ".js"}
		for _, ext := range extensions {
			if _, err := os.Stat("/var/www/html" + r.URL.Path + ext); !os.IsNotExist(err) {
				http.Redirect(w, r, r.URL.Path+ext, http.StatusMovedPermanently)
				return
			}
		}
	}

	// Check if the response is cached
	if cachedResponse, found := cacheFunction.GetFromCache(r.URL.Path); found {
		setContentType(w, r.URL.Path)
		w.Write(cachedResponse)
		return
	}

	// Serve the requested file
	filePath := "/var/www/html" + r.URL.Path
	if filePath == "/var/www/html/" {
		filePath = "/var/www/html/index.html"
	}

	extensions := []string{".php", ".html", ".htmlx", ".jpg", ".png", ".zip", ".css", ".js"}
	fileExists := false
	for _, ext := range extensions {
		if _, err := os.Stat(filePath + ext); !os.IsNotExist(err) {
			filePath += ext
			fileExists = true
			break
		}
	}

	if !fileExists {
		log.Printf("No content found for Path=%s, serving Hammy index\n", r.URL.Path)
		http.ServeFile(w, r, "serverPlugin/pages/hammy-index.html")
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
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
		return
	}

	// If the file is a PHP file, execute it
	if strings.HasSuffix(filePath, ".php") {
		cmd := exec.Command("php", filePath)
		output, err := cmd.Output()
		if err != nil {
			if strings.Contains(err.Error(), "exec: \"php\": executable file not found in $PATH") {
				log.Printf("PHP execution error: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error executing PHP file! PHP is not installed!"))
				return
			}
			log.Printf("Error executing PHP file: %v", err)
			w.WriteHeader(http.StatusInternalServerError)

			hammy500Page, err := os.ReadFile("serverPlugin/pages/hammy-500.html")
			if err != nil {
				log.Printf("Hammy 500 page not found! Serving basic 500 message.")
				hammy500Page = []byte("<html><body><h1>500 - Internal Server Error</h1><p>Something went wrong on our end.</p></body></html>")
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(hammy500Page)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(output)
		return
	}

	// Set the appropriate content type for the file
	setContentType(w, filePath)

	// Write content to response
	w.Write(content)

	// Cache the response
	cacheFunction.AddToCache(r.URL.Path, content)
}

// setContentType sets the Content-Type header based on the file extension
func setContentType(w http.ResponseWriter, filePath string) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".html", ".htmlx":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".zip":
		w.Header().Set("Content-Type", "application/zip")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
}

// startServer starts the HTTP server and logs any errors
func startServer(server *http.Server) {
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", server.Addr, err)
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
