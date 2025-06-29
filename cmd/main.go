package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// функция main вызывается автоматически при запуске приложения
func main() {
	// обрабатываем аргументы командной строки
	parseFlags()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// функция run будет полезна при инициализации зависимостей сервера перед запуском
func run() error {
	// Берум аргументы запуска приложения из переменных окружения. Иначе - смотрим в переданных явно аргументах
	// Адрес сервера
	if envServerAddr, hasEnv := os.LookupEnv("RUN_ADDRESS"); hasEnv {
		routerAddr = normalizeAddress(envServerAddr)
	}

	// Строка подключения к базе данных postgres
	if envDBAddr, hasEnv := os.LookupEnv("DATABASE_URI"); hasEnv {
		databaseConnStr = envDBAddr
	}

	// Адрес системы расчёта начислений
	if envAccrualConnStr, hasEnv := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); hasEnv {
		accrualCalculationRouterAddr = envAccrualConnStr
	}

	// Инициализация репозиториев с базой данных
	var repo repository.IRepository[entities.ShURL]
	var err error
	if flagDBConnStr != "" {
		repo, err = postgres.NewPostgresShURLRepository(flagDBConnStr)
	} else {
		repo, err = jsonfile.NewJSONFileShURLRepository(flagDBFilePath)
	}

	if err != nil {
		return err
	}

	defer repo.CloseConnection()

	// Инициализация сервисов
	shURLService := services.NewShURLService(repo)

	// Инициализация обработчиков
	shURLHandler := handlers.NewShURLHandler(shURLService, flagRedirectRouterAddr)

	//Инициализация логгера
	zapLogger, err := logger.NewLogger("Info", true)
	if err != nil {
		return err
	}
	defer zapLogger.Sync()

	// Проверка подключения к БД
	pingFunc := func(w http.ResponseWriter, r *http.Request) {
		if repo.PingDB() {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}

	// Сравниваем нормализованные адреса. Если адрес один - запускаем то и то на одном порту
	if flagShortenerRouterAddr == flagRedirectRouterAddr {
		r := chi.NewRouter()
		r.Use(auth.AuthMiddleware())
		r.Use(logger.LoggingMiddleware(zapLogger))
		r.Use(gzipencoder.GZIPEncodingMiddleware())
		r.Get("/ping", pingFunc)
		r.Get("/api/user/urls", shURLHandler.GetShURLsByUserID)
		r.Delete("/api/user/urls", shURLHandler.DeleteMany)
		r.Get("/{token}", shURLHandler.GetFullURL)
		r.Post("/api/shorten", shURLHandler.ShortenURL)
		r.Post("/api/shorten/batch", shURLHandler.ShortenURLsBatch)
		r.Post("/", shURLHandler.ShortenURL)
		fmt.Println("Running server on", flagShortenerRouterAddr)
		return http.ListenAndServe(flagShortenerRouterAddr, r)
	}

	// Если разные - разные сервера для разных хэндлеров в разных горутинах
	redirectRouter := chi.NewRouter()
	redirectRouter.Use(auth.AuthMiddleware()) //Нужно при обращении к /api/user/urls (GET и DELETE)
	redirectRouter.Use(logger.LoggingMiddleware(zapLogger))
	redirectRouter.Use(gzipencoder.GZIPEncodingMiddleware())
	redirectRouter.Get("/ping", pingFunc) //Дублируется в обоих роутерах
	redirectRouter.Get("/api/user/urls", shURLHandler.GetShURLsByUserID)
	redirectRouter.Delete("/api/user/urls", shURLHandler.DeleteMany)
	redirectRouter.Get("/{token}", shURLHandler.GetFullURL)

	shortenerRouter := chi.NewRouter()
	shortenerRouter.Use(logger.LoggingMiddleware(zapLogger))
	shortenerRouter.Use(gzipencoder.GZIPEncodingMiddleware())
	shortenerRouter.Post("/api/shorten", shURLHandler.ShortenURL)
	shortenerRouter.Post("/api/shorten/batch", shURLHandler.ShortenURLsBatch)
	redirectRouter.Get("/ping", pingFunc) //Дублируется в обоих роутерах
	shortenerRouter.Post("/", shURLHandler.ShortenURL)

	errCh := make(chan error)

	go func() {
		fmt.Println("Running short-to-long redirect server on", flagRedirectRouterAddr)
		errCh <- http.ListenAndServe(flagRedirectRouterAddr, redirectRouter)
	}()

	go func() {
		fmt.Println("Running URL shortener on", flagShortenerRouterAddr)
		errCh <- http.ListenAndServe(flagShortenerRouterAddr, shortenerRouter)
	}()

	// Блокируем основную горутину и обрабатываем ошибки
	return <-errCh
}
