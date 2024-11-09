package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
)

var (
	GitVersion string
	GitCommit  string
	BuildDate  string
)

var (
	// ErrNoEnv is returned when the configuration is invalid
	ErrNoEnv = errors.New("no environment variables found")
)

// flag custom type to process custom header
type customHeaders []string

func main() {
	// command line arguments to set environment variables prefix
	// e.g. -env-prefix=APP_
	var envPrefix string
	flag.StringVar(&envPrefix, "env-prefix", "APP_", "environment variables prefix to look")

	// base path for directory to serve
	// e.g. -base-path=data
	var basePath string
	flag.StringVar(&basePath, "base-path", "/", "base url path to serve")

	// directory to serve
	// e.g. -dir=data
	var dir string
	flag.StringVar(&dir, "dir", ".", "bundle directory to serve")

	// address to listen on
	// e.g. -address=:8080
	var address string
	flag.StringVar(&address, "address", ":8080", "address to listen on")

	// read timeout in seconds
	// e.g. -read-timeout=10
	var readTimeout int
	flag.IntVar(&readTimeout, "read-timeout", 15, "read timeout in seconds")

	// write timeout in seconds
	// e.g. -write-timeout=10
	var writeTimeout int
	flag.IntVar(&writeTimeout, "write-timeout", 15, "write timeout in seconds")

	// security ContentSecurityPolicy
	// e.g. -security-content-security-policy
	var contentSecurityPolicy string
	flag.StringVar(&contentSecurityPolicy, "security-content-security-policy", "", "ContentSecurityPolicy sets the `Content-Security-Policy` header providing security against cross-site scripting (XSS), clickjacking and other code injection attacks")

	// security XSS Protection
	// e.g. -security-xss-protection
	var xssProtection string
	flag.StringVar(&xssProtection, "security-xss-protection", "1; mode=block", "XSSProtection provides protection against cross-site scripting attack (XSS)")

	// security ContentTypeNosniff
	// e.g. -security-contenttype-nosniff
	var contentTypeNosniff string
	flag.StringVar(&contentTypeNosniff, "security-contenttype-nosniff", "nosniff", "ContentTypeNosniff provides protection against overriding Content-Type")

	// security XFrameOptions
	// e.g. -security-xframe-options
	var xFrameOptions string
	flag.StringVar(&xFrameOptions, "security-xframe-options", "SAMEORIGIN", "XFrameOptions can be used to indicate whether or not a browser should be allowed to render a page in a <frame>, <iframe> or <object>")

	// security HSTSMaxAge
	// e.g. -security-hsts-maxage
	var hstsMaxAge int
	flag.IntVar(&hstsMaxAge, "security-hsts-maxage", 0, "HSTSMaxAge sets the `Strict-Transport-Security` header to indicate how long (in seconds) browsers should remember that this site is only to be accessed using HTTPS.")

	// security HSTSExcludeSubdomains
	// e.g. -security-hsts-excluede-subdomains
	var hstsExcludeSubdomains bool
	flag.BoolVar(&hstsExcludeSubdomains, "security-hsts-excluede-subdomains", false, "HSTSExcludeSubdomains won't include subdomains tag in the `Strict Transport Security`header")

	// security disable
	// e.g. -security-disable
	var securityDisable bool
	flag.BoolVar(&securityDisable, "security-disable", false, "Disable security middleware (security-content-security-policy, security-xss-protection, security-contenttype-nosniff, security-xframe-options, security-hsts-maxage,security-hsts-excluede-subdomains)")

	// help flag
	var help bool
	flag.BoolVar(&help, "help", false, "print help")

	// version flag
	var version bool
	flag.BoolVar(&version, "version", false, "print version")

	// command line arguments to add one or more custom response headers
	// e.g. -set-custom-header='HEADER_NAME:HEADER_VALUE'
	var headers customHeaders
	flag.Func("set-custom-header", "set custom response header format: 'HEADER_NAME:HEADER_VALUE' ", func(value string) error {
		headers = append(headers, value)
		return nil
	})

	flag.Usage = func() {
		fmt.Println("Usage:")
		flag.PrintDefaults()
	}
	// parse command line arguments
	flag.Parse()

	// print help
	if help {
		flag.Usage()
		os.Exit(0)
	}

	// print version
	if version {
		fmt.Printf("Version: %s\n", GitVersion)
		fmt.Printf("Commit: %s\n", GitCommit)
		fmt.Printf("Build date: %s\n", BuildDate)
		os.Exit(0)
	}

	mux := echo.New()
	mux.Logger.SetLevel(log.INFO)
	mux.Logger.SetPrefix("go-serve")
	mux.HideBanner = true
	mux.Use(middleware.RequestID())
	mux.Use(middleware.Recover())
	mux.Use(middleware.Logger())

	if !securityDisable {
		mux.Use(middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection:         xssProtection,
			ContentTypeNosniff:    contentTypeNosniff,
			XFrameOptions:         xFrameOptions,
			HSTSMaxAge:            hstsMaxAge,
			HSTSExcludeSubdomains: hstsExcludeSubdomains,
			ContentSecurityPolicy: contentSecurityPolicy,
		}))
	}

	mux.Use(headers.setCustomHeaders)

	mux.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   dir,
		Browse: false,
		Index:  "index.html",
		HTML5:  true,
	}))
	// if basePath is not set, serve from root path /env.json
	// if basePath is set, serve from /basePath/env.json
	var configPath string
	if basePath == "/" {
		configPath = "env"
	} else {
		if basePath[len(basePath)-1:] == "/" {
			// remove last character
			basePath = basePath[:len(basePath)-1]
		}
		configPath = basePath + "/env"
	}

	// handle requests
	mux.Static(basePath, dir)
	// environment variables as JSON
	mux.GET(configPath+".json", func(c echo.Context) error {
		config, err := loadEnv(envPrefix)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		if len(config) == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"message": ErrNoEnv.Error()})
		}

		return c.JSON(http.StatusOK, config)
	})

	// environment variables as JS
	mux.GET(configPath+".js", func(c echo.Context) error {
		config, err := loadEnv(envPrefix)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		if len(config) == 0 {
			return c.String(http.StatusNotFound, ErrNoEnv.Error())
		}

		jsConfig, err := json.Marshal(config)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		return c.String(http.StatusOK, "window._env = "+string(jsConfig)+";")
	})

	s := http.Server{
		Addr:         address,
		Handler:      mux,
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
	}

	// startup message
	fmt.Printf("# go-serve %s, commit: %s, build date: %s\n", GitVersion, GitCommit, BuildDate)
	fmt.Println(strings.Repeat("#", 80))
	fmt.Println("Listening on: " + address)
	fmt.Println("Serving from: " + basePath)
	fmt.Println("Path to configs: " + configPath)
	fmt.Println("Environment variables prefixed: " + envPrefix)
	fmt.Println("Press Ctrl+C to quit")
	fmt.Println(strings.Repeat("#", 80))

	// start server
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mux.Logger.Fatal(err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 15 seconds.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := mux.Shutdown(ctx); err != nil {
		mux.Logger.Fatal(err)
	}

}

// setCustomHeaders middleware adds a customer headers to the response.
func (ch *customHeaders) setCustomHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		for _, e := range *ch {
			if strings.Contains(e, ":") {
				pair := strings.SplitN(e, ":", 2)
				if c.Response().Header().Get(pair[0]) == "" {
					c.Response().Header().Set(pair[0], pair[1])
				} else {
					c.Response().Header().Set(pair[0], c.Response().Header().Get(pair[0])+";"+pair[1])
				}
			}
		}
		return next(c)
	}
}

// loadEnv loads environment variables from the environment based on the prefix
func loadEnv(prefix string) (map[string]string, error) {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if strings.HasPrefix(pair[0], prefix) {
			env[strings.TrimPrefix(pair[0], prefix)] = pair[1]
		}
	}

	if env == nil {
		return nil, errors.New("no environment variables found prefixed " + prefix)
	}

	return env, nil
}
