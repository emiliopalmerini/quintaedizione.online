package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	generatoriData "github.com/emiliopalmerini/quintaedizione-data-ita/data/generatori"
	mappeData "github.com/emiliopalmerini/quintaedizione-data-ita/data/mappe"
	jsondata "github.com/emiliopalmerini/quintaedizione-data-ita/data/srd"
	combatEncounter "github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/application/encounter"
	combatMemory "github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/infrastructure/persistence/memory"
	combatSrdPersistence "github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/infrastructure/persistence/srd"
	combatHandlers "github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/infrastructure/web/handlers"
	combatTemplates "github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/infrastructure/web/templates"
	generatoriApp "github.com/emiliopalmerini/quintaedizione.online/internal/generatori/application"
	generatoriHandlers "github.com/emiliopalmerini/quintaedizione.online/internal/generatori/infrastructure/web/handlers"
	"github.com/emiliopalmerini/quintaedizione.online/internal/infrastructure"
	mappePersistence "github.com/emiliopalmerini/quintaedizione.online/internal/mappe/infrastructure/persistence"
	mappeHandlers "github.com/emiliopalmerini/quintaedizione.online/internal/mappe/web/handlers"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/application/filters"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/application/parsers"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/application/search"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/application/services"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/infrastructure/datastore"
	srdPersistence "github.com/emiliopalmerini/quintaedizione.online/internal/srd/infrastructure/persistence"
	web "github.com/emiliopalmerini/quintaedizione.online/internal/srd/web"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/templates"
	pkgweb "github.com/emiliopalmerini/quintaedizione.online/pkg/web"
	landingTemplates "github.com/emiliopalmerini/quintaedizione.online/web/templates"
)

func main() {
	config := infrastructure.LoadConfig()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if config.IsProduction() {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	// ── SRD setup ──────────────────────────────────────────────

	log.Println("Loading SRD JSON data...")
	crossLinker, err := parsers.NewCrossLinker(jsondata.Files)
	if err != nil {
		log.Printf("Warning: Failed to initialize cross-linker: %v", err)
	}
	renderer := parsers.NewMarkdownRenderer(crossLinker)
	loader := datastore.NewLoader(jsondata.Files, renderer)
	data, sources, err := loader.LoadAll()
	if err != nil {
		log.Fatalf("Failed to load JSON data: %v", err)
	}
	datastore.EnrichDescriptions(data)
	store := datastore.NewStore(data)
	log.Printf("SRD data loaded from %d source(s)", len(sources))
	for _, src := range sources {
		log.Printf("  source: %s (%s)", src.Name, src.ID)
	}

	for _, name := range store.Collections() {
		log.Printf("  %s: %d documents", name, store.Count(name))
	}

	repo := srdPersistence.NewDocumentRepository(store)

	var templateEngine *templates.TemplEngine
	if config.IsProduction() {
		templateEngine = templates.NewTemplEngine()
	} else {
		templateEngine = templates.NewDevTemplEngine()
	}

	filterRegistry := filters.NewInMemoryFilterRegistry()
	filters.RegisterDefaultFilters(filterRegistry)
	filters.RegisterEditionFilter(filterRegistry, sources)

	filterService := services.NewFilterService(filterRegistry)

	// Find the default source short name for legacy URL redirects and deduplication
	defaultSourceShort := ""
	for _, src := range sources {
		if src.Default {
			defaultSourceShort = src.ShortName
			break
		}
	}
	if defaultSourceShort == "" && len(sources) > 0 {
		defaultSourceShort = sources[0].ShortName
	}

	searchRepo := srdPersistence.NewSearchRepository(store, defaultSourceShort)
	searchService := search.NewFuzzySearchService(searchRepo)
	contentService := services.NewContentService(repo, filterService, services.WithDefaultSource(defaultSourceShort))

	// ── Metrics setup ─────────────────────────────────────────
	metricsRegistry := prometheus.NewRegistry()
	metricsRegistry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	metricsRegistry.MustRegister(collectors.NewGoCollector())
	appMetrics := pkgweb.NewMetrics(metricsRegistry)

	multiSource := len(sources) > 1
	srdHandlers := web.NewHandlers(contentService, searchService, templateEngine, defaultSourceShort, multiSource, store.AvailableSources, web.WithSearchRecorder(appMetrics))

	// ── Combattimenti setup ────────────────────────────────────

	log.Println("Loading Combattimenti data...")
	encounterRepo := combatMemory.NewEncounterRepository()

	encounterService := combatEncounter.NewService(logger, encounterRepo)
	queryHandler := combatEncounter.NewQueryHandler(logger, encounterRepo)
	monsterReader := combatSrdPersistence.NewMonsterReader(repo)
	cartPricer := combatEncounter.NewCartPricer(logger, monsterReader, encounterRepo)

	encounterHandler := combatHandlers.NewEncounterHandler(encounterService, queryHandler, cartPricer, monsterReader, logger)
	pickerHandler := combatHandlers.NewMonsterPickerHandler(monsterReader, logger)
	log.Println("Combattimenti ready")

	// ── Mappe setup ───────────────────────────────────────────

	log.Println("Loading Mappe data...")
	mappaRepo := mappePersistence.NewMappaRepository(mappeData.Files, "mappe.json")
	galleryHandler := mappeHandlers.NewGalleryHandler(mappaRepo, logger)
	log.Printf("Mappe ready: %d maps loaded", len(mappaRepo.FindAll()))

	// ── Generatori setup ─────────────────────────────────────

	log.Println("Loading Generatori data...")
	generatoriService, err := generatoriApp.NewService(generatoriData.Files)
	if err != nil {
		log.Fatalf("Failed to load generatori data: %v", err)
	}
	generatorHandler := generatoriHandlers.NewGeneratorHandler(generatoriService, logger)
	log.Printf("Generatori ready: %d tables loaded", len(generatoriService.List()))

	// ── Router setup ───────────────────────────────────────────

	mux := http.NewServeMux()

	// Landing page
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := landingTemplates.LandingPage().Render(r.Context(), w); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	// Static files (shared) — cache for 1 day (assets are unversioned)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.FileServer(http.Dir("./web/static")).ServeHTTP(w, r)
	})))

	// Health check
	healthChecker := &pkgweb.HealthChecker{
		Checks: []pkgweb.HealthCheck{
			{
				Name: "store",
				Check: func() bool {
					return len(store.Collections()) > 0
				},
			},
		},
	}
	mux.HandleFunc("GET /health", healthChecker.Handler())

	// SEO
	mux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=86400, public")
		http.ServeFile(w, r, "./web/static/robots.txt")
	})
	mux.HandleFunc("GET /sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=86400, public")
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		// TODO: generate comprehensive sitemap
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>https://quintaedizione.online/</loc><priority>1.0</priority></url><url><loc>https://quintaedizione.online/srd</loc><priority>0.9</priority></url><url><loc>https://quintaedizione.online/combattimenti</loc><priority>0.9</priority></url></urlset>`))
	})

	// Legacy redirects: old /:collection URLs → /srd/:collection
	mux.HandleFunc("GET /search", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/srd/search?"+r.URL.RawQuery, http.StatusMovedPermanently)
	})
	mux.HandleFunc("GET /search/dropdown", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/srd/search/dropdown?"+r.URL.RawQuery, http.StatusMovedPermanently)
	})

	// SRD routes under /srd
	srdMux := http.NewServeMux()
	srdHandlers.RegisterRoutes(srdMux)
	mux.Handle("/srd/", http.StripPrefix("/srd", srdMux))

	// Build edition options for the encounter calculator
	editions := make([]combatTemplates.EditionOption, 0, len(sources))
	for _, src := range sources {
		editions = append(editions, combatTemplates.EditionOption{
			SourceID:  src.ID,
			Name:      src.Name,
			ShortName: src.ShortName,
			Ruleset:   src.Ruleset,
			IsDefault: src.Default,
		})
	}

	// Combattimenti routes under /combattimenti
	combatMux := http.NewServeMux()
	combatMux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		pkgweb.SetCacheHeaders(w, 3600) // 1 hour
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := combatTemplates.Home(editions).Render(r.Context(), w); err != nil {
			logger.Error("Failed to render combattimenti home", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	combatMux.HandleFunc("POST /calculate", encounterHandler.CalculateHandler)
	combatMux.HandleFunc("GET /party-input", encounterHandler.PartyInputHandler)
	combatMux.HandleFunc("GET /api/difficulties", encounterHandler.GetDifficultiesHandler)
	combatMux.HandleFunc("GET /monsters", pickerHandler.Handler)
	mux.Handle("/combattimenti/", http.StripPrefix("/combattimenti", combatMux))

	// Mappe routes under /mappe
	mappeMux := http.NewServeMux()
	mappeMux.HandleFunc("GET /{$}", galleryHandler.HandleGallery)
	mappeMux.HandleFunc("GET /gallery", galleryHandler.HandleGallery)
	mappeMux.HandleFunc("GET /{slug}", galleryHandler.HandleDetail)
	mux.Handle("/mappe/", http.StripPrefix("/mappe", mappeMux))

	// Generatori routes under /generatori
	generatoriMux := http.NewServeMux()
	generatorHandler.RegisterRoutes(generatoriMux)
	mux.Handle("/generatori/", http.StripPrefix("/generatori", generatoriMux))

	// ── Middleware chain ────────────────────────────────────────

	rateLimiter := pkgweb.NewRateLimiter()

	var handler http.Handler = mux
	handler = pkgweb.ThemeMiddleware(handler)
	handler = pkgweb.CORSMiddleware(handler)
	handler = pkgweb.RateLimitMiddleware(rateLimiter)(handler)
	handler = appMetrics.Middleware(handler)
	handler = pkgweb.SecurityMiddleware(handler)
	handler = pkgweb.ErrorRecoveryMiddleware(logger)(handler)

	// ── Server ─────────────────────────────────────────────────

	srv := &http.Server{
		Addr:              config.GetAddress(),
		Handler:           handler,
		ReadHeaderTimeout: 15 * time.Second,
	}

	// ── Internal metrics server ──────────────────────────────
	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9090"
	}
	metricsMux := http.NewServeMux()
	metricsMux.Handle("GET /metrics", promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{}))

	metricsSrv := &http.Server{
		Addr:              ":" + metricsPort,
		Handler:           metricsMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("Starting metrics server on :%s", metricsPort)
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()

	go func() {
		log.Printf("Starting quintaedizione.online on %s", config.GetAddress())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Metrics server forced to shutdown: %v", err)
	}
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server shutdown completed")
}
