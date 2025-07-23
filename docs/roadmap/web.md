# `web`: Web Application Component

This document outlines the design for the `web` component of the `infra` system, which primarily serves a web application using the Datastar framework and enables Add to Home Screen (A2HS) functionality. It leverages an embedded NATS server for internal messaging and a Chi router for HTTP handling.

### 1. Vision

To provide a lightweight, responsive, and easily deployable web interface for `infra` that leverages modern web capabilities, specifically A2HS, to offer a native-app-like experience. The web component will integrate seamlessly with the `infra` backend, utilizing Datastar's reactive capabilities for dynamic content and interactions, and NATS for robust internal communication.

### 2. Core Functionality

*   **Datastar Integration:** The web application is built using the `datastar-go` framework, leveraging its server-driven UI and signal-based reactivity for efficient data flow and user experience.
*   **Embedded NATS Server:** An `embeddednats` server is utilized for internal messaging, providing a robust and self-contained communication bus within the `infra` application. This includes JetStream for persistent message streams.
*   **Chi Router:** The `go-chi/chi` router is used for efficient and flexible HTTP request routing.
*   **Add to Home Screen (A2HS) Support:** Enable users to install the `infra` web application directly to their device's home screen, providing quick access and a more integrated feel.
*   **Backend Communication:** Establish robust communication channels with the `infra` Go backend for data exchange and command execution, primarily via NATS.
*   **Static Asset Serving:** Efficiently serve static web assets (HTML, CSS, JavaScript, images) using Go's `//go:embed` directive.

### 3. Architectural Flow

The primary data flow within the web component is:

**NATS --> DataStar --> DataStarUI --> Browser**

This means data is published to NATS, consumed by Datastar, rendered via DatastarUI (if applicable), and then presented in the browser.

### 4. Add to Home Screen (A2HS) Requirements

To enable A2HS, the web component must implement the following:

*   **Web App Manifest:** A `manifest.json` file will be provided, containing essential metadata about the web application (e.g., name, short name, icons, start URL, display mode, theme color). This manifest is crucial for browsers to understand how to present the web app when added to the home screen.
*   **Service Worker:** A `service-worker.js` script will be implemented to:
    *   **Cache Assets:** Cache static assets and potentially dynamic content to enable offline capabilities and improve loading performance.
    *   **Intercept Network Requests:** Control how network requests are handled, allowing for offline fallback and custom caching strategies.
    *   **Push Notifications (Future Consideration):** While not an initial requirement, the Service Worker lays the groundwork for future push notification capabilities.
*   **HTTPS:** A2HS functionality requires the web application to be served over HTTPS. This will be a deployment consideration.

### 5. Integration with Datastar

Datastar's signal pattern will be instrumental in facilitating A2HS and other web functionalities:

*   **Dynamic Manifest Generation:** The `manifest.json` could potentially be dynamically generated or served by the Go backend, allowing for configuration based on backend state or environment variables.
*   **Service Worker Registration:** The web application will register the `service-worker.js` using standard JavaScript, and Datastar can be used to trigger or manage this registration process based on user interaction or application state.
*   **Backend-Driven UI Updates:** Datastar's core strength will be used to update the web UI based on events or data changes originating from the `infra` Go backend, providing a highly interactive experience.
*   **Schema Evolution:** `natsrpc` will be used for defining types with Protobufs, enabling schema evolution at runtime for NATS-based communication.
*   **GUI Components:** `datastarui` will be utilized for building graphical user interface components.

### 6. Technical Considerations

*   **Go Web Framework:** The `web` folder contains the Go web server code, utilizing `net/http` and `github.com/go-chi/chi/v5` for routing.
*   **Templating:** HTML templates are used to render the web pages, with `//go:embed` for embedding `index.html`.
*   **Build Process:** A build process will be defined to compile Go code, bundle static assets, and generate the necessary A2HS files.

### 7. Go Package API (Conceptual)

The `web` package should expose a simple API for starting the web server.

```go
package web

// StartServer initializes and starts the web application server.
func StartServer() error
```
