# 007 hugo



# Hugo + Datastar WYSIWYG Editor (Single‑Markdown Spec)

> **Goal:** A minimal, production‑inclined WYSIWYG page inside a Hugo site that live‑previews and collaboratively syncs content using **Datastar** over **SSE**, with a tiny **Go** backend (no Node) that writes to `content/` and broadcasts updates. Everything you need is in this one markdown to copy.

---

## 0) Features

* **WYSIWYG editing** (contenteditable) for Markdown/HTML blocks with a simple toolbar (bold/italic/links/headings/lists/code/undo/redo).
* **Live preview** using Datastar SSE morphs (Idiomorph) to sync multiple browsers.
* **File‑backed persistence**: Go server writes to `./site/content/_editor/doc.md` (configurable).
* **Collaboration‑ready**: every save broadcasts the canonical HTML to all viewers of the same doc.
* **No Node build**: pure Go + Hugo + vanilla JS. Optional PocketBase/NATS later.

---

## 1) Prereqs

* **Go** 1.22+
* **Hugo (extended)** v0.125+ (or recent)
* **Datastar** JS via CDN (no build step)

```bash
# macOS example
brew install go hugo
```

---

## 2) File/Folder Layout

Create a workspace folder like this (names can change). The Go server and Hugo site live side‑by‑side:

```
wysi/
  server/
    main.go
  site/
    config.toml
    content/
      _editor/
        doc.md              # the document we edit (seeded by you)
    layouts/
      _default/
        editor.html         # the editor/preview UI
      partials/
        editor-toolbar.html # small toolbar partial
```

> **Note:** The Go server writes to `site/content/_editor/doc.md` and serves SSE on `:8080`. Hugo dev server serves pages on `:1313`.

---

## 3) Go Backend (SSE + Save)

**`server/main.go`** — a compact Datastar‑style broadcaster using pure SSE and an in‑process hub. It accepts `POST /save` with HTML + Markdown mirror, writes the Markdown file, remembers canonical HTML, and broadcasts to `/sse?doc=doc` channel.

```go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type event struct{ Data string }

type hub struct {
	mu   sync.RWMutex
	chs  map[string]map[chan event]struct{} // topic -> set of client chans
	last map[string]string                  // topic -> last HTML snapshot
}

func newHub() *hub { return &hub{chs: map[string]map[chan event]struct{}{}, last: map[string]string{}} }

func (h *hub) subscribe(topic string) (chan event, func()) {
	h.mu.Lock(); defer h.mu.Unlock()
	c := make(chan event, 8)
	if h.chs[topic] == nil { h.chs[topic] = map[chan event]struct{}{} }
	h.chs[topic][c] = struct{}{}
	return c, func() {
		h.mu.Lock(); defer h.mu.Unlock()
		delete(h.chs[topic], c)
		close(c)
	}
}

func (h *hub) publish(topic, html string) {
	h.mu.Lock()
	h.last[topic] = html
	chs := h.chs[topic]
	h.mu.Unlock()
	for c := range chs { select { case c <- event{Data: html}: default: } }
}

func (h *hub) lastHTML(topic string) (string, bool) {
	h.mu.RLock(); defer h.mu.RUnlock()
	html, ok := h.last[topic]
	return html, ok
}

// simple payload coming from the editor
type saveReq struct {
	DocID   string `json:"doc_id"`
	HTML    string `json:"html"`
	Markdown string `json:"markdown"`
}

func main() {
	h := newHub()

	// SSE endpoint (Datastar-compatible: text/event-stream)
	http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok { http.Error(w, "stream unsupported", 500); return }
		topic := r.URL.Query().Get("doc")
		if topic == "" { topic = "doc" }
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(200)

		c, cancel := h.subscribe(topic)
		defer cancel()

		// send last snapshot if any
		if html, ok := h.lastHTML(topic); ok {
			fmt.Fprintf(w, "event: morph\n")
			fmt.Fprintf(w, "data: %s\n\n", jsonEscapeLine(html))
			flusher.Flush()
		}

		// heartbeat
		go func() {
			for {
				time.Sleep(20 * time.Second)
				fmt.Fprintf(w, ": ping\n\n")
				flusher.Flush()
			}
		}()

		// stream updates
		for ev := range c {
			fmt.Fprintf(w, "event: morph\n") // Datastar listens for morph by default
			fmt.Fprintf(w, "data: %s\n\n", jsonEscapeLine(ev.Data))
			flusher.Flush()
		}
	})

	// Save endpoint: writes Markdown to Hugo content and broadcasts canonical HTML
	http.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { http.Error(w, "method", 405); return }
		var req saveReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, err.Error(), 400); return }
		if req.DocID == "" { req.DocID = "doc" }

		// persist Markdown
		mdPath := filepath.Join("..", "site", "content", "_editor", req.DocID+".md")
		if err := os.MkdirAll(filepath.Dir(mdPath), 0o755); err != nil { http.Error(w, err.Error(), 500); return }
		if err := os.WriteFile(mdPath, []byte(req.Markdown), 0o644); err != nil { http.Error(w, err.Error(), 500); return }

		// broadcast HTML snapshot
		h.publish(req.DocID, req.HTML)
		w.WriteHeader(204)
	})

	// Tiny health
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })

	log.Println("Server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func jsonEscapeLine(s string) string {
	// SSE data lines must not contain raw newlines except as record terminators.
	// We send a compact single line JSON string (already HTML) to Datastar.
	b := &strings.Builder{}
	w := bufio.NewWriter(b)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(s)
	w.Flush()
	line := strings.TrimSpace(b.String())
	return line
}
```

**Run the server:**

```bash
cd wysi/server
go run .
```

---

## 4) Hugo Site

### 4.1 `site/config.toml`

```toml
baseURL = "/"
languageCode = "en-us"
title = "WYSI with Datastar"
# we’ll serve editor via hugo server locally
```

### 4.2 Seed Content `site/content/_editor/doc.md`

```md
---
title: Live Doc
---

# Hello, Datastar x Hugo

This is **your** collaborative, WYSIWYG‑editable page. Open this page in two browsers to see live morphs.
```

### 4.3 Editor Template `site/layouts/_default/editor.html`

This single template shows **Editor** (left) and **Preview** (right). The page connects to our SSE (`/sse?doc=doc`) and morphs `#preview` on each update.

```html
{{- define "main" -}}
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>WYSIWYG Editor — Datastar</title>
  <!-- Datastar: lightweight SSE + Idiomorph -->
  <script defer src="https://unpkg.com/@starfederation/datastar@latest/dist/datastar.min.js"></script>
  <style>
    html,body{height:100%;margin:0;font:16px/1.5 system-ui,ui-sans-serif,Segoe UI,Roboto}
    .wrap{display:grid;grid-template-columns:1fr 1fr;gap:16px;height:100vh;padding:16px;box-sizing:border-box}
    .panel{border:1px solid #ddd;border-radius:12px;overflow:auto;display:flex;flex-direction:column}
    .toolbar{display:flex;gap:8px;padding:8px;border-bottom:1px solid #eee;position:sticky;top:0;background:#fff}
    .editor{padding:16px;min-height:calc(100vh - 120px);outline:none}
    .editor:empty:before{content:attr(data-placeholder);opacity:.5}
    .preview{padding:16px}
    button{border:1px solid #ddd;border-radius:8px;padding:6px 10px;background:#fafafa;cursor:pointer}
    button:active{transform:translateY(1px)}
    .status{font-size:12px;color:#555;margin-left:auto}
  </style>
</head>
<body data-datastar>
  <div class="wrap">
    <section class="panel">
      <div class="toolbar">
        {{ partial "editor-toolbar.html" . }}
        <span class="status" id="status">Idle</span>
      </div>
      <div id="editor" class="editor" contenteditable="true" data-placeholder="Start typing…"></div>
    </section>

    <section class="panel">
      <div class="toolbar"><strong>Preview</strong></div>
      <div id="preview" class="preview">
        {{- $p := site.GetPage "_editor/doc.md" -}}
        {{- with $p -}}
          {{ .Content | safeHTML }}
        {{- end -}}
      </div>
    </section>
  </div>

  <script>
    // --- constants
    const DOC_ID = 'doc';
    const SSE_URL = 'http://localhost:8080/sse?doc='+encodeURIComponent(DOC_ID);
    const SAVE_URL = 'http://localhost:8080/save';

    // --- naive HTML <-> Markdown mirror
    // For a pure-Go stack, keep Markdown as-is when possible.
    // Here we round-trip: prefer user HTML in editor and also keep a crude markdown mirror.
    function htmlToMarkdown(html){
      // super minimal mirror (keep block HTML as-is); swap <strong>/<em> etc.
      return html
        .replaceAll(/<strong>(.*?)<\/strong>/g, '**$1**')
        .replaceAll(/<b>(.*?)<\/b>/g, '**$1**')
        .replaceAll(/<em>(.*?)<\/em>/g, '*$1*')
        .replaceAll(/<i>(.*?)<\/i>/g, '*$1*')
        .replaceAll(/<h1>(.*?)<\/h1>/g, '# $1\n')
        .replaceAll(/<h2>(.*?)<\/h2>/g, '## $1\n')
        .replaceAll(/<h3>(.*?)<\/h3>/g, '### $1\n')
        .replaceAll(/<br\s*\/?>(?!\n)/g, '\n')
        .replaceAll(/<p>(.*?)<\/p>/g, '$1\n\n')
        .replaceAll(/<ul>([\s\S]*?)<\/ul>/g, (_,m)=>m.replaceAll(/<li>([\s\S]*?)<\/li>/g,'- $1\n'))
        .replaceAll(/<ol>([\s\S]*?)<\/ol>/g, (_,m)=>m.replaceAll(/<li>([\s\S]*?)<\/li>/g,(x,i)=>`${i+1}. ${x.replace(/<li>|<\/li>/g,'')}\n`))
        .replaceAll(/<a[^>]*href="([^"]+)"[^>]*>(.*?)<\/a>/g,'[$2]($1)')
        .replace(/\n{3,}/g, '\n\n')
        .trim();
    }

    // --- editor setup
    const editor = document.getElementById('editor');
    const preview = document.getElementById('preview');
    const status = document.getElementById('status');

    // seed editor from preview's initial HTML
    editor.innerHTML = preview.innerHTML;

    // toolbar actions
    function cmd(name, value=null){ document.execCommand(name, false, value); editor.dispatchEvent(new Event('input')); }
    window.__ed = {
      bold: () => cmd('bold'),
      italic: () => cmd('italic'),
      h1: () => { document.execCommand('formatBlock', false, 'h1'); editor.dispatchEvent(new Event('input')); },
      h2: () => { document.execCommand('formatBlock', false, 'h2'); editor.dispatchEvent(new Event('input')); },
      p:  () => { document.execCommand('formatBlock', false, 'p');  editor.dispatchEvent(new Event('input')); },
      ul: () => cmd('insertUnorderedList'),
      ol: () => cmd('insertOrderedList'),
      link: () => { const url = prompt('URL'); if(url) cmd('createLink', url); },
      code: () => cmd('formatBlock','pre'),
      undo: () => cmd('undo'),
      redo: () => cmd('redo'),
      save: debounce(saveNow, 0),
    };

    // auto-save on input
    let dirty=false; let t;
    editor.addEventListener('input', ()=>{ dirty=true; status.textContent='Editing…'; scheduleSave(); });
    function scheduleSave(){ clearTimeout(t); t=setTimeout(()=>__ed.save(), 500); }

    async function saveNow(){
      if(!dirty) return; dirty=false; status.textContent='Saving…';
      const html = sanitize(editor.innerHTML);
      const markdown = htmlToMarkdown(html);
      try{
        await fetch(SAVE_URL, {
          method:'POST', headers:{'Content-Type':'application/json'},
          body: JSON.stringify({ doc_id: DOC_ID, html, markdown })
        });
        status.textContent='Saved ✓';
      }catch(e){ console.error(e); status.textContent='Save failed'; }
    }

    function sanitize(html){
      // keep it simple: remove script/style tags
      const d = new DOMParser().parseFromString(html,'text/html');
      d.querySelectorAll('script,style').forEach(n=>n.remove());
      return d.body.innerHTML.trim();
    }

    function debounce(fn, ms){ let id; return (...a)=>{ clearTimeout(id); id=setTimeout(()=>fn(...a), ms); } }

    // --- Datastar connection (SSE -> morph #preview)
    // Minimal wire-up: Create a hidden receiver that Datastar watches and morphs #preview.
    const ds = document.createElement('div');
    ds.setAttribute('data-datastar-sse', SSE_URL);
    ds.setAttribute('data-datastar-sse-event', 'morph');
    ds.setAttribute('data-datastar-morph', '#preview');
    document.body.appendChild(ds);

    // if another client updates, update editor too (keep cursor reasonable)
    const observer = new MutationObserver(()=>{
      // when preview morphs, copy HTML to editor if remote changed
      if(!dirty){ editor.innerHTML = preview.innerHTML; }
    });
    observer.observe(preview, {childList:true,subtree:true});

    // save before unload
    window.addEventListener('beforeunload', ()=>{ if(dirty) saveNow(); });
  </script>
</body>
</html>
{{- end -}}
```

### 4.4 Toolbar Partial `site/layouts/partials/editor-toolbar.html`

```html
<button onclick="__ed.bold()"><strong>B</strong></button>
<button onclick="__ed.italic()"><em>I</em></button>
<button onclick="__ed.h1()">H1</button>
<button onclick="__ed.h2()">H2</button>
<button onclick="__ed.p()">P</button>
<button onclick="__ed.ul()">• List</button>
<button onclick="__ed.ol()">1. List</button>
<button onclick="__ed.link()">Link</button>
<button onclick="__ed.code()">Code</button>
<button onclick="__ed.undo()">Undo</button>
<button onclick="__ed.redo()">Redo</button>
<button onclick="__ed.save()">Save</button>
```

---

## 5) Run It

Open two terminals.

**Terminal A — Go SSE/Save server**

```bash
cd wysi/server
go run .
```

**Terminal B — Hugo dev server**

```bash
cd wysi/site
hugo server -D
```

Open: `http://localhost:1313/_editor/doc/` (Hugo will map the content to this URL). If you don’t see the editor, create an explicit page route by adding a list template or visit the layout directly by making a page that uses `editor.html`. Simplest: create `_index.md` under `content/_editor/`:

```md
---
title: Editor
_url: /editor/
---
```

Then navigate to `http://localhost:1313/editor/`.

> **Note:** In production, you can serve the Go SSE/Save behind the same domain (reverse proxy) and update `SSE_URL`/`SAVE_URL` to relative paths.

---

## 6) How It Works (Datastar Flow)

* Page embeds a hidden Datastar receiver with `data-datastar-sse="/sse?doc=doc"` and `data-datastar-morph="#preview"`.
* When the Go server receives a save, it **broadcasts** the canonical HTML to the topic `doc`.
* Datastar sees an SSE event `morph` and **Idiomorphs** `#preview` DOM.
* Other open editors detect the preview change (MutationObserver) and mirror into the editor if local state isn’t dirty → **collab feel**.

---

## 7) Variations & Extensions

1. **Multiple documents**: pass `?doc=<slug>` on the page and update `DOC_ID` from the URL; the Go server saves to `<slug>.md`.
2. **Markdown‑first**: instead of editing raw HTML, plug a Markdown WYSIWYG (e.g., SimpleMDE) or use a `<textarea>` + `marked` to render preview; still broadcast HTML.
3. **Persistence backends**:

   * **PocketBase**: POST to PB collection, then PB trigger publishes to this SSE hub (or broadcast from Go after successful write).
   * **NATS**: On save, publish to `editor.<doc>.updated`; a small NATS consumer (JetStream) calls `h.publish`.
4. **Auth**: add a very small session cookie check in Go; or put `/save` behind your reverse proxy auth.
5. **CRDT**: replace MutationObserver merge with Y.js/Automerge (requires bundling) or NATS‑fed OT; Datastar still handles preview morphs.
6. **Images**: handle drag‑drop upload to `/upload` then insert `<img>` tag at caret.
7. **Hotkeys**: bind `meta+s` to `__ed.save()`.

---

## 8) Minimal Routing Tip for Hugo

If you prefer a dedicated route: create `content/editor/_index.md` and set `type = "editor"`, then put the template as `layouts/editor/list.html` (copy body of `editor.html`). Hugo will render `/editor/` directly.

---

## 9) Production Notes

* Serve SSE with **gzip disabled** and keep‑alive.
* If placing behind Nginx/Caddy, use `proxy_buffering off;` for SSE.
* Add **ETag/Last‑Modified** to content endpoint if you later fetch Markdown separately.
* Debounce saves (already 500ms here). Consider **versioning** to avoid clobbers.

---

## 10) Optional: Tiny NATS Bridge (drop‑in)

If you want the **broadcast** to fan out via NATS (JetStream), add NATS to the Go server:

```go
// import "github.com/nats-io/nats.go"
// On startup:
// nc, _ := nats.Connect(nats.DefaultURL)
// js, _ := nc.JetStream()
// js.Publish("editor."+req.DocID+".updated", []byte(req.HTML))
// Then subscribe once and call h.publish on each message.
```

This keeps Datastar’s client side exactly the same while integrating with your larger control‑plane.

---

## 11) Troubleshooting

* **No live preview?** Devtools → Network → `sse` should stream. If not, check CORS between `:1313` and `:8080`. Add permissive CORS headers to the Go server for dev:

```go
w.Header().Set("Access-Control-Allow-Origin", "*")
```

* **Hugo doesn’t show the editor page?** Ensure you created a route (`/editor/`) that renders `editor.html` or moved the template to a type‑specific layout.
* **Save writes but no morph?** Confirm the server prints no errors and that the client appended the Datastar receiver div.

---

## 12) License

Do what you want. Ship it.

---

## 13) Site‑wide Navigation & Multi‑Page Editing (Hugo tree)

> Expand the single‑page editor into a **site browser** so users can open, create, rename, move, and delete any content page, edit front‑matter, and switch between pages without leaving the UI.

### 13.1 Goals

* **Sidebar tree** of `content/` (folders + Markdown bundles).
* **Search** (title, slug, path) with client‑side index (lunr‑style) or simple substring.
* **Multi‑doc editing**: switch documents; SSE topic becomes `doc=<path slug>`.
* **Front‑matter** editor (YAML/TOML) with safe round‑trip.
* **File ops**: new/move/rename/delete with server‑side validation.
* **Live tree updates**: FS changes (or Git pulls) broadcast via SSE to keep the sidebar in sync.

### 13.2 Server API (Go) — add endpoints

Extend the Go server with a tiny REST surface and a tree broadcaster. Keep the existing `/sse` for doc morphs; add a **tree** channel.

```go
// Add: go.mod deps (none required); optional: github.com/fsnotify/fsnotify for live tree.
// Pseudocode snippets to merge into main.go

// 1) Represent the tree
type Node struct {
    Name string   `json:"name"`
    Path string   `json:"path"`   // content‑relative (e.g. "blog/post.md")
    URL  string   `json:"url"`    // best‑effort route (derive from front‑matter or path)
    Kind string   `json:"kind"`   // "file" | "dir"
    Kids []Node   `json:"kids,omitempty"`
}

func scanTree(root string) (Node, error) { /* walk ./site/content and build Node */ return Node{}, nil }

// 2) Serve the tree as JSON
http.HandleFunc("/api/tree", func(w http.ResponseWriter, r *http.Request){
    t, err := scanTree(filepath.Join("..","site","content"))
    if err != nil { http.Error(w, err.Error(), 500); return }
    writeJSON(w, t)
})

// 3) Open a page (return front‑matter + body)
// Detect front‑matter fences: "---" (YAML) or "+++" (TOML).
http.HandleFunc("/api/open", func(w http.ResponseWriter, r *http.Request){ /* read file, split FM+body */ })

// 4) Save a page (front‑matter + body)
http.HandleFunc("/api/save-page", func(w http.ResponseWriter, r *http.Request){ /* merge FM, body, write */ })

// 5) File ops
http.HandleFunc("/api/new", func(w http.ResponseWriter, r *http.Request){ /* create file/folder */ })
http.HandleFunc("/api/move", func(w http.ResponseWriter, r *http.Request){ /* rename/move */ })
http.HandleFunc("/api/delete", func(w http.ResponseWriter, r *http.Request){ /* delete */ })

// 6) Tree SSE — broadcast when fs changes or after writes
// topic: "tree"; payload: compact JSON of Node
func publishTree(h *hub) { if t, err := scanTree(filepath.Join("..","site","content")); err==nil { h.publish("tree", mustJSON(t)) } }
```

> Call `publishTree(h)` after any `/api/*` write op. Optionally wire **fsnotify** to emit on disk changes (Git pulls, external edits).

### 13.3 Client UI — sidebar + router

In `editor.html`, transform the layout to include a **left sidebar** and make the editor load the selected page.

```html
<!-- Replace .wrap grid with three columns: sidebar | editor | preview -->
<div class="wrap" style="grid-template-columns: 300px 1fr 1fr;">
  <aside class="panel">
    <div class="toolbar"><strong>Site</strong></div>
    <div style="padding:8px"><input id="q" placeholder="Search…" style="width:100%"/></div>
    <nav id="tree" class="preview" style="padding:8px"></nav>
  </aside>
  <!-- existing Editor panel -->
  <!-- existing Preview panel -->
</div>

<script>
  // Track current doc path (relative to content/)
  let CURRENT_PATH = new URLSearchParams(location.search).get('path') || '_editor/doc.md';

  // Update SSE endpoints whenever CURRENT_PATH changes
  function setDoc(path){
    CURRENT_PATH = path;
    // Repoint SSE receiver for doc topic
    const sse = document.querySelector('[data-datastar-sse]');
    sse.setAttribute('data-datastar-sse', 'http://localhost:8080/sse?doc='+encodeURIComponent(CURRENT_PATH));
    // request an immediate refresh (server may push last snapshot)
    fetch('http://localhost:8080/api/open?path='+encodeURIComponent(CURRENT_PATH))
      .then(r=>r.json()).then(({html, markdown})=>{
        preview.innerHTML = html;
        editor.innerHTML = html;
      });
    history.replaceState(null, '', '?path='+encodeURIComponent(CURRENT_PATH));
  }

  // Render simple tree (dir -> ul/li)
  function renderTree(node){
    if(!node) return '';
    if(node.kind==='dir'){
      const kids = (node.kids||[]).map(renderTree).join('');
      const label = node.name || '/';
      return `<details open><summary>${label}</summary><ul>${kids}</ul></details>`;
    }
    // file
    return `<li><a href="#" data-open="${node.path}">${node.name}</a></li>`;
  }

  // Load tree initially
  async function loadTree(){
    const t = await fetch('http://localhost:8080/api/tree').then(r=>r.json());
    document.getElementById('tree').innerHTML = renderTree(t);
  }
  loadTree();

  // Click to open
  document.addEventListener('click', (e)=>{
    const a = e.target.closest('a[data-open]');
    if(a){ e.preventDefault(); setDoc(a.getAttribute('data-open')); }
  });

  // Search (basic substring on path/name)
  document.getElementById('q').addEventListener('input', async (e)=>{
    const q = e.target.value.toLowerCase();
    const t = await fetch('http://localhost:8080/api/tree').then(r=>r.json());
    function filter(node){
      if(node.kind==='dir'){
        const kids=(node.kids||[]).map(filter).filter(Boolean);
        if(kids.length) return {...node, kids};
        return null;
      }
      const hay = (node.name+' '+node.path).toLowerCase();
      return hay.includes(q) ? node : null;
    }
    const ft = filter(t) || {name:'No results', kind:'dir', kids:[]};
    document.getElementById('tree').innerHTML = renderTree(ft);
  });

  // Subscribe to tree SSE updates and re-render
  const treeDS = document.createElement('div');
  treeDS.setAttribute('data-datastar-sse', 'http://localhost:8080/sse?doc=tree');
  treeDS.setAttribute('data-datastar-sse-event', 'tree');
  document.body.appendChild(treeDS);
  // Minimal listener — reuse sse endpoint, parse JSON payload
  const es = new EventSource('http://localhost:8080/sse?doc=tree');
  es.addEventListener('morph', async (ev)=>{ try{ const t=JSON.parse(JSON.parse(ev.data)); document.getElementById('tree').innerHTML = renderTree(t);}catch{} });
</script>
```

> **Note:** We reuse the existing `/sse` and send the tree JSON on topic `tree`. For simplicity, we handle it with a dedicated `EventSource` and manual render.

### 13.4 Front‑matter editor

Add a small front‑matter drawer under the toolbar so editors can set `title`, `slug`, `draft`, etc. Keep it schema‑lite and round‑trip TOML/YAML by preserving unknown keys.

```html
<div class="toolbar">
  {{ partial "editor-toolbar.html" . }}
  <details style="margin-left:auto"><summary>Front‑matter</summary>
    <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px;padding:8px">
      <label>Title <input id="fm_title"/></label>
      <label>Slug <input id="fm_slug"/></label>
      <label>Draft <input id="fm_draft" type="checkbox"/></label>
      <button type="button" onclick="saveFrontMatter()">Apply</button>
    </div>
  </details>
</div>
<script>
  async function loadFrontMatter(){
    const j = await fetch('http://localhost:8080/api/open?path='+encodeURIComponent(CURRENT_PATH)).then(r=>r.json());
    const fm = j.frontmatter||{};
    fm_title.value = fm.title||'';
    fm_slug.value  = fm.slug||'';
    fm_draft.checked = !!fm.draft;
  }
  async function saveFrontMatter(){
    const j = await fetch('http://localhost:8080/api/open?path='+encodeURIComponent(CURRENT_PATH)).then(r=>r.json());
    j.frontmatter = {...(j.frontmatter||{}), title: fm_title.value, slug: fm_slug.value, draft: fm_draft.checked};
    j.html = sanitize(editor.innerHTML);
    j.markdown = htmlToMarkdown(j.html);
    await fetch('http://localhost:8080/api/save-page', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ path: CURRENT_PATH, frontmatter: j.frontmatter, body: j.markdown })});
    // tree may change if slug/path changed → server publishes tree
  }
  // call loadFrontMatter() whenever setDoc() switches pages
  const _origSetDoc = setDoc; setDoc = async function(p){ _origSetDoc(p); setTimeout(loadFrontMatter, 50); }
  // initial
  loadFrontMatter();
</script>
```

### 13.5 URL/Slug mapping and bundles

* **Where to write?** Use the clicked node’s `path` as the source of truth. For bundle folders (`my-post/index.md`), keep assets next to it.
* **Slug change → move:** if front‑matter `slug` changes and the file is under a section, offer to **rename/move** the file/folder with `/api/move`.
* **Routes:** derive `URL` as `/section/slug/` for list/single; show a **View** link that opens the page served by `hugo server`.

### 13.6 Hugo‑generated JSON (optional alternative)

If you prefer Hugo to describe the site instead of scanning files, add a JSON output:

```
# config.toml
[outputs]
  home = ["HTML","JSON"]

# layouts/index.json
{{/* emit a compact pages list */}}
[
{{- $first := true -}}
{{- range .Site.RegularPages -}}
  {{- if not $first }},{{ end -}}
  {"title":{{ .Title | jsonify }},"path":{{ .File.Path | jsonify }},"url":{{ .Permalink | jsonify }}}
  {{- $first = false -}}
{{- end -}}
]
```

Then fetch `/index.json` to render the sidebar.

### 13.7 “Hugo build” preview tab (for shortcodes)

The live preview uses raw HTML; Hugo shortcodes/partials won’t render there. Add a **secondary tab** that iframes the built page:

```html
<div class="toolbar"><strong>Preview</strong>
  <span style="margin-left:auto"></span>
  <button onclick="openBuilt()">Hugo Build</button>
</div>
<script>
  function openBuilt(){ window.open('http://localhost:1313/'+(CURRENT_PATH.replace(/^(.+?)\/(.+)\.md$/, '$1/$2/')), '_blank'); }
</script>
```

### 13.8 Git (optional)

Use `go-git` to commit on each save with author from a cookie; expose `/api/history?path=…` to show last commits and **revert**.

### 13.9 Permissions (optional)

Add a simple allowlist per directory and enforce in `/api/*`. For multi‑tenant setups, mount separate `content/` roots per tenant and namespace SSE topics accordingly.

---

**Outcome:** Editors can browse the whole Hugo site, open any page, change front‑matter, and edit content WYSIWYG with Datastar‑powered live sync. The tree and pages stay up‑to‑date even as files change on disk or via Git pulls.

---

## 14) Optional NATS Fan‑Out (runtime toggle)

> Enable cross‑process/cross‑DC broadcasts via **NATS / JetStream** with a toggle. When off, the in‑process SSE hub works exactly as before. When on, saves publish to NATS subjects and any node subscribed will fan‑in to its local SSE hub.

### 14.1 Toggle options

Pick ONE of the following (both supported):

* **Env vars**

  * `BUS=local` (default) or `BUS=nats`
  * `NATS_URL=nats://127.0.0.1:4222`
  * `NATS_JS=0|1` (0 = plain core NATS, 1 = JetStream)
* **CLI flags**

  * `-bus=local|nats`
  * `-nats-url=<url>`
  * `-nats-js` (boolean)

### 14.2 Subjects

* **Document HTML snapshot updates:** `editor.updates` with headers:

  * `path`: content‑relative path (e.g. `blog/post.md`)
  * `event`: `html`
  * `id`: dedupe id (UUID)
* **Tree updates:** `editor.tree` with header `event: tree`, `id: <uuid>`; body = compact JSON of tree.

### 14.3 Go changes (merge into `server/main.go`)

Add imports:

```go
import (
    // ... existing
    "flag"
    "crypto/rand"
    "encoding/hex"
    nats "github.com/nats-io/nats.go"
)
```

Add config + NATS globals:

```go
type cfg struct {
    Bus     string // "local" | "nats"
    NatsURL string
    UseJS   bool
}

var (
    nc  *nats.Conn
    js  nats.JetStreamContext
    app cfg
)

func uuid() string { b := make([]byte, 16); rand.Read(b); return hex.EncodeToString(b) }
```

Read flags/env in `main()` before handlers:

```go
flag.StringVar(&app.Bus, "bus", getenv("BUS", "local"), "bus: local|nats")
flag.StringVar(&app.NatsURL, "nats-url", getenv("NATS_URL", nats.DefaultURL), "nats url")
useJS := getenv("NATS_JS", "0") == "1"
flag.BoolVar(&app.UseJS, "nats-js", useJS, "use jetstream")
flag.Parse()

if app.Bus == "nats" {
    var err error
    nc, err = nats.Connect(app.NatsURL)
    if err != nil { log.Fatalf("nats: %v", err) }
    if app.UseJS { js, err = nc.JetStream(); if err != nil { log.Fatalf("js: %v", err) } }
    // subscribe to fan‑in editor updates
    _, err = nc.Subscribe("editor.updates", func(m *nats.Msg){
        // de‑dupe via id
        id := m.Header.Get("id")
        path := m.Header.Get("path")
        if id == "" || path == "" { return }
        // forward into local SSE hub
        h.publish(path, string(m.Data))
    })
    if err != nil { log.Fatalf("sub: %v", err) }
    _, err = nc.Subscribe("editor.tree", func(m *nats.Msg){
        // emit latest tree JSON into local topic "tree"
        h.publish("tree", string(m.Data))
    })
    if err != nil { log.Fatalf("sub tree: %v", err) }
    log.Printf("NATS connected %s (js=%v)", app.NatsURL, app.UseJS)
}
```

Helper `getenv`:

```go
func getenv(k, def string) string { if v := os.Getenv(k); v != "" { return v }; return def }
```

Publish on save and tree updates:

```go
func publishDoc(path, html string) {
    // local always
    h.publish(path, html)
    if app.Bus != "nats" || nc == nil { return }
    hdr := nats.Header{}
    hdr.Set("event", "html")
    hdr.Set("path", path)
    hdr.Set("id", uuid())
    msg := &nats.Msg{Subject: "editor.updates", Header: hdr, Data: []byte(html)}
    if app.UseJS && js != nil {
        js.PublishMsg(msg)
    } else {
        nc.PublishMsg(msg)
    }
}

func publishTreeNATS(treeJSON string) {
    h.publish("tree", treeJSON)
    if app.Bus != "nats" || nc == nil { return }
    hdr := nats.Header{}
    hdr.Set("event", "tree")
    hdr.Set("id", uuid())
    msg := &nats.Msg{Subject: "editor.tree", Header: hdr, Data: []byte(treeJSON)}
    if app.UseJS && js != nil { js.PublishMsg(msg) } else { nc.PublishMsg(msg) }
}
```

Use the new functions in handlers:

```go
// in /save handler after writing markdown
publishDoc(req.DocID, req.HTML)

// wherever you previously called publishTree(h)
if t, err := scanTree(filepath.Join("..","site","content")); err==nil {
    publishTreeNATS(mustJSON(t))
}
```

Add `mustJSON` helper:

```go
func mustJSON(v any) string { b,_ := json.Marshal(v); return string(b) }
```

### 14.4 JetStream Streams (optional, when `-nats-js`)

Create durable streams so updates are replayable to late subscribers (e.g., a worker indexing content). Auto‑provision on startup:

```go
if app.Bus=="nats" && app.UseJS {
    js.AddStream(&nats.StreamConfig{
        Name: "EDITOR_UPDATES", Subjects: []string{"editor.updates"},
        MaxMsgsPerSubject: 10, MaxAge: 24*time.Hour,
    })
    js.AddStream(&nats.StreamConfig{
        Name: "EDITOR_TREE", Subjects: []string{"editor.tree"},
        MaxMsgsPerSubject: 10, MaxAge: 24*time.Hour,
    })
}
```

### 14.5 Client remains unchanged

The browser still connects to your local SSE endpoint. Multi‑node fan‑out happens server‑side via NATS and each node fans into its own local SSE hub.

### 14.6 Deploy matrix

* **Single node dev**: `BUS=local` (default). Nothing to change.
* **Multi node**: set `BUS=nats NATS_URL=nats://nats:4222` on **each** editor server. Optionally `NATS_JS=1` if your cluster has JetStream.
* **Mixed**: some nodes local‑only; others NATS‑backed. They’ll still serve local editors; NATS nodes will synchronize across regions.

### 14.7 Loop prevention & ordering

* We don’t republish messages received from NATS; we only fan‑in to local SSE.
* If you need strict ordering, enable JetStream and consume with a **durable**; here we use fire‑and‑forget for simplicity.

### 14.8 Multi‑tenant namespacing (optional)

Prefix subjects with a tenant key: `tenantA.editor.updates`, `tenantA.editor.tree`. Add `-tenant` flag/env and compute subjects accordingly.

---

**Result:** Flip one env/flag to propagate edits and tree changes across all your nodes via NATS/JetStream, without changing the client or Hugo setup. Turn it off to run completely self‑contained.

---

## 15) Git History + Revert (go-git), with Ordering

> Record every save as a Git commit (no external Git binary), show a sidebar history, view diffs, and revert to any revision. Also ensure **ordering** of commits even with NATS fan‑out.

### 15.1 Add go-git deps

```bash
# in server/ directory
go get github.com/go-git/go-git/v5
```

### 15.2 Repo bootstrap

On server start, init (or open) a repo at `../site/` and set a default author.

```go
import (
    // ...existing
    git "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/config"
    "github.com/go-git/go-git/v5/plumbing/object"
)

var repo *git.Repository

func openOrInitRepo() *git.Repository {
    r, err := git.PlainOpen("..") // repo root at ../ (contains site/)
    if err == git.ErrRepositoryNotExists {
        r, err = git.PlainInit("..", false)
        if err != nil { log.Fatalf("git init: %v", err) }
        // set default branch config if needed
        _, _ = r.CreateRemote(&config.RemoteConfig{Name:"origin",URLs:[]string{}})
    } else if err != nil { log.Fatalf("git open: %v", err) }
    return r
}
```

Call in `main()` after flags/env:

```go
repo = openOrInitRepo()
```

### 15.3 Commit helper + ordering

We serialize writes/commits with a global mutex to avoid index.lock races and preserve ordering under concurrent NATS events.

```go
var commitMu sync.Mutex

type authorInfo struct{ Name, Email string }

func gitCommit(pathRel, msg string, author authorInfo) error {
    commitMu.Lock(); defer commitMu.Unlock()

    w, err := repo.Worktree()
    if err != nil { return err }
    // ensure file is added relative to repo root
    if _, err := w.Add(pathRel); err != nil { return err }
    // also add the tree cache files (optional)

    if author.Name == "" { author = authorInfo{Name:"Editor", Email:"editor@local"} }

    _, err = w.Commit(msg, &git.CommitOptions{
        Author: &object.Signature{Name: author.Name, Email: author.Email, When: time.Now()},
        Committer: &object.Signature{Name: author.Name, Email: author.Email, When: time.Now()},
    })
    return err
}
```

**Path mapping:** for a content file `site/content/foo/bar.md`, pass `"site/content/foo/bar.md"` as `pathRel`.

> **Ordering note:** When `BUS=nats`, multiple nodes may save the same doc. Prefer **single‑writer** (one region) or enable JetStream and only the **consumer** that materializes to disk performs the Git commit. In this spec, we keep local commits for direct edits; remote updates coming via NATS update the preview only (no commit), unless you run a dedicated **committer** node.

### 15.4 Capture author from request

Very small helper to read author from headers, cookies, or fall back to defaults.

```go
func authorFrom(r *http.Request) authorInfo {
    a := authorInfo{ Name: r.Header.Get("X-User-Name"), Email: r.Header.Get("X-User-Email") }
    if a.Name == "" { a.Name = "Editor" }
    if a.Email == "" { a.Email = "editor@local" }
    return a
}
```

### 15.5 Wire commits into save handlers

For the HTML‑first `/save` (mirrors to Markdown) and `/api/save-page` (front‑matter + body), commit the file under `site/content/...`.

```go
// in /save handler, after writing mdPath
rel := filepath.ToSlash(filepath.Join("site","content","_editor", req.DocID+".md"))
if err := gitCommit(rel, "edit "+req.DocID, authorFrom(r)); err != nil { log.Printf("git commit: %v", err) }

// in /api/save-page (you will implement), after writing the target file
// suppose target path is absPath; compute rel from repo root ".."
rel, _ := filepath.Rel("..", absPath)
rel = filepath.ToSlash(rel)
if err := gitCommit(rel, "save "+rel, authorFrom(r)); err != nil { log.Printf("git commit: %v", err) }
```

### 15.6 History endpoint

Return recent commits and per‑file history.

```go
// GET /api/history?path=site/content/foo/bar.md&n=50
http.HandleFunc("/api/history", func(w http.ResponseWriter, r *http.Request){
    path := r.URL.Query().Get("path")
    nStr := r.URL.Query().Get("n"); if nStr=="" { nStr = "50" }
    n, _ := strconv.Atoi(nStr)

    type item struct{ Hash, Message, Author, Email, When string }
    var out []item

    // if path provided, walk file history; else recent repo commits
    if path != "" {
        // simple approach: iterate all commits and check for path changes
        logIter, err := repo.Log(&git.LogOptions{FileName: &path})
        if err != nil { http.Error(w, err.Error(), 500); return }
        defer logIter.Close()
        for i:=0; i<n; i++ { c, err := logIter.Next(); if err!=nil { break }
            out = append(out, item{ Hash: c.Hash.String(), Message: c.Message, Author: c.Author.Name, Email: c.Author.Email, When: c.Author.When.Format(time.RFC3339) })
        }
    } else {
        logIter, err := repo.Log(&git.LogOptions{})
        if err != nil { http.Error(w, err.Error(), 500); return }
        defer logIter.Close()
        for i:=0; i<n; i++ { c, err := logIter.Next(); if err!=nil { break }
            out = append(out, item{ Hash: c.Hash.String(), Message: c.Message, Author: c.Author.Name, Email: c.Author.Email, When: c.Author.When.Format(time.RFC3339) })
        }
    }
    writeJSON(w, out)
})
```

### 15.7 Diff endpoint

Show a unified diff between two commits for a file.

```go
// GET /api/diff?path=site/content/foo/bar.md&a=<hash>&b=<hash>
http.HandleFunc("/api/diff", func(w http.ResponseWriter, r *http.Request){
    path := r.URL.Query().Get("path")
    ah := r.URL.Query().Get("a")
    bh := r.URL.Query().Get("b")
    if path=="" || ah=="" || bh=="" { http.Error(w, "args", 400); return }

    aCommit, err := repo.CommitObject(plumbing.NewHash(ah))
    if err != nil { http.Error(w, err.Error(), 500); return }
    bCommit, err := repo.CommitObject(plumbing.NewHash(bh))
    if err != nil { http.Error(w, err.Error(), 500); return }

    aTree, _ := aCommit.Tree(); bTree, _ := bCommit.Tree()
    patch, err := aTree.Patch(bTree)
    if err != nil { http.Error(w, err.Error(), 500); return }

    // filter to file
    // (simple emit full patch; client can slice by filename)
    w.Header().Set("Content-Type","text/plain; charset=utf-8")
    w.Write([]byte(patch.String()))
})
```

*Add imports:*

```go
"github.com/go-git/go-git/v5/plumbing"
"strconv"
```

### 15.8 Revert endpoint

Reset a single file to the content at a given commit.

```go
// POST /api/revert { path: "site/content/foo/bar.md", hash: "..." }
http.HandleFunc("/api/revert", func(w http.ResponseWriter, r *http.Request){
    if r.Method != http.MethodPost { http.Error(w, "method", 405); return }
    var req struct{ Path, Hash string }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, err.Error(), 400); return }

    c, err := repo.CommitObject(plumbing.NewHash(req.Hash))
    if err != nil { http.Error(w, err.Error(), 500); return }
    f, err := c.File(req.Path)
    if err != nil { http.Error(w, err.Error(), 404); return }
    reader, err := f.Blob.Reader()
    if err != nil { http.Error(w, err.Error(), 500); return }
    defer reader.Close()
    b, _ := io.ReadAll(reader)

    // write file
    if err := os.WriteFile(req.Path, b, 0o644); err != nil { http.Error(w, err.Error(), 500); return }

    // commit revert
    if err := gitCommit(req.Path, "revert "+filepath.Base(req.Path)+" to "+req.Hash[:7], authorFrom(r)); err != nil { log.Printf("git: %v", err) }

    // broadcast updates
    html := string(b) // if your files are Markdown, convert to HTML here before broadcasting
    publishDoc(strings.TrimPrefix(strings.TrimPrefix(req.Path, "site/content/"),"/"), html)
    w.WriteHeader(204)
})
```

> If your repo stores Markdown, convert to HTML before `publishDoc` using your existing mirror (`htmlToMarkdown` inverse) or render via Hugo (slower). For fast preview, keep the HTML payload as the editor’s canonical broadcast while Markdown stays on disk.

### 15.9 Client: History panel

Add a collapsible History panel beside the toolbar and wire actions.

```html
<details class="toolbar" style="margin-top:8px"><summary>History</summary>
  <div id="history" class="preview" style="max-height:240px; overflow:auto"></div>
  <div style="padding:8px;display:flex;gap:8px">
    <input id="revA" placeholder="A hash" />
    <input id="revB" placeholder="B hash" />
    <button onclick="showDiff()">Diff</button>
    <button onclick="revertTo()">Revert to A</button>
  </div>
  <pre id="diff" style="padding:8px;white-space:pre-wrap"></pre>
</details>
<script>
async function loadHistory(){
  const path = 'site/'+('content/'+CURRENT_PATH).replace(/\/g,'/');
  const list = await fetch('http://localhost:8080/api/history?path='+encodeURIComponent(path)+'&n=50').then(r=>r.json());
  const el = document.getElementById('history');
  el.innerHTML = list.map(it=>`<div><code>${it.Hash.slice(0,7)}</code> — ${it.Message} <small>by ${it.Author} @ ${it.When}</small></div>`).join('');
  if(list[0]){ revA.value = list[0].Hash; }
}
async function showDiff(){
  const path = 'site/'+('content/'+CURRENT_PATH).replace(/\/g,'/');
  const txt = await fetch(`http://localhost:8080/api/diff?path=${encodeURIComponent(path)}&a=${encodeURIComponent(revA.value)}&b=${encodeURIComponent(revB.value)}`).then(r=>r.text());
  document.getElementById('diff').textContent = txt;
}
async function revertTo(){
  const path = 'site/'+('content/'+CURRENT_PATH).replace(/\/g,'/');
  await fetch('http://localhost:8080/api/revert', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({path, hash: revA.value})});
  await setDoc(CURRENT_PATH); // reload editor/preview
  await loadHistory();
}
// load on start and whenever doc changes
const _orig2 = setDoc; setDoc = async function(p){ await _orig2(p); setTimeout(loadHistory, 80); }
loadHistory();
</script>
```

### 15.10 Advanced ordering strategies (optional)

* **Single‑writer per doc:** route saves (via NATS) to a leader for that path; only the leader writes & commits; others just broadcast.
* **JetStream consumer per doc/section:** create durable consumers keyed by section so messages for a doc are processed FIFO.
* **Vector clock in headers:** include `version` header; if a node receives an older version than disk, skip commit and only preview.
* **Git merge queue:** stage remote changes to a branch, periodically fast‑forward merge; show “merge required” in UI.

---

**Outcome:** Every edit is versioned, users can browse history, diff revisions, and revert safely. With the mutex + (optional) JetStream consumer, commit **ordering** remains deterministic even under concurrent, multi‑node edits.









