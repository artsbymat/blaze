package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"blaze/internal/engine"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var (
	liveReloadClients = make(map[*websocket.Conn]bool)
	clientsMux        sync.Mutex
	upgrader          = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func main() {
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)

	servePort := serveCmd.String("port", "3000", "Port to serve on")

	if len(os.Args) < 2 {
		fmt.Println("Usage: ssg <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  build    Build the site")
		fmt.Println("  serve    Serve the site with hot reload")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "build":
		buildCmd.Parse(os.Args[2:])
		if err := build(); err != nil {
			log.Fatal(err)
		}
	case "serve":
		serveCmd.Parse(os.Args[2:])
		if err := serve(*servePort); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func build() error {
	ssg, err := engine.NewSSG("content", "templates", "public", "blaze.config.json")
	if err != nil {
		return err
	}

	return ssg.Build()
}

func serve(port string) error {
	if err := build(); err != nil {
		return err
	}

	go watchAndRebuild()

	http.HandleFunc("/livereload", liveReloadHandler)
	http.HandleFunc("/", injectLiveReload)

	fmt.Printf("Serving at http://localhost:%s\n", port)
	fmt.Println("Watching for changes...")

	return http.ListenAndServe(":"+port, nil)
}

func watchAndRebuild() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	watchFilesAndDirs := []string{"blaze.config.json", "content", "templates"}
	for _, filesAndDirs := range watchFilesAndDirs {
		filepath.Walk(filesAndDirs, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			watcher.Add(path)
			return nil
		})
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
				fmt.Printf("Change detected: %s\n", event.Name)
				if err := build(); err != nil {
					log.Printf("Build error: %v\n", err)
				} else {
					fmt.Println("Rebuild complete!")
					triggerReload()
				}
			}
		case err := <-watcher.Errors:
			log.Printf("Watcher error: %v\n", err)
		}
	}
}

func liveReloadHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v\n", err)
		return
	}

	clientsMux.Lock()
	liveReloadClients[conn] = true
	clientsMux.Unlock()

	defer func() {
		clientsMux.Lock()
		delete(liveReloadClients, conn)
		clientsMux.Unlock()
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func triggerReload() {
	clientsMux.Lock()
	defer clientsMux.Unlock()

	for client := range liveReloadClients {
		if err := client.WriteMessage(websocket.TextMessage, []byte("reload")); err != nil {
			log.Printf("Error sending reload: %v\n", err)
		}
	}
}

func injectLiveReload(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	filePath := filepath.Join("public", path)

	if strings.HasSuffix(filePath, ".html") {
		content, err := os.ReadFile(filePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		htmlStr := string(content)
		liveReloadScript := `
<script>
(function() {
	const ws = new WebSocket('ws://' + window.location.host + '/livereload');
	ws.onmessage = function() {
		console.log('Reloading...');
		window.location.reload();
	};
	ws.onclose = function() {
		console.log('Connection closed. Retrying...');
		setTimeout(function() { window.location.reload(); }, 1000);
	};
})();
</script>
</body>`

		htmlStr = strings.Replace(htmlStr, "</body>", liveReloadScript, 1)

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(htmlStr))
		return
	}

	http.FileServer(http.Dir("public")).ServeHTTP(w, r)
}
