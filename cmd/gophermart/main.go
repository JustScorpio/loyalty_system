package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/JustScorpio/loyalty_system/internal/accrual"
	"github.com/JustScorpio/loyalty_system/internal/handlers"
	"github.com/JustScorpio/loyalty_system/internal/middleware/auth"
	"github.com/JustScorpio/loyalty_system/internal/middleware/gzipencoder"
	"github.com/JustScorpio/loyalty_system/internal/middleware/logger"
	"github.com/JustScorpio/loyalty_system/internal/repository/postgres"
	"github.com/JustScorpio/loyalty_system/internal/services"
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
		routerAddr = envServerAddr
	}

	// Строка подключения к базе данных postgres
	if envDBAddr, hasEnv := os.LookupEnv("DATABASE_URI"); hasEnv {
		databaseConnStr = envDBAddr
	}

	fmt.Println("Connection string: ", databaseConnStr)

	// Адрес системы расчёта начислений
	if envAccrualConnStr, hasEnv := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); hasEnv {
		accrualCalculationRouterAddr = envAccrualConnStr
	}

	db, err := postgres.NewDBConnection(databaseConnStr)
	if err != nil {
		return err
	}

	// Инициализация репозиториев с базой данных
	usersRepo, err := postgres.NewPgUsersRepo(db)
	if err != nil {
		return err
	}
	defer usersRepo.CloseConnection()
	ordersRepo, err := postgres.NewPgOrdersRepo(db)
	if err != nil {
		return err
	}
	defer ordersRepo.CloseConnection()
	withdrawalsRepo, err := postgres.NewPgWithdrawalsRepo(db)
	if err != nil {
		return err
	}
	defer withdrawalsRepo.CloseConnection()

	//Инициализация клиента для работы с системой рассчёта баллов
	accrualSystemClient := accrual.NewClient(accrualCalculationRouterAddr, 5*time.Second) //Таймаут 5 секунд

	//Инициализация менеджера транзакций
	txManager := postgres.NewPgxTransactionManager(db)

	// Инициализация сервисов
	loyaltyService := services.NewLoyaltyService(usersRepo, ordersRepo, withdrawalsRepo, accrualSystemClient, txManager)

	// Инициализация обработчиков
	loyaltyHandler := handlers.NewLoyaltyHandler(loyaltyService)

	//Инициализация логгера
	zapLogger, err := logger.NewLogger("Info", true)
	if err != nil {
		return err
	}
	defer zapLogger.Sync()

	r := chi.NewRouter()

	//Базовые middleware
	r.Use(logger.LoggingMiddleware(zapLogger))
	r.Use(gzipencoder.GZIPEncodingMiddleware())

	//Публичные маршруты
	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", loyaltyHandler.Register)
		r.Post("/api/user/login", loyaltyHandler.Login)
	})

	//Защищённые маршруты с auth middleware
	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware())
		r.Post("/api/user/orders", loyaltyHandler.UploadOrder)
		r.Get("/api/user/orders", loyaltyHandler.GetUserOrders)
		r.Get("/api/user/balance", loyaltyHandler.GetBalance)
		r.Post("/api/user/balance/withdraw", loyaltyHandler.UploadWithdrawal)
		r.Get("/api/user/withdrawals", loyaltyHandler.GetUserWithdrawals)
	})

	fmt.Println("Running server on", routerAddr)
	return http.ListenAndServe(routerAddr, r)
}
