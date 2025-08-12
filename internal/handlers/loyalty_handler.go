package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/JustScorpio/loyalty_system/internal/customcontext"
	"github.com/JustScorpio/loyalty_system/internal/customerrors"
	"github.com/JustScorpio/loyalty_system/internal/middleware/auth" //В файле middleware не только сама middleware, но и ауфные функции и константы
	"github.com/JustScorpio/loyalty_system/internal/models"
	"github.com/JustScorpio/loyalty_system/internal/services"
)

type LoyaltyHandler struct {
	service *services.LoyaltyService
}

func NewLoyaltyHandler(service *services.LoyaltyService) *LoyaltyHandler {
	return &LoyaltyHandler{
		service: service,
	}
}

// Регистрация пользователя
func (h *LoyaltyHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// разрешаем только POST-запросы
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	//Если Body пуст
	if len(body) == 0 {
		http.Error(w, "Body is empty", http.StatusBadRequest)
		return
	}

	//Только Content-Type: JSON
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var reqData struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err = json.Unmarshal(body, &reqData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user = *models.NewUser(reqData.Login, reqData.Password)

	//Создаём пользователя
	err = h.service.CreateUser(r.Context(), user)

	if err != nil {
		statusCode := http.StatusInternalServerError
		var httpErr *customerrors.HTTPError
		if errors.As(err, &httpErr) {
			statusCode = httpErr.Code
		}

		w.WriteHeader(statusCode)
		return
	}

	//Авторизуем пользователя и устанавливаем куки
	token, err := auth.GenerateToken(user.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.JwtCookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(auth.TokenLifeTime),
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusCreated)
}

// Аутентификация пользователя
func (h *LoyaltyHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// разрешаем только POST-запросы
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	//Если Body пуст
	if len(body) == 0 {
		http.Error(w, "Body is empty", http.StatusBadRequest)
		return
	}

	//Только Content-Type: JSON
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var reqData struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err = json.Unmarshal(body, &reqData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//Проверяем пользователя
	user, err := h.service.GetUser(r.Context(), reqData.Login)
	if err != nil || user == nil || user.Password != reqData.Password { // В реальном приложении использовать bcrypt!
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	//Авторизуем пользователя и устанавливаем куки
	token, err := auth.GenerateToken(user.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.JwtCookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(auth.TokenLifeTime),
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}

// Получить баланс пользователя
func (h *LoyaltyHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		// разрешаем только Get-запросы
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := customcontext.GetUserID(r.Context())
	if userID == "" {
		// UserID в куке пуст
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Получение сущностей из сервиса
	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		//Самая странная ситуация когда пользователь авторизован, но в базе его уже нет
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var respData struct {
		Current   float32 `json:"current"`
		Withdrawn float32 `json:"withdrawn"`
	}

	respData.Current = user.CurrentPoints
	respData.Withdrawn = user.WithdrawnPoints

	jsonData, err := json.Marshal(respData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// Загрузить номер заказа
func (h *LoyaltyHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// разрешаем только POST-запросы
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	//Если Body пуст
	if len(body) == 0 {
		http.Error(w, "Body is empty", http.StatusBadRequest)
		return
	}

	//Извлекаем номер заказа
	var orderNum = string(body)

	userID := customcontext.GetUserID(r.Context())
	if userID == "" {
		// UserID в куке пуст
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	order := *models.NewOrder(userID, orderNum)

	//Создаём order
	err = h.service.CreateOrder(r.Context(), order)

	//Определяем статус код
	statusCode := http.StatusAccepted
	if err != nil {
		var httpErr *customerrors.HTTPError
		if errors.As(err, &httpErr) {
			statusCode = httpErr.Code
		}
	}

	w.WriteHeader(statusCode)
}

// Получить все заказы пользователя
func (h *LoyaltyHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		// разрешаем только Get-запросы
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := customcontext.GetUserID(r.Context())
	if userID == "" {
		// UserID в куке пуст
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Получение сущностей из сервиса
	orders, err := h.service.GetUserOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type respItem struct {
		Number     string        `json:"number"`
		Status     models.Status `json:"status"`
		Accrual    *float32      `json:"accrual,omitempty"`
		UploadedAt time.Time     `json:"uploaded_at"`
	}

	var respData []respItem

	for _, order := range orders {
		item := respItem{
			Number:     order.Number,
			Status:     order.Status,
			UploadedAt: order.UploadedAt,
		}
		if order.Status == models.StatusProcessed {
			item.Accrual = &order.Accrual // Указываем Accrual только для Processed
		}

		respData = append(respData, item)
	}

	jsonData, err := json.Marshal(respData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// Загрузить номер заказа
func (h *LoyaltyHandler) UploadWithdrawal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// разрешаем только POST-запросы
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	//Если Body пуст
	if len(body) == 0 {
		http.Error(w, "Body is empty", http.StatusBadRequest)
		return
	}

	//Только Content-Type: JSON
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var reqData struct {
		Order string  `json:"order"`
		Sum   float32 `json:"sum"`
	}

	if err = json.Unmarshal(body, &reqData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := customcontext.GetUserID(r.Context())
	if userID == "" {
		// UserID в куке пуст
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	withdrawal := *models.NewWithdrawal(userID, reqData.Order, reqData.Sum)

	//Создаём withdrawal
	err = h.service.CreateWithdrawal(r.Context(), withdrawal)

	//Определяем статус код
	statusCode := http.StatusOK
	if err != nil {
		var httpErr *customerrors.HTTPError
		if errors.As(err, &httpErr) {
			statusCode = httpErr.Code
		}
	}

	w.WriteHeader(statusCode)
}

// Получить все списания пользователя
func (h *LoyaltyHandler) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		// разрешаем только Get-запросы
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := customcontext.GetUserID(r.Context())
	if userID == "" {
		// UserID в куке пуст
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Получение сущностей из сервиса
	withdrawals, err := h.service.GetUserWithdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type respItem struct {
		Order      string    `json:"order"`
		Sum        float32   `json:"sum"`
		PocessedAt time.Time `json:"processed_at"`
	}

	var respData []respItem

	for _, withdrawal := range withdrawals {
		item := respItem{
			Order:      withdrawal.Order,
			Sum:        withdrawal.Sum,
			PocessedAt: withdrawal.ProcessedAt,
		}

		respData = append(respData, item)
	}

	jsonData, err := json.Marshal(respData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
