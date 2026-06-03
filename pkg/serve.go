package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"n8go-docs/diagnostics"
	"n8go-docs/editor"
	"n8go-docs/utils"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

const editorMountPath = "/_editor"

func DefaultNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 - Not Found"))
}

func FileServerWithCustom404(fs http.FileSystem, addr string) http.Handler {
	color.Green("Serving on %s", displayServerAddr(addr))
	fsh := http.FileServer(fs)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fs.Open(path.Clean(r.URL.Path))
		if err == nil {
			fsh.ServeHTTP(w, r)
			return
		}
		DefaultNotFound(w, r)
	})
}

func runServer(configPath string, port int) error {
	site, theme, themeDir, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	pipeline := buildPipeline(site, theme, themeDir)

	// Initial full build
	if err := pipeline.Build(); err != nil {
		return err
	}

	// File watcher — triggers a full rebuild on any change in docs_dir
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) && utils.ShouldRebuild(event.Name, event.Op) {
					color.Yellow("File system has changed, regenerating...")
					if site.DefaultSearch {
						_ = os.Remove(site.OutputPath + "/search/index.json")
					}
					if err := pipeline.Build(); err != nil {
						diagnostics.PrintError(err, "failed to regenerate")
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("watcher error: %s\n", err)
			}
		}
	}()

	if err := watcher.Add(site.InputPath); err != nil {
		return err
	}

	// Build HTTP mux: static file server + editor API
	mux := http.NewServeMux()

	ed := editor.New(pipeline)
	ed.Mount(mux, editorMountPath)

	// Catch-all: serve static site files
	fileServer := FileServerWithCustom404(http.Dir(site.OutputPath), "")
	mux.Handle("/", fileServer)

	addr := resolveDevAddr(site.DevAddr, port)
	color.Green("Serving on %s", displayServerAddr(addr))
	color.Cyan("Editor API at %s%s", displayServerAddr(addr), editorMountPath)

	return http.ListenAndServe(addr, mux)
}

func resolveDevAddr(devAddr string, port int) string {
	if devAddr != "" && port == defaultServePort {
		return devAddr
	}
	return ":" + strconv.Itoa(port)
}

func displayServerAddr(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "http://localhost" + addr
	}
	return "http://" + addr
}
