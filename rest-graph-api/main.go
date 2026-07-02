package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// --- Config ---

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Path    PathConfig    `yaml:"path"`
	Compare CompareConfig `yaml:"compare"`
}

type ServerConfig struct {
	Listen string    `yaml:"listen"`
	Port   int       `yaml:"port"`
	TLS    TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type PathConfig struct {
	DataDir       string `yaml:"data_dir"`
	CompareScript string `yaml:"compare_script"`
}

type CompareConfig struct {
	ResultHost string `yaml:"result_host"`
}

func defaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Listen: "",
			Port:   8080,
		},
	}
}

func loadConfig(path string) (Config, error) {
	cfg := defaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("cannot read config file %q: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("cannot parse config file %q: %w", path, err)
	}
	return cfg, nil
}

func (s ServerConfig) listenAddr() string {
	return fmt.Sprintf("%s:%d", s.Listen, s.Port)
}

// --- Types ---

type CompareRequest struct {
	IDs []int64 `json:"ids"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// --- глобальный конфиг, доступный хендлерам ---
var appConfig Config

// --- Handlers ---

func handleGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	rawID := strings.TrimPrefix(r.URL.Path, "/api/v1/graph/")
	rawID = strings.Trim(rawID, "/")

	if rawID == "" {
		writeError(w, http.StatusBadRequest, "missing ID in path")
		return
	}

	id, err := parsePositiveInt64(rawID)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid ID: %v", err))
		return
	}

	result, err := processGraph(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func handleCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req CompareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
		return
	}

	if len(req.IDs) < 2 {
		writeError(w, http.StatusBadRequest, "at least two IDs are required")
		return
	}

	for _, id := range req.IDs {
		if id <= 0 {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("ID must be positive, got: %d", id))
			return
		}
	}

	cfg := appConfig.Path
	id1, id2 := req.IDs[0], req.IDs[1]

	// ── Шаг 1: проверяем существование data_dir ──────────────────────────────
	if err := checkDirExists(cfg.DataDir); err != nil {
		msg := fmt.Sprintf("data_dir does not exist: %s", cfg.DataDir)
		log.Printf("[compare] ERROR: %s", msg)
		writeError(w, http.StatusNotFound, msg)
		return
	}

	// ── Шаг 2: создаём data_dir/compare если нет ─────────────────────────────
	compareBase := filepath.Join(cfg.DataDir, "compare")
	if err := os.MkdirAll(compareBase, 0o755); err != nil {
		msg := fmt.Sprintf("cannot create compare dir %q: %v", compareBase, err)
		log.Printf("[compare] ERROR: %s", msg)
		writeError(w, http.StatusInternalServerError, msg)
		return
	}

	// ── Шаг 3: проверяем существование data_dir/ID1 и data_dir/ID2 ───────────
	dir1 := filepath.Join(cfg.DataDir, strconv.FormatInt(id1, 10))
	dir2 := filepath.Join(cfg.DataDir, strconv.FormatInt(id2, 10))

	missing := []string{}
	for _, p := range []string{dir1, dir2} {
		if err := checkPathExists(p); err != nil {
			missing = append(missing, p)
		}
	}
	if len(missing) > 0 {
		msg := fmt.Sprintf("data directories not found: %s", strings.Join(missing, ", "))
		log.Printf("[compare] ERROR: %s", msg)
		writeError(w, http.StatusNotFound, msg)
		return
	}

	// ── Шаг 4: создаём data_dir/compare/ID1-ID2 если нет ────────────────────
	pairName := fmt.Sprintf("%d-%d", id1, id2)
	resultDir := filepath.Join(compareBase, pairName)
	if err := os.MkdirAll(resultDir, 0o755); err != nil {
		msg := fmt.Sprintf("cannot create result dir %q: %v", resultDir, err)
		log.Printf("[compare] ERROR: %s", msg)
		writeError(w, http.StatusInternalServerError, msg)
		return
	}

	// ── Шаг 5: запускаем скрипт, stdout → index.html ─────────────────────────
	indexPath := filepath.Join(resultDir, "index.html")
	if err := runCompareScript(r.Context(), cfg.CompareScript, dir1, dir2, indexPath); err != nil {
		msg := fmt.Sprintf("compare script failed: %v", err)
		log.Printf("[compare] ERROR: %s", msg)
		writeError(w, http.StatusInternalServerError, msg)
		return
	}

	// ── Шаг 6: формируем ответ с URL ─────────────────────────────────────────
	var resultURL string
	if host := appConfig.Compare.ResultHost; host != "" {
		// Используем явно заданный хост из конфига.
		// host уже содержит scheme, например "http://10.0.0.1"
		// — просто дописываем путь.
		resultURL = fmt.Sprintf("%s/data/compare/%s", strings.TrimRight(host, "/"), pairName)
	} else {
		// Fallback: берём scheme и Host из входящего запроса
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		resultURL = fmt.Sprintf("%s://%s/data/compare/%s", scheme, r.Host, pairName)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ids":    req.IDs,
		"status": "ok",
		"url":    resultURL,
		"result": nil,
	})
}

// --- Business logic ---

func processGraph(id int64) (map[string]any, error) {
	log.Printf("[graph] processing id=%d", id)
	return map[string]any{
		"id":     id,
		"status": "ok",
		"url":    "https://ya.ru",
		"data":   nil,
	}, nil
}

// runCompareScript запускает скрипт с двумя аргументами-путями,
// перенаправляя stdout в outFile, а stderr — в лог сервиса.
func runCompareScript(ctx context.Context, script, arg1, arg2, outFile string) error {
	log.Printf("[compare] running: %s %s %s > %s", script, arg1, arg2, outFile)

	f, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("cannot create output file %q: %w", outFile, err)
	}
	defer f.Close()

	cmd := exec.CommandContext(ctx, script, arg1, arg2)
	cmd.Stdout = f
	cmd.Stderr = log.Writer()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("script exited with error: %w", err)
	}
	return nil
}

// --- Path helpers ---

func checkDirExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%q exists but is not a directory", path)
	}
	return nil
}

func checkPathExists(path string) error {
	_, err := os.Lstat(path)
	return err
}

// --- Helpers ---

func parsePositiveInt64(s string) (int64, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("not a number")
	}
	if n <= 0 {
		return 0, fmt.Errorf("must be positive")
	}
	return n, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

// --- Main ---

const defaultConfigPath = "config/rest-graph-api.yaml"

func main() {
	configPath := flag.String("c", defaultConfigPath, "path to configuration file")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	appConfig = cfg

	log.Printf("loaded config from %q", *configPath)
	log.Printf("  server.listen         : %q", cfg.Server.Listen)
	log.Printf("  server.port           : %d", cfg.Server.Port)
	log.Printf("  path.data_dir         : %q", cfg.Path.DataDir)
	log.Printf("  path.compare_script   : %q", cfg.Path.CompareScript)
	log.Printf("  compare.result_host   : %q", cfg.Compare.ResultHost)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/graph/", handleGraph)
	mux.HandleFunc("/api/v1/compare/", handleCompare)

	addr := cfg.Server.listenAddr()
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

