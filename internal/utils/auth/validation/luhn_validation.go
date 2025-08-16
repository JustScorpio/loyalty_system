package validation

import (
	"strconv"
	"strings"
	"unicode"
)

// Проверяет, соответствует ли номер алгоритму Луна
func LuhnValidate(number string) bool {
	// Удаляем все пробелы и нецифровые символы
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, number)

	// Проверяем длину (минимум 2 цифры)
	if len(cleaned) < 2 {
		return false
	}

	sum := 0
	// Идем справа налево
	for i := 0; i < len(cleaned); i++ {
		digit, err := strconv.Atoi(string(cleaned[len(cleaned)-1-i]))
		if err != nil {
			return false
		}

		// Каждую вторую цифру умножаем на 2
		if i%2 == 1 {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}
		sum += digit
	}

	// Сумма должна быть кратна 10
	return sum%10 == 0
}
