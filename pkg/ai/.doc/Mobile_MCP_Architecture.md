# Mobiel MCP

## Overview
Each Gateway on Users phones has an AI Client, and we need to expose our MCP to them.

**Main Goal**: Enable Gemini app on peoples mobile to find and transact with my MCP server, so customers can easily do things.

**Reference**: https://www.kimi.com/chat/d2i70m4n907avbefvi90

## MCP Server Platform Support

### Gemini + MCP Support Status

| Client / Platform | MCP Support? | How to Configure |
|-----------------------|------------------|-----------------------|
| **Gemini Desktop (Tome)** | ✅ Yes | [Download Tome](https://github.com/runebookai/tome/releases) → Add Gemini API key → Install MCP servers from Smithery registry |
| **Android Studio Gemini Agent** | ✅ Yes | Add `mcp.json` file with your server → Gemini agent uses MCP tools |
| **Gemini Code Assist (VS Code)** | ✅ Yes | Edit `~/.gemini/settings.json` → Add your MCP server under `"mcpServers"` |
| **Custom Gemini Apps (Python/Node)** | ✅ Yes | Use `google-genai` SDK + `mcp` SDK to connect to your MCP server |

### What This Means for You

- **Your MCP server can be discovered** by Gemini-powered clients if you:
  - Publish it to a registry like [Smithery](https://smithery.ai)
  - Or provide clear instructions for manual setup (e.g., `npx your-mcp-server`, Docker, etc.)

- **Customers using Gemini** (on desktop, Android Studio, or VS Code) can:
  - Install your MCP server
  - Use natural language to interact with your server's tools
  - Transact or perform actions via your server

### Next Steps for You

1. **Package your MCP server** (e.g., as an npm package or Docker container)
2. **Publish it to Smithery** or GitHub with a README
3. **Provide a one-liner install command**, like:
   ```bash
   npx -y your-mcp-server
   ```
4. **Add a sample config snippet** for Gemini users, like:
   ```json
   "mcpServers": {
     "your-server": {
       "command": "npx",
       "args": ["-y", "your-mcp-server"]
     }
   }
   ```

## Google Discovery Strategy

Google will **only** surface a web page if it can (1) **crawl** it, (2) **understand** it, and (3) **trust** that it satisfies search intent better than the other pages on the topic.

For an MCP server that lives *inside* a mobile app or behind a custom protocol, Google **cannot** do any of those three things natively—so it will **never** appear in web search results.

### 1. Publish a discoverable web page
Create a lightweight public site (or GitHub README) that contains:

- A plain-English explanation of what the MCP server does  
- A link to the install command (e.g., `npx @your-org/mcp-server`)  
- A `/.well-known/mcp.json` endpoint (optional but nice) that returns the MCP manifest so tooling can auto-detect it

### 2. Make the page Google-friendly
- Register the domain in Google Search Console  
- Add a `sitemap.xml` and submit it in Search Console   
- Use descriptive `<title>` and `<meta name="description">` tags that include the keywords people will actually search for (`"my-service MCP server"`, `"Gemini MCP plugin"`, etc.)   
- Keep URLs short and keyword-rich (`https://my-service.dev/mcp`)   
- Ensure the page is mobile-first and loads fast   

### 3. Build trust & authority
- Host the server's source on GitHub with a README that mirrors the landing page  
- Encourage early users to link back to the landing page (backlinks are still a ranking factor)   
- Publish short tutorials or blog posts on Medium/Dev.to that include the keyword "MCP" and link back to the canonical page 

### 4. Provide a friction-less install path from the SERP
Once the page ranks, add a **one-click install snippet** so the user can copy-paste into their Gemini client:

```bash
npx -y @your-org/mcp-server
```

### 5. Optional: list on Smithery
Smithery.ai is becoming the de-facto registry for MCP servers. Listing there gives you an extra backlink and exposes the server to Gemini/Tome users who browse the catalog.

**Bottom line:** Google can **never** crawl the MCP server itself, but it *can* crawl a simple web page that tells users how to install or connect to it. Make that page the canonical source of truth, optimize it for the keywords people use, and Google will do the rest.

## Apple Platform Strategy

Apple is already moving on it, just with their own branding and rules.

### What Apple is shipping  
- **Xcode 16** (iOS 18 SDK) adds **"Swift Assist"** – an LLM that can call *local* and *remote* tools inside the IDE. Under the hood it uses the same JSON-RPC pattern as MCP (Apple calls it "Swift Tool Extensions").  
- **SiriKit Re-Write** (rumored for iOS 19) is expected to let third-party *App Intents* be surfaced to Siri and the new LLM stack the same way MCP servers surface tools to Gemini.  
- **App Intents** (the framework that powers Shortcuts, Siri, Spotlight) already supports schema-defined parameters & result types—essentially an MCP tool descriptor written in Swift.

### The discoverability problem is identical  
Just like Google, Apple's search (Spotlight, Siri, App Store) only indexes **web links** or **App Store listings**, not the tool bundle inside your app.  
To surface your MCP-equivalent tool to Apple users you need:  
- An App Store page (or TestFlight build) that contains the keyword "MCP" or "Swift Tool Extension"  
- A web page with the same content so Siri Suggestions can surface it  
- A `.well-known/apple-app-site-association` file so Spotlight can deep-link to the install flow inside your app

### Short-term cheat sheet  
- Wrap your MCP server in a minimal iOS app that simply registers one App Intent → resolves to your server.  
- Submit that shell app to the App Store with a title like *"MyService MCP Client"*.  
- In the description include the exact phrase "MCP server" so Apple's search index picks it up.

## Gateway Provider Integration

### 1. **Payment Gateways** (where money actually moves)  
These are the heavy-lifters that process cards, wallets, bank debits and local methods in 100+ countries.

| Provider | Reach | One-liner |
| --- | --- | --- |
| **Stripe** | 135+ currencies, 46 countries | De-facto API standard; every dev tool already has a Stripe SDK  |
| **PayPal / Braintree** | 200+ markets | Highest consumer trust; checkout conversion lift up to 44 %  |
| **Adyen** | 45+ countries, 150+ local methods | Single contract, unified API; used by Uber, Spotify, eBay  |
| **Checkout.com** | 150+ currencies | Fast onboarding, quote-based pricing; big in EU & MENA  |
| **Worldpay (FIS)** | 126+ currencies, 300+ APMs | Incumbent in travel & gaming; strong fraud stack  |
| **Amazon Pay** | 18 countries, 12 currencies | One-click checkout for 300 M Amazon accounts  |
| **2Checkout (Verifone)** | 200 markets | Popular for SaaS & digital goods; no monthly fees  |

> **Rule of thumb:** If you only have bandwidth for **two**, integrate **Stripe + PayPal/Braintree** first. They cover 80 % of global volume and are supported by every MCP-style "tool-calling" client today.

### 2. **Distribution / Discovery Gateways** (where users *find* your server)  
These are the new LLM "app stores" that surface MCP or MCP-like tools.

| Gateway | What it is | How to get listed |
| --- | --- | --- |
| **Smithery.ai** | De-facto MCP registry | Submit a manifest → one-line install for Gemini/Tome users |
| **OpenAI GPT Store** (Actions) | Custom GPTs that call HTTPS endpoints | Wrap your MCP as an Action, publish in the store |
| **Anthropic Claude "Computer Use"** | Experimental tool-use API | Same pattern—expose MCP via HTTPS, list in docs |
| **Microsoft Copilot Studio** | Enterprise plugin catalog | Register as a "connector" for Teams & 365 Copilot |
| **Apple Swift Assist Registry** (future) | Xcode 16 tool extensions | Package as Swift Package + App Intent → App Store |
| **WeChat Mini-Programs / Alipay Mini-Apps** | 1.3 B users, local wallets | If China is on your roadmap, wrap MCP calls in a Mini-App |

### Quick action checklist
1. **Payments**: Stripe + PayPal/Braintree integration (both have Node & Python SDKs that work inside an MCP server).  
2. **Discovery**: Publish your MCP manifest to **Smithery** today; keep the same HTTPS endpoints ready for OpenAI & Anthropic when they open wider tool catalogs.  
3. **Future-proofing**: Add a thin Swift Package that registers an App Intent → ready for Apple's inevitable "Swift Assist Store."

## Payment Implementation Details

### Payment flow explanation
The **consumer-facing Gemini app** that people already have on their phones is **not** going to let your MCP server "piggy-back" on its own payment rails.  
That app only accepts **funds into the user's Gemini exchange account** (bank, debit, PayPal, Apple Pay, etc.) so the user can buy crypto or pay for Gemini Advanced subscriptions.  
It does **not expose an API** for third-party servers to **charge the user's card or wallet** on your behalf.

### What *can* happen

1. **Your MCP server** (running wherever you host it) makes its own **payment-gateway calls** (Stripe, PayPal, Adyen, etc.).  
2. When the Gemini client invokes your MCP tool, you respond with a **checkout URL** or **payment-intent secret** that the client displays to the user.  
3. The user **leaves the Gemini chat**, taps the link, completes payment on your hosted page (or Apple Pay sheet, Stripe Checkout, etc.), then returns—exactly the same flow as any website checkout.

In short, Gemini will happily surface your tool and relay text, but **you—not Gemini—handle the money**; there's no built-in "pass-through" payment capability.

### Apple Pay integration
When your MCP server needs the user to pay, you can hand back a **Stripe** (or Braintree, Adyen, etc.) **Apple Pay payment link**.  
On an iPhone the Gemini app will open that link in Safari, the Safari tab triggers the **Apple Pay sheet**, and the user completes the transaction with Face ID/Touch ID.  
The money goes straight to your Stripe account; Gemini never sees the card details and never acts as a pass-through.

## China Market Strategy

You'll need a **separate stack** for China, because none of the Western rails (Stripe, Apple Pay, Google Pay) work there and the "Gemini app" itself isn't available in the mainland stores.

### 1. Discovery inside China  
- **WeChat Mini-Programs** (小程序) – 1.3 B MAU. Wrap your MCP calls in a Mini-Program; users find it via search or QR code.  
- **Alipay Mini-Apps** – 1 B MAU; same concept.  
- **Baidu ERNIE Bot plugins** – Baidu's LLM (ERNIE 4.0) now has an "agent/plugin" store. You can expose your MCP endpoints as HTTPS actions the same way you would for Gemini.

### 2. Payments inside China  
| Provider | Coverage | Notes |
|---|---|---|
| **WeChat Pay** | 1.2 B wallets | Mandatory for Mini-Programs; instant onboarding via Tencent Cloud. |
| **Alipay** | 1 B wallets | Same SDK pattern; supports face-pay at POS. |
| **UnionPay QuickPass** | All Chinese banks | Required if you want debit-card fallback. |
| **Stripe China** | Cross-border only | Lets foreigners charge CNY and settle in USD, but **users still need WeChat/Alipay to pay**. |

### 3. Technical flow  
1. Host an **HTTPS wrapper** of your MCP server in either Tencent Cloud (for WeChat) or Ant Cloud (for Alipay).  
2. Register the Mini-Program / Mini-App → get an **App ID**.  
3. Use the provider's **payment SDK** (`wx.requestPayment` for WeChat, `my.tradePay` for Alipay) to pop the native checkout sheet.  
4. Return the result back to the chat so the LLM can continue.

Bottom line: create **two MCP adapters**—one for Gemini/Apple Pay (rest-of-world) and one wrapped as a WeChat Mini-Program + WeChat Pay for China.

### China Legal Requirements

**No, you no longer *need* a JV for most software launches in 2025**, but whether you *choose* one depends on the exact nature of your product and your tolerance for risk.

#### The "Negative List" test  
China now uses a **"Negative List"** approach:  
- If your software is **not** on the list (most SaaS, cloud, AI tools, and MCP-style integrations are **not**), you can set up a **Wholly-Foreign-Owned Enterprise (WFOE)** or simply operate from offshore with a **cross-border VATS filing**.  
- Only **restricted sectors** (e.g., certain telecom services, value-added media, fintech, or anything touching "critical information infrastructure") still **require** a JV or special license.

#### WFOE vs JV in practice  
| Structure | Ownership | Time-to-Market | Typical Use Case |
|---|---|---|---|
| **WFOE** | 100 % foreign | 6–10 weeks after docs | SaaS, dev tools, MCP servers not on Negative List  |
| **JV** | ≥ 25 % local partner | 4–6 months | Regulated telco, payment processing, or when local partner brings essential licenses  |

#### Common compliant short-cuts  
- **Pure offshore delivery**: Host your MCP server outside China, collect WeChat/Alipay via Stripe HK or Adyen. Users pay in CNY; you settle in USD. No China entity needed.  
- **Representative Office (RO)**: Good for market research or liaison, but **cannot take payments**.  
- **Mini-Program-only play**: Tencent **does not** require a JV to publish a Mini-Program; a WFOE or even a **Hong Kong company** suffices, as long as you VAT-register and appoint a local support entity (often a 1-person agency).

#### When a JV still makes sense  
- Your product **stores or processes** Chinese user data **inside China** (then you need a **Cybersecurity Multi-Level Protection Scheme (MLPS)** certificate—easier with a JV partner).  
- You want to sell **directly to SOEs** or **government clients** who insist on a domestic entity.  
- You need **ICP, EDI, or other telecom permits** that are hard to obtain without a local majority partner.

#### Rule of thumb  
Start with the **Negative List check** → if you're **off the list**, open a **WFOE** (or skip it and operate offshore) and launch quickly.  
Upgrade to a JV **only if** you hit a regulatory wall or a strategic partner brings irreplaceable value.

(Always run the final decision past a China-licensed attorney; the list can change mid-year.)

## Kimi Integration

**Kimi**, a large language model trained by **Moonshot AI**, already supports the MCP protocol, so your server can:

1. Expose JSON-RPC tools exactly like any other MCP server.
2. Be invoked directly by Kimi (or any other Gemini/Tome client) once the user adds your server to their MCP configuration.
3. Receive tool calls, return structured responses, and trigger follow-up prompts.

### Test Configuration
Host your MCP server behind HTTPS (or run it locally) and add a stanza like this to the user's `mcp.json`:

```json
"mcpServers": {
  "your-server": {
    "command": "npx",
    "args": ["-y", "your-mcp-server"],
    "env": { "OPENAI_API_KEY": "…" /* or any secrets */ }
  }
}
```

Kimi will then be able to call any tool you expose (`list_tools`, `call_tool`, etc.) in the same session.

## Technical Implementation - Crush CLI

**Crush is a Go CLI** (built with Charm's Bubble Tea framework), and you have **two easy ways** to run it on your servers:

### 1. **Binary CLI** (what you're using locally)  
- Just drop the same static binary on your server:  
  ```bash
  wget https://github.com/charmbracelet/crush/releases/latest/download/crush_linux_amd64
  chmod +x crush_linux_amd64
  ./crush_linux_amd64 --help
  ```
- Zero dependencies—works on any modern Linux distro.

### 2. **Go package / library**  
- Import it directly in your own Go server:
  ```go
  import "github.com/charmbracelet/crush/pkg/crush"
  ```
- Then embed the TUI loop or the core session engine inside your long-running service.  
- You can expose the same MCP endpoints you already wrote while re-using Crush's session, model-switching and LSP-integration code.

### State Management Challenge
If you simply spawn the **Crush CLI binary** for every inbound HTTP request, each request will get its **own fresh process** and therefore **no shared state**.  
That's perfectly fine for stateless MCP tools, but it breaks anything that relies on long-lived config, auth tokens, or memory caches inside the CLI.

To keep **shared state** across users you have two choices:

#### 1. **Keep the CLI running as a long-lived daemon**
Run the Crush binary once as a **persistent server** (or wrap it in a small Go supervisor):

```bash
./crush_linux_amd64 --serve --port 8080
```

Then your HTTP handler just **proxies the MCP JSON-RPC** to that single process.  
All requests hit the same instance → state (tokens, model context, etc.) is retained.

#### 2. **Embed the Crush Go pkg inside your own Go HTTP server**
Import the package and instantiate **one global `crush.Session`** (or whatever struct holds the state) at server start-up:

```go
var sess = crush.NewSession()

http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
    // forward JSON-RPC to sess.HandleRPC(...)
})
```

Every HTTP call now shares the same session object, so state persists across users.

## Multi-Binary Process Management

With 20 different binaries you have two proven patterns to choose from—pick the one that matches your latency and ops budget.

### 1. One-process-per-binary, stateless
- Spawn each binary as a **long-running daemon** on its own port (or Unix socket).  
- Front them with a tiny **reverse proxy / multiplexer** (20-line Go file with `net/http/httputil` or `caddy`).  
  - Incoming MCP request → proxy → correct daemon.  
  - Each daemon keeps its own memory state, config, tokens, etc.  
- Pro: zero code changes to the binaries.  
- Con: 20 OS processes to monitor/restart.

### 2. Embed everything in one Go server
- Import all 20 packages (`github.com/you/bin1`, `bin2`, …) into a single Go binary.  
- Instantiate one long-lived **struct per tool** at start-up and register each under its own MCP tool name:  

```go
mcp.Register("bin1", bin1.NewSession())
mcp.Register("bin2", bin2.NewSession())
// …
```

- Expose a **single `/mcp`** endpoint; the MCP router dispatches to the right tool.  
- Pro: one process, shared infra (logging, metrics, graceful reload).  
- Con: you must ensure the 20 packages don't have conflicting globals or init side-effects.

### Quick pick guide
- Need to ship **today** and the binaries already work? → Pattern 1 (daemons + proxy).  
- Want **single-binary, single-port, zero external processes**? → Pattern 2 (embed).

Either way, Kimi (or any MCP client) will only ever see one endpoint URL; you control how the 20 tools are mapped behind the scenes.

## Process Supervision Strategy

Once you have 20 long-lived daemons, **reliable process supervision** is the only thing standing between "just works" and "page me at 3 a.m."

### 1. Pick a supervisor (single-node)
| Option | Install | Pros | Cons |
|---|---|---|---|
| **systemd** | already on every modern Linux | native, logs via `journalctl`, auto-restart on crash | tied to OS |
| **supervisord** | `apt install supervisor` | one config file for 20 procs, web UI | Python, extra package |
| **runit / s6** | `apt install runit` | ultra-light, milliseconds restarts | learning curve |

Example **systemd** unit you can template:

```ini
# /etc/systemd/system/crush-bin@.service
[Unit]
Description=Crush binary %i
After=network.target

[Service]
Type=simple
User=crush
ExecStart=/opt/crush/bin/%i --serve --port %i
Restart=always
RestartSec=2
Environment=ENV_FILE=/opt/crush/%i.env

[Install]
WantedBy=multi-user.target
```

Enable all 20 with one loop:

```bash
for p in {8000..8019}; do
    systemctl enable --now crush-bin@$p
done
```

### 2. Health & restart hooks
- Add a **liveness endpoint** (`/healthz`) to each binary or just let systemd restart on exit code ≠ 0.  
- Use **log aggregation** (Vector, Fluent Bit) to forward `stdout` → Loki/ELK so you see failures in one place.  
- Optional: expose **Prometheus metrics** (process_uptime_seconds, last_mcp_error) and let AlertManager page you.

### 3. Multi-node / container path
If you expect > 1 machine:

```yaml
# docker-compose.yml fragment
services:
  bin1:
    image: crush:latest
    command: ["./bin1", "--serve", "--port", "8080"]
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/healthz"]
  bin2:
    …
```

Then use **Docker Compose v2 restart policies** or **Kubernetes Deployments** with a single `restartPolicy: Always`.

### TL;DR
1. Wrap each binary in a systemd (or supervisor) unit.  
2. `systemctl enable --now` all 20 in a loop.  
3. Ship logs and metrics to one dashboard.  
You'll get crash-restart, zero-downtime deploys, and a single `journalctl -u crush-bin@8003 -f` when something breaks.

## Go Supervisor Implementation

If you want "OS-level" supervision with a **single Go program** that itself starts, monitors, and restarts the 20 binaries, treat that Go program as **PID 1** (or close to it) and give it the same responsibilities `systemd` would have:

1. fork/exec each binary  
2. capture/rotate its stdout/stderr  
3. restart on crash or health-check failure  
4. expose a control socket or HTTP API for "stop / start / reload"

### A. Go-built **init replacement** (PID 1 inside a container)
Use a tiny Go supervisor library:

```go
import "github.com/oklog/run"

g := run.Group{}
for _, bin := range binaries {
    cmd := exec.Command(bin.Path, bin.Args...)
    g.Add(func() error { return cmd.Run() },
          func(err error) { _ = cmd.Process.Kill() })
}
log.Fatal(g.Run())
```

- Drop this binary into a **scratch** or **distroless** container with your 20 static binaries.  
- Kubernetes / Nomad starts the container; your Go supervisor is PID 1 and restarts anything that dies.

### B. Go **service supervisor** (runs under the real init, but owns the 20 processes)
Keep the host's init (systemd) for boot, but delegate **only the 20 binaries** to your Go supervisor.

- Your supervisor **daemonizes itself** (or runs as a systemd service).  
- Uses **os/exec** + **os/signal** + **fsnotify** to manage the subprocesses.  
- Writes pidfiles, rotates logs, and exposes a REST or Unix-socket control plane.

Example skeleton:

```go
package main

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    go supervise(ctx, binaries)
    // graceful shutdown on SIGTERM/SIGINT
    <-interrupt()
    cancel()
}

func supervise(ctx context.Context, bins []Bin) {
    for _, b := range bins {
        go func(b Bin) {
            for {
                cmd := exec.CommandContext(ctx, b.Path, b.Args...)
                log.Printf("starting %s", b.Name)
                if err := cmd.Run(); err != nil {
                    log.Printf("%s exited: %v, restarting in 2s", b.Name, err)
                }
                select {
                case <-ctx.Done():
                    return
                case <-time.After(2 * time.Second):
                }
            }
        }(b)
    }
}
```

### Packaging checklist
- Static-compile your 20 binaries (`CGO_ENABLED=0 go build`).  
- Static-compile the supervisor.  
- Place them in `/opt/crush/bin/` inside the same image or host dir.  
- Ship a **Dockerfile** or **systemd unit** that starts the supervisor; the supervisor itself owns the 20 children.

This keeps you 100% in Go, gives you OS-level supervision semantics, and still lets you deploy anywhere from a bare VM to Kubernetes.

## Goreman as Alternative

**goreman is good enough** for this job.

- It's a **single-binary, Go re-implementation** of Foreman, so you get the familiar `Procfile` workflow without Ruby.  
- Handles **start/stop/restart** and **log aggregation** for all 20 processes in one window.  
- Lets you scale individual lines with `-c` (e.g., `goreman start -c bin1=2,bin2=1`).  
- Works fine in **production** if you wrap it in systemd or run it inside a container; just make sure your `Procfile` uses absolute paths and sets `GOREMAN_PORT` if the binaries need distinct ports.

### Single Binary Approach
You can keep it down to **one** OS-level service if you want.  
Just fold the goreman **Procfile** (or equivalent) **inside** your single Go binary, and have that binary become the one and only systemd service:

```go
// main.go
func main() {
    if len(os.Args) > 1 && os.Args[1] == "supervisor" {
        runSupervisor() // embed the goreman logic or your own
    } else {
        runMCP()        // normal MCP server / tool entrypoint
    }
}
```

systemd unit:

```ini
[Service]
ExecStart=/opt/crush/crush supervisor
Restart=always
```

Now the OS starts **exactly one program**—your own binary—which in turn manages the 20 children.

## Cross-Platform Deployment

Keep **one cross-platform supervisor binary** (your own Go program) that embeds the 20 tools and **works the same way on Windows, macOS and Linux**.

### 1. Build one static executable  
```bash
GOOS=windows GOARCH=amd64 go build -o crush.exe
GOOS=darwin  GOARCH=amd64 go build -o crush-darwin
GOOS=linux   GOARCH=amd64 go build -o crush-linux
```
Embed the 20 static binaries with `go:embed` or ship them alongside in a single directory; your supervisor finds them at runtime.

### 2. Ship the same supervisor logic everywhere  
- **Windows** – register as a **Windows Service** via `golang.org/x/sys/windows/svc`.  
- **macOS** – install as a **LaunchDaemon** (`launchctl load /Library/LaunchDaemons/com.mycrush.supervisor.plist`).  
- **Linux** – install as a **systemd service** (unit file above).

### 3. One file, one installer script  
e.g. on Windows:  
```powershell
sc.exe create Crush binPath= "C:\Program Files\Crush\crush.exe supervisor"
```

So you end up with **one artefact per OS** that starts itself at boot and spawns the 20 binaries—no external goreman or extra services required.

## Fly.io China Strategy

Here's the practical recipe for "**Hong Kong as a springboard into mainland China with Fly.io**":

| Step | What & Why |
|---|---|
| **1. Pick Fly.io region `hkg`** | It's physically in **Hong Kong** (~2 ms to China Telecom & CU backbones) so packets don't leave Greater China. |
| **2. Enable Anycast IP** | `fly ips allocate-v4 --shared --region hkg` gives you an IP that is **not blocked** by the GFW for standard HTTPS (443). |
| **3. Use a CN-friendly CDN in front (optional)** | Put **Tencent EdgeOne** or **Alibaba DCDN** in front of your Fly app.  These CDNs have PoPs inside the mainland and will pull from your HK origin.  Users hit the mainland PoP → HK origin; you get **<50 ms latency** and **zero ICP filing** because content is considered "foreign origin". |
| **4. Expose HTTPS only** | GFW throttles non-443 traffic.  Terminate TLS at Fly's edge or at the CDN. |
| **5. Cross-border payment flow** | HK entity + **Stripe HK** or **Adyen HK** can charge CNY and settle in USD.  The user still pays via **WeChat Pay / Alipay** through the same checkout page; the money never transits the mainland rails. |
| **6. DNS optimisation** | Point `api.yourapp.com` to **both** the CDN edge (for mainland traffic) and the Fly Anycast IP (for RoW).  TTL 60 s so you can fail over instantly. |

### Result:  
- **Mainland users** reach your HK-hosted MCP server in ~20–40 ms, HTTPS only, **no ICP licence** required.  
- **You** run one Fly app in `hkg`, one HK entity, one Stripe HK account.

## Summary

- **Rest of world**: MCP server → Smithery + Stripe/Apple Pay checkout links.  
- **China**: WeChat Mini-Program + WeChat Pay + optional WFOE if you want to store data locally.