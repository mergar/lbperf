package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"send-data/internal/auth"
	"send-data/internal/clientip"
	"send-data/internal/config"
	"send-data/internal/postprocess"
	"send-data/internal/ratelimit"
)

type Handler struct {
	cfg   *config.Config
	limit *ratelimit.Limiter
	post  *postprocess.Runner
}

func New(cfg *config.Config) *Handler {
	return &Handler{
		cfg:   cfg,
		limit: ratelimit.New(cfg.Storage.SpoolDir),
		post:  postprocess.New(cfg.PostProcess.Enabled, cfg.PostProcess.Script),
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/scripts/getid.php", h.GetID)
	mux.HandleFunc("/scripts/enable_token.php", h.EnableToken)
	mux.HandleFunc("/scripts/disable_token.php", h.DisableToken)
	mux.HandleFunc("/scripts/report_system.php", h.ReportSystem)
}

func (h *Handler) GetID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fail(w, http.StatusMethodNotAllowed, "")
		return
	}

	q := r.URL.Query()
	if msg := checkQuerySize(q, 1, 64); msg != "" {
		fail(w, http.StatusServiceUnavailable, msg)
		return
	}

	key, ok := auth.ParseKey(q.Get("key"))
	if !ok {
		if q.Get("key") == "" {
			fail(w, http.StatusServiceUnavailable, "unable to get key")
		} else {
			fail(w, http.StatusServiceUnavailable, "wrong key")
		}
		return
	}

	ip := clientip.FromRequest(r, h.cfg.Proxy.TrustedAddrs)
	if err := h.limit.Allow("getid", ip, h.cfg.RateLimit.GetIDMaxPerIP, h.cfg.RateLimit.GetIDIntervalSec); err != nil {
		if errors.Is(err, ratelimit.ErrRateLimited) {
			fail(w, http.StatusServiceUnavailable, "Rate limit per IP, please try later\n")
			return
		}
		fail(w, http.StatusServiceUnavailable, "")
		return
	}

	token := auth.ComputeToken(key, h.cfg.Auth.TokenSecret)
	okBody(w, "KEY="+key+"\r\nTOKEN="+token+"\r\n")
}

func (h *Handler) EnableToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fail(w, http.StatusMethodNotAllowed, "")
		return
	}

	key, token, ok := h.parseAuthQuery(w, r)
	if !ok {
		return
	}
	if !auth.ValidateToken(key, token, h.cfg.Auth.TokenSecret) {
		fail(w, http.StatusServiceUnavailable, "wrong key/token")
		return
	}

	dataDir := h.cfg.DataDir(token)
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		fail(w, http.StatusServiceUnavailable, "")
		return
	}

	okBody(w, "STATUS=OK\r\n")
}

func (h *Handler) DisableToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fail(w, http.StatusMethodNotAllowed, "")
		return
	}

	key, token, ok := h.parseAuthQuery(w, r)
	if !ok {
		return
	}
	if !auth.ValidateToken(key, token, h.cfg.Auth.TokenSecret) {
		fail(w, http.StatusServiceUnavailable, "wrong key/token")
		return
	}

	dataDir := h.cfg.DataDir(token)
	if err := os.Remove(dataDir); err != nil && !os.IsNotExist(err) {
		fail(w, http.StatusServiceUnavailable, "")
		return
	}

	okBody(w, "STATUS=OK\r\n")
}

func (h *Handler) ReportSystem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, http.StatusMethodNotAllowed, "")
		return
	}

	key, token, ok := h.parseAuthQuery(w, r)
	if !ok {
		return
	}
	if !auth.ValidateToken(key, token, h.cfg.Auth.TokenSecret) {
		fail(w, http.StatusServiceUnavailable, "wrong key/token")
		return
	}

	ip := clientip.FromRequest(r, h.cfg.Proxy.TrustedAddrs)
	if err := h.limit.Allow("report", ip, h.cfg.RateLimit.ReportMaxPerIP, h.cfg.RateLimit.ReportIntervalSec); err != nil {
		if errors.Is(err, ratelimit.ErrRateLimited) {
			fail(w, http.StatusServiceUnavailable, "Rate limit per IP, please try later\n")
			return
		}
		fail(w, http.StatusServiceUnavailable, "")
		return
	}

	dataDir := h.cfg.DataDir(token)
	if _, err := os.Stat(dataDir); err != nil {
		if os.IsNotExist(err) {
			fail(w, http.StatusServiceUnavailable, "not enabled")
			return
		}
		fail(w, http.StatusServiceUnavailable, "")
		return
	}

	if cl := r.ContentLength; cl >= 0 && cl > int64(h.cfg.Limits.MaxPostBodySize) {
		fail(w, http.StatusServiceUnavailable, "wrong body size")
		return
	}

	limited := io.LimitReader(r.Body, int64(h.cfg.Limits.MaxPostBodySize)+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		fail(w, http.StatusServiceUnavailable, "")
		return
	}
	if len(body) <= h.cfg.Limits.MinPostBodySize || len(body) > h.cfg.Limits.MaxPostBodySize {
		fail(w, http.StatusServiceUnavailable, "wrong body size")
		return
	}

	ts := time.Now().Unix()
	identify, err := json.Marshal(queryToMap(r.URL.Query()))
	if err != nil {
		fail(w, http.StatusServiceUnavailable, "")
		return
	}

	identifyPath := dataDir + "/identify.json." + strconv.FormatInt(ts, 10)
	if err := os.WriteFile(identifyPath, identify, 0600); err != nil {
		fail(w, http.StatusServiceUnavailable, "")
		return
	}
	if err := os.WriteFile(dataDir+"/data.tgz", body, 0600); err != nil {
		fail(w, http.StatusServiceUnavailable, "")
		return
	}

	h.post.Run(dataDir)

//	origin := originSite(r, h.cfg)
	origin := "http://172.16.0.133"
	okBody(w, "Graph URL: "+origin+"/data/"+strconv.FormatInt(ts, 10)+"\r\n")
//	okBody(w, "Graph URL: "+origin+"/api/v1/graph/"+strconv.FormatInt(ts, 10)+"\r\n")
}

func (h *Handler) parseAuthQuery(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	q := r.URL.Query()
	if msg := checkQuerySize(q, 2, h.cfg.Limits.MaxAuthGetSize); msg != "" {
		fail(w, http.StatusServiceUnavailable, msg)
		return "", "", false
	}

	key, ok := auth.ParseKey(q.Get("key"))
	if !ok {
		if q.Get("key") == "" {
			fail(w, http.StatusServiceUnavailable, "unable to get key")
		} else {
			fail(w, http.StatusServiceUnavailable, "wrong key")
		}
		return "", "", false
	}

	token, ok := auth.ParseToken(q.Get("token"))
	if !ok {
		if q.Get("token") == "" {
			fail(w, http.StatusServiceUnavailable, "unable to get token")
		} else {
			fail(w, http.StatusServiceUnavailable, "wrong token")
		}
		return "", "", false
	}

	return key, token, true
}

func checkQuerySize(q url.Values, count, maxSize int) string {
	if len(q) != count {
		return "wrong args"
	}
	sz := queryValueSize(q)
	if maxSize > 0 && sz > maxSize {
		return fmt.Sprintf("wrong size: %d / %d", maxSize, sz)
	}
	return ""
}

func queryValueSize(q url.Values) int {
	size := 0
	for _, vals := range q {
		for _, v := range vals {
			size += len(v)
		}
	}
	return size
}

func queryToMap(q url.Values) map[string]string {
	out := make(map[string]string, len(q))
	for k, vals := range q {
		if len(vals) > 0 {
			out[k] = vals[0]
		}
	}
	return out
}

func originSite(r *http.Request, cfg *config.Config) string {
	if cfg.Origin.Site != "" {
		return cfg.Origin.Site
	}

	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	return proto + "://" + r.Host
}

func okBody(w http.ResponseWriter, body string) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

func fail(w http.ResponseWriter, code int, body string) {
	w.WriteHeader(code)
	if body != "" {
		_, _ = w.Write([]byte(body))
	}
}
