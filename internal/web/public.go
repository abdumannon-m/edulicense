package web

import (
	"fmt"
	"net/http"
	"strings"

	"edu-license/internal/app"
	"github.com/go-chi/chi/v5"
)

func (s *Server) home(locale string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		messages := app.MessagesForLocale(locale)
		path := "/"
		if locale == "uz" {
			path = "/uz"
		}
		data := app.PublicPageData{
			BaseURL:      s.cfg.AppBaseURL,
			Path:         path,
			CanonicalURL: canonical(s.cfg.AppBaseURL, path),
			Messages:     messages,
			Title:        messages.MetaTitle,
			Description:  messages.MetaDescription,
		}
		s.renderer.Render(w, http.StatusOK, "public_home", data)
	}
}

func (s *Server) privacy(locale string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		messages := app.MessagesForLocale(locale)
		path := "/privacy"
		if locale == "uz" {
			path = "/uz/privacy"
		}
		data := app.PublicPageData{
			BaseURL:      s.cfg.AppBaseURL,
			Path:         path,
			CanonicalURL: canonical(s.cfg.AppBaseURL, path),
			Messages:     messages,
			Title:        messages.Privacy.Title + " · " + messages.BrandShort,
			Description:  messages.MetaDescription,
		}
		s.renderer.Render(w, http.StatusOK, "public_privacy", data)
	}
}

func (s *Server) verify(locale string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		cert, err := s.store.CertificateBySlug(r.Context(), slug)
		if err != nil {
			s.notFound(w, r)
			return
		}
		messages := app.MessagesForLocale(locale)
		path := "/verify/" + slug
		if locale == "uz" {
			path = "/uz" + path
		}
		data := app.PublicPageData{
			BaseURL:      s.cfg.AppBaseURL,
			Path:         path,
			CanonicalURL: canonical(s.cfg.AppBaseURL, path),
			Messages:     messages,
			Title:        messages.Verify.MetaTitle + " · " + cert.Institution + " · " + messages.BrandShort,
			Description:  messages.Verify.Intro,
			Certificate:  cert,
		}
		s.renderer.Render(w, http.StatusOK, "public_verify", data)
	}
}

func (s *Server) robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "User-agent: *\nAllow: /\nDisallow: /admin\n\nSitemap: %s/sitemap.xml\n", strings.TrimRight(s.cfg.AppBaseURL, "/"))
}

func (s *Server) sitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	paths := []string{"/", "/uz", "/privacy", "/uz/privacy"}
	fmt.Fprintln(w, `<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Fprintln(w, `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	for _, path := range paths {
		fmt.Fprintf(w, "<url><loc>%s</loc></url>\n", canonical(s.cfg.AppBaseURL, path))
	}
	fmt.Fprintln(w, `</urlset>`)
}
