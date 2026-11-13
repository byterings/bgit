//go:build devharness

// Dev harness: serves the desktop frontend to a normal browser, backed by the
// real Go backend, so screens can be developed and tested without building the
// Wails app or installing WebKit/GTK.
//
//	go run -tags devharness ./desktop -home /tmp/bgit-harness-home
//
// It always runs against an isolated home directory, never the real one, and
// only accepts same-origin requests from localhost.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// harnessBridge emulates the parts of the Wails runtime the frontend uses.
// Native dialogs do not exist in a browser: path pickers return
// window.__harnessChoice and opened links are recorded in window.__harnessOpened.
const harnessBridge = `(() => {
  const call = async (name, args) => {
    const response = await fetch("/__harness/call/" + name, {
      method: "POST",
      headers: { "Content-Type": "application/json", "X-Bgit-Harness": "1" },
      body: JSON.stringify(args),
    });
    const payload = await response.json();
    if (payload.error) {
      throw payload.error;
    }
    return payload.result;
  };
  window.__harnessChoice = "";
  window.__harnessOpened = [];
  window.__harnessClipboard = "";
  window.go = {
    main: {
      App: new Proxy({}, {
        get: (_, name) => (...args) => {
          if (/^Choose/.test(name)) {
            return Promise.resolve(window.__harnessChoice || "");
          }
          if (name === "OpenGitHubKeySettings" || name === "OpenExternalURL") {
            window.__harnessOpened.push(args[0] || "github-key-settings");
            return Promise.resolve(null);
          }
          return call(name, args);
        },
      }),
    },
  };
  window.runtime = {
    ClipboardSetText: async (text) => { window.__harnessClipboard = text; return true; },
    BrowserOpenURL: (url) => { window.__harnessOpened.push(url); },
    EventsOn() {},
    EventsOff() {},
  };
})();
`

func main() {
	addr := flag.String("addr", "127.0.0.1:34116", "listen address (localhost only)")
	dist := flag.String("dist", "", "frontend directory to serve (default: found automatically)")
	home := flag.String("home", "", "isolated home directory (required)")
	flag.Parse()

	if !strings.HasPrefix(*addr, "127.0.0.1:") && !strings.HasPrefix(*addr, "localhost:") {
		log.Fatal("the dev harness only listens on localhost")
	}
	if err := isolateHome(*home); err != nil {
		log.Fatal(err)
	}
	distDir, err := findFrontend(*dist)
	if err != nil {
		log.Fatal(err)
	}

	app := NewApp()
	// The same repair the real app runs when its window starts.
	repairDesktopHook()
	mux := http.NewServeMux()
	mux.HandleFunc("/__harness/bridge.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		io.WriteString(w, harnessBridge)
	})
	mux.HandleFunc("/__harness/call/", func(w http.ResponseWriter, r *http.Request) {
		handleCall(app, w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveFrontend(distDir, w, r)
	})

	// Listen first, so the "ready" line is only printed once the address is
	// really ours. A port that is already in use fails here with a clear error.
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("cannot start the dev harness on %s: %v. Stop the other harness or pass -addr with another port", *addr, err)
	}
	log.Printf("bgit dev harness on http://%s (home: %s, frontend: %s)", listener.Addr(), os.Getenv("HOME"), distDir)
	log.Fatal(http.Serve(listener, mux))
}

// findFrontend returns the frontend folder. The harness is started both from
// the repository root (`make desktop-harness`) and from desktop/, so both
// layouts are tried when no folder is given.
func findFrontend(dist string) (string, error) {
	candidates := []string{dist}
	if strings.TrimSpace(dist) == "" {
		candidates = []string{filepath.Join("desktop", "frontend", "dist"), filepath.Join("frontend", "dist")}
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(filepath.Join(candidate, "index.html")); err == nil {
			return filepath.Abs(candidate)
		}
	}
	return "", fmt.Errorf("frontend not found (looked in %s). Run from the repository root or pass -dist", strings.Join(candidates, ", "))
}

// isolateHome points everything the backend reads at a throwaway directory.
func isolateHome(home string) error {
	if strings.TrimSpace(home) == "" {
		return fmt.Errorf("-home is required: the harness never runs against your real home directory")
	}
	abs, err := filepath.Abs(home)
	if err != nil {
		return err
	}
	if realHome, err := os.UserHomeDir(); err == nil && filepath.Clean(realHome) == filepath.Clean(abs) {
		return fmt.Errorf("-home must not be your real home directory")
	}
	if err := os.MkdirAll(abs, 0700); err != nil {
		return err
	}
	for key, value := range map[string]string{
		"HOME":                abs,
		"USERPROFILE":         abs,
		"XDG_CONFIG_HOME":     filepath.Join(abs, ".config"),
		"GIT_CONFIG_GLOBAL":   filepath.Join(abs, ".gitconfig"),
		"GIT_CONFIG_NOSYSTEM": "1",
		"GIT_TERMINAL_PROMPT": "0",
	} {
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return nil
}

func serveFrontend(dist string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		http.FileServer(http.Dir(dist)).ServeHTTP(w, r)
		return
	}

	page, err := os.ReadFile(filepath.Join(dist, "index.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// The bridge must exist before any page script runs.
	html := strings.Replace(string(page), "<head>", `<head><script src="/__harness/bridge.js"></script>`, 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, html)
}

// handleCall invokes an exported App method by name with JSON arguments, the
// same way the Wails bridge does.
func handleCall(app *App, w http.ResponseWriter, r *http.Request) {
	// The custom header forces a CORS preflight for cross-site requests, which
	// this server never approves, so other websites cannot drive the backend.
	if r.Method != http.MethodPost || r.Header.Get("X-Bgit-Harness") != "1" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/__harness/call/")
	method := reflect.ValueOf(app).MethodByName(name)
	if !method.IsValid() {
		writeCallResult(w, nil, fmt.Sprintf("unknown backend method %q", name))
		return
	}

	var rawArgs []json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawArgs); err != nil {
		writeCallResult(w, nil, "invalid arguments: "+err.Error())
		return
	}
	methodType := method.Type()
	if len(rawArgs) != methodType.NumIn() {
		writeCallResult(w, nil, fmt.Sprintf("%s expects %d argument(s), got %d", name, methodType.NumIn(), len(rawArgs)))
		return
	}

	args := make([]reflect.Value, len(rawArgs))
	for i, raw := range rawArgs {
		value := reflect.New(methodType.In(i))
		if err := json.Unmarshal(raw, value.Interface()); err != nil {
			writeCallResult(w, nil, fmt.Sprintf("argument %d of %s: %v", i+1, name, err))
			return
		}
		args[i] = value.Elem()
	}

	var result interface{}
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	for _, out := range method.Call(args) {
		if out.Type().Implements(errorType) {
			if !out.IsNil() {
				writeCallResult(w, nil, out.Interface().(error).Error())
				return
			}
			continue
		}
		result = out.Interface()
	}
	writeCallResult(w, result, "")
}

func writeCallResult(w http.ResponseWriter, result interface{}, errorMessage string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	payload := map[string]interface{}{"result": result}
	if errorMessage != "" {
		payload = map[string]interface{}{"error": errorMessage}
	}
	json.NewEncoder(w).Encode(payload)
}
