package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"

	"textadventure/engine"
)

// ─── Session store ─────────────────────────────────────────────────────────

var sessions sync.Map // map[string]*engine.GameSession

func newSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func loadSession(r *http.Request) (*engine.GameSession, string, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, "", false
	}
	val, ok := sessions.Load(cookie.Value)
	if !ok {
		return nil, cookie.Value, false
	}
	return val.(*engine.GameSession), cookie.Value, true
}

func saveSession(w http.ResponseWriter, id string, sess *engine.GameSession) {
	sessions.Store(id, sess)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ─── HTML helpers ──────────────────────────────────────────────────────────

// linesToHTML converts engine output lines into <p> elements.
// Empty lines become a spacer paragraph; all content is HTML-escaped.
func linesToHTML(lines []string) string {
	var sb strings.Builder
	for _, line := range lines {
		if line == "" {
			sb.WriteString("<p class=\"gap\">&nbsp;</p>\n")
		} else {
			sb.WriteString("<p>" + html.EscapeString(line) + "</p>\n")
		}
	}
	return sb.String()
}

// ─── Index page ────────────────────────────────────────────────────────────

var indexTmpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Escape from Thornwood Manor</title>
  <script src="https://unpkg.com/htmx.org@2.0.4"
    integrity="sha384-HGfztofotfshcF7+8n44JQL2oJmowVChPTg48S+jvZoztPfvwD79OC/LTtG6dMp+"
    crossorigin="anonymous"></script>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      background: #0d0d0d;
      color: #33e866;
      font-family: "Courier New", Courier, monospace;
      font-size: 14px;
      height: 100dvh;
      display: flex;
      flex-direction: column;
    }
    #output {
      flex: 1;
      overflow-y: auto;
      padding: 16px 16px 8px;
    }
    #output p {
      line-height: 1.6;
      white-space: pre-wrap;
      word-break: break-word;
    }
    #output p.gap { line-height: 0.8; }
    #output p.cmd { color: #88ffaa; }
    #input-row {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 10px 16px;
      background: #1a1a1a;
      border-top: 1px solid #1a3322;
      flex-shrink: 0;
    }
    #input-row .prompt { color: #33e866; font-weight: bold; }
    #cmd {
      flex: 1;
      background: transparent;
      border: none;
      outline: none;
      color: #33e866;
      font-family: inherit;
      font-size: inherit;
      caret-color: #33e866;
    }
    #cmd::placeholder { color: #1f4a2a; }
    button {
      background: transparent;
      border: 1px solid #33e866;
      color: #33e866;
      font-family: inherit;
      font-size: inherit;
      padding: 2px 12px;
      cursor: pointer;
    }
    button:hover  { background: #1a3322; }
    button:disabled { opacity: 0.3; cursor: default; }
    .htmx-request button { opacity: 0.3; pointer-events: none; }
  </style>
</head>
<body>
  <div id="output">
{{.InitialOutput}}  </div>

  <form id="input-row"
        hx-post="/command"
        hx-target="#output"
        hx-swap="beforeend"
        hx-on::after-request="onResponse()">
    <span class="prompt">&gt;</span>
    <input id="cmd" name="input"
           autocomplete="off" spellcheck="false" autofocus
           placeholder="type a command…">
    <button type="submit">Enter</button>
  </form>

  <script>
    const output = document.getElementById('output');
    const cmd    = document.getElementById('cmd');

    function scrollToBottom() {
      output.scrollTop = output.scrollHeight;
    }

    function onResponse() {
      cmd.value = '';
      cmd.focus();
      scrollToBottom();
    }

    scrollToBottom();
  </script>
</body>
</html>
`))

// ─── Handlers ──────────────────────────────────────────────────────────────

func handleIndex(w http.ResponseWriter, r *http.Request) {
	var lines []string

	sess, id, ok := loadSession(r)
	if ok {
		// Returning session: orient the player with a fresh look.
		sess.HandleInput("look")
		lines = sess.FlushOutput()
	} else {
		// New session: surface the intro and initial room description.
		sess = engine.NewSession()
		lines = sess.FlushOutput()
		id = newSessionID()
		saveSession(w, id, sess)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := indexTmpl.Execute(w, struct{ InitialOutput template.HTML }{
		InitialOutput: template.HTML(linesToHTML(lines)),
	}); err != nil {
		log.Printf("template error: %v", err)
	}
}

func handleCommand(w http.ResponseWriter, r *http.Request) {
	sess, id, ok := loadSession(r)
	if !ok {
		// Stale or missing cookie — start a fresh session transparently.
		sess = engine.NewSession()
		sess.FlushOutput() // discard intro; player is mid-conversation
		id = newSessionID()
		saveSession(w, id, sess)
	}

	input := strings.TrimSpace(r.FormValue("input"))
	if input == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	sess.HandleInput(input)
	lines := sess.FlushOutput()

	// Echo the command first, then the engine's response.
	var sb strings.Builder
	fmt.Fprintf(&sb, "<p class=\"cmd\">&gt; %s</p>\n", html.EscapeString(input))
	sb.WriteString(linesToHTML(lines))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, sb.String())
}

// ─── Main ──────────────────────────────────────────────────────────────────

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", handleIndex) // exact "/" only
	mux.HandleFunc("POST /command", handleCommand)

	addr := ":8080"
	log.Printf("listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
