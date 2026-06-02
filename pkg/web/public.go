package web

import (
	"fmt"
	"net/http"
	"strings"

	"edu-license/pkg/app"
	"edu-license/pkg/store"
	"github.com/go-chi/chi/v5"
	"github.com/skip2/go-qrcode"
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
		certificate, err := s.store.CertificateBySlug(r.Context(), slug)
		if err != nil {
			if store.IsNotFound(err) {
				http.Error(w, "certificate not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		messages := app.MessagesForLocale(locale)
		path := "/verify/" + certificate.Slug
		if locale == "uz" {
			path = "/uz" + path
		}
		data := app.PublicPageData{
			BaseURL:      s.cfg.AppBaseURL,
			Path:         path,
			CanonicalURL: canonical(s.cfg.AppBaseURL, path),
			Messages:     messages,
			Title:        "Certificate verification · " + messages.BrandShort,
			Description:  "Verify an Edu License certificate by QR code or certificate ID.",
			Certificate:  certificate,
		}
		s.renderer.Render(w, http.StatusOK, "public_verify", data)
	}
}

func (s *Server) certificateQR(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	certificate, err := s.store.CertificateBySlug(r.Context(), slug)
	if err != nil {
		if store.IsNotFound(err) {
			http.Error(w, "certificate not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	target := canonical(s.cfg.AppBaseURL, "/verify/"+certificate.Slug)
	png, err := qrcode.Encode(target, qrcode.Medium, 320)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(png)
}

func (s *Server) robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "User-agent: *\nAllow: /\nDisallow: /admin\nDisallow: /verify\nDisallow: /uz/verify\n\nSitemap: %s/sitemap.xml\n", strings.TrimRight(s.cfg.AppBaseURL, "/"))
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
