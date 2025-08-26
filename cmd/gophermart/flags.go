package main

import (
	"flag"
)

// Адрес и порт для запуска сервера
var routerAddr string

// Строка подключения к БД postgresql
var databaseConnStr string

// Адрес системы расчёта начислений
var accrualCalculationRouterAddr string

// parseFlags обрабатывает аргументы командной строки и сохраняет их значения в соответствующих переменных
func parseFlags() {
	flag.StringVar(&routerAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&databaseConnStr, "d", "Host=127.0.0.1;Port=5432;Database=exampledb;Username=postgres;Password=password;", "postgresql database connection string")
	flag.StringVar(&accrualCalculationRouterAddr, "r", ":8080", "address of the accrual calculation system")
	flag.Parse()
}
