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
	flag.StringVar(&databaseConnStr, "b", "Host=127.0.0.1;Port=5432;Database=exampledb;Username=postgres;Password=password;", "postgresql database connection string")
	flag.StringVar(&accrualCalculationRouterAddr, "r", "127.0.0.1:8080", "address of the accrual calculation system")
	flag.Parse()

	// routerAddr = normalizeAddress(routerAddr)
	// accrualCalculationRouterAddr = normalizeAddress(accrualCalculationRouterAddr)
}

// // Нормализация адресов
// func normalizeAddress(addr string) string {

// 	// Добавляем порт, если его нет
// 	if !strings.Contains(addr, ":") {
// 		addr += ":8080"
// 	}

// 	// Убираем часть http://
// 	if strings.HasPrefix(addr, "http://") {
// 		addr = strings.Replace(addr, "http://", "", 1)
// 	}

// 	// Убираем 127.0.0.1 и localhost
// 	if strings.HasPrefix(addr, "127.0.0.1:") {
// 		addr = strings.Replace(addr, "127.0.0.1", "", 1)
// 	}
// 	if strings.HasPrefix(addr, "localhost:") {
// 		addr = strings.Replace(addr, "localhost", "", 1)
// 	}

// 	return addr
// }
