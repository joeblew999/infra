package web

import (
	"net/http"
)

// Page handlers for the main web application
// These handlers render the main navigation pages

func (app *App) renderHomePage(w http.ResponseWriter, r *http.Request) {
	app.renderer.RenderHomePage(w, r)
}

func (app *App) handle404(w http.ResponseWriter, r *http.Request) {
	app.renderer.Render404Page(w, r)
}
