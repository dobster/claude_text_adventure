package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
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
// Empty lines become a spacer paragraph; all text is HTML-escaped.
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

// ─── Handlers ──────────────────────────────────────────────────────────────

// handleIndex serves templates/index.html (assumed relative to CWD).
// It creates a new session on first visit and sets the session cookie.
// The intro text is buffered inside the session; /intro will flush it.
func handleIndex(w http.ResponseWriter, r *http.Request) {
	_, _, ok := loadSession(r)
	if !ok {
		sess := engine.NewSession() // intro is buffered, not yet flushed
		id := newSessionID()
		saveSession(w, id, sess)
	}
	http.ServeFile(w, r, "templates/index.html")
}

// handleIntro returns the intro fragment for the output log.
// For a brand-new session it flushes the buffered intro text; for a
// returning session (empty buffer) it runs `look` to re-orient the player.
func handleIntro(w http.ResponseWriter, r *http.Request) {
	sess, id, ok := loadSession(r)
	if !ok {
		// Cookie missing or session evicted (e.g. server restart).
		sess = engine.NewSession()
		id = newSessionID()
		saveSession(w, id, sess)
	}

	lines := sess.FlushOutput()
	if len(lines) == 0 {
		// Returning session: buffer was already flushed; show current room.
		sess.HandleInput("look")
		lines = sess.FlushOutput()
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, linesToHTML(lines))
}

// handleCommand processes one player command and returns an HTML fragment.
func handleCommand(w http.ResponseWriter, r *http.Request) {
	sess, id, ok := loadSession(r)
	if !ok {
		// Stale or missing cookie — restart transparently.
		sess = engine.NewSession()
		sess.FlushOutput() // discard intro; player is mid-session
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

	// Echo the command first, then the engine response.
	var sb strings.Builder
	fmt.Fprintf(&sb, "<p class=\"cmd\">&gt; %s</p>\n", html.EscapeString(input))
	sb.WriteString(linesToHTML(lines))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, sb.String())
	_ = id
}

// ─── Main ──────────────────────────────────────────────────────────────────

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", handleIndex)       // exact "/"
	mux.HandleFunc("GET /intro", handleIntro)     // initial output fragment
	mux.HandleFunc("POST /command", handleCommand) // player commands

	addr := ":8080"
	log.Printf("listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
