package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/JustScorpio/loyalty_system/internal/accrual"
	"github.com/JustScorpio/loyalty_system/internal/customerrors"
	"github.com/JustScorpio/loyalty_system/internal/models"
	"github.com/JustScorpio/loyalty_system/internal/repository"
	"github.com/JustScorpio/loyalty_system/internal/utils"
)

type LoyaltyService struct {
	//ВАЖНО: В Go интерфейсы УЖЕ ЯВЛЯЮТСЯ ССЫЛОЧНЫМ ТИПОМ (под капотом — указатель на структуру)
	usersRepo       repository.IRepository[models.User]
	ordersRepo      repository.IRepository[models.Order]
	withdrawalsRepo repository.IRepository[models.Withdrawal]
	accrualClient   *accrual.Client
	txManager       repository.ITransactionManager
	taskQueue       chan Task // канал-очередь задач
}

type TaskType int

const (
	TaskCreateUser TaskType = iota
	TaskGetUser
	TaskCreateOrder
	TaskGetUserOrders
	TaskCreateWithdrawal
	TaskGetUserWithdrawals
)

type Task struct {
	Type     TaskType
	Context  context.Context
	Payload  interface{}
	ResultCh chan TaskResult
}

type TaskResult struct {
	Result interface{}
	Err    error
}

var alreadyExistsError = customerrors.NewAlreadyExistsError(errors.New("entity already exists"))
var notActuallyAnError = customerrors.NewOkError(errors.New("")) //Its a need
var unprocessableEntityError = customerrors.NewUnprocessableEntityError(errors.New("unprocessable entity"))
var paymentRequiredError = customerrors.NewPaymentRequiredError(errors.New("payment required"))

func NewLoyaltyService(usersRepo repository.IRepository[models.User], ordersRepo repository.IRepository[models.Order], withdrawalsRepo repository.IRepository[models.Withdrawal], accrualClient *accrual.Client, txManager repository.ITransactionManager) *LoyaltyService {
	service := &LoyaltyService{
		usersRepo:       usersRepo,
		ordersRepo:      ordersRepo,
		withdrawalsRepo: withdrawalsRepo,
		accrualClient:   accrualClient,
		txManager:       txManager,
		taskQueue:       make(chan Task, 300),
	}

	go service.taskProcessor()

	return service
}

func (s *LoyaltyService) taskProcessor() {
	for task := range s.taskQueue {

		var result interface{}
		var err error

		switch task.Type {
		case TaskCreateUser:
			user := task.Payload.(*models.User)
			err = s.createUser(task.Context, *user)
		case TaskGetUser:
			login := task.Payload.(string)
			result, err = s.usersRepo.Get(task.Context, login)
		case TaskCreateOrder:
			order := task.Payload.(*models.Order)
			err = s.createOrder(task.Context, *order)
		case TaskGetUserOrders:
			login := task.Payload.(string)
			result, err = s.getUserOrders(task.Context, login)
		case TaskCreateWithdrawal:
			withdrawal := task.Payload.(*models.Withdrawal)
			err = s.createWithdrawal(task.Context, *withdrawal)
		case TaskGetUserWithdrawals:
			login := task.Payload.(string)
			result, err = s.getUserWithdrawals(task.Context, login)
		}

		if task.ResultCh != nil {
			switch task.Type {
			case TaskGetUser, TaskGetUserOrders, TaskGetUserWithdrawals:
				task.ResultCh <- TaskResult{
					Result: result,
					Err:    err,
				}
			case TaskCreateUser, TaskCreateOrder, TaskCreateWithdrawal:
				task.ResultCh <- TaskResult{
					Err: err,
				}
			}
			close(task.ResultCh)
		}
	}
}

// Поставить задачу в очередь
func (s *LoyaltyService) enqueueTask(task Task) (interface{}, error) {
	if task.ResultCh == nil {
		task.ResultCh = make(chan TaskResult, 1)
	}

	s.taskQueue <- task

	select {
	case <-task.Context.Done():
		return nil, task.Context.Err()
	case res := <-task.ResultCh:
		return res.Result, res.Err
	}
}

func (s *LoyaltyService) CreateUser(ctx context.Context, newUser models.User) error {
	_, err := s.enqueueTask(Task{
		Type:    TaskCreateUser,
		Context: ctx,
		Payload: &newUser,
	})

	return err
}

func (s *LoyaltyService) GetUser(ctx context.Context, login string) (*models.User, error) {
	res, err := s.enqueueTask(Task{
		Type:    TaskGetUser,
		Context: ctx,
		Payload: login,
	})

	return res.(*models.User), err
}

func (s *LoyaltyService) CreateOrder(ctx context.Context, newOrder models.Order) error {
	_, err := s.enqueueTask(Task{
		Type:    TaskCreateOrder,
		Context: ctx,
		Payload: &newOrder,
	})

	return err
}

func (s *LoyaltyService) GetUserOrders(ctx context.Context, login string) ([]models.Order, error) {
	res, err := s.enqueueTask(Task{
		Type:    TaskGetUserOrders,
		Context: ctx,
		Payload: login,
	})

	return res.([]models.Order), err
}

func (s *LoyaltyService) CreateWithdrawal(ctx context.Context, newWithdrawal models.Withdrawal) error {
	_, err := s.enqueueTask(Task{
		Type:    TaskCreateWithdrawal,
		Context: ctx,
		Payload: &newWithdrawal,
	})

	return err
}

func (s *LoyaltyService) GetUserWithdrawals(ctx context.Context, login string) ([]models.Withdrawal, error) {
	res, err := s.enqueueTask(Task{
		Type:    TaskGetUserWithdrawals,
		Context: ctx,
		Payload: login,
	})

	return res.([]models.Withdrawal), err
}

func (s *LoyaltyService) createUser(ctx context.Context, user models.User) error {

	login := user.Login

	// Проверка наличие логина в БД
	existedUser, err := s.usersRepo.Get(ctx, login)
	if err == nil && existedUser != nil {
		return alreadyExistsError
	}

	err = s.usersRepo.Create(ctx, &user)
	if err != nil {
		return err
	}

	return nil
}

func (s *LoyaltyService) createOrder(ctx context.Context, order models.Order) error {

	number := order.Number

	if !utils.LuhnValidate(number) {
		return unprocessableEntityError
	}

	// Проверка наличие заказа в БД
	existedOrder, err := s.ordersRepo.Get(ctx, number)
	if err == nil && existedOrder != nil {
		if order.UserID == existedOrder.UserID {
			return notActuallyAnError
		} else {
			return alreadyExistsError
		}
	}

	err = s.ordersRepo.Create(ctx, &order)
	if err != nil {
		return err
	}

	return nil
}

func (s *LoyaltyService) createWithdrawal(ctx context.Context, withdrawal models.Withdrawal) error {

	order := withdrawal.Order

	if !utils.LuhnValidate(order) {
		return unprocessableEntityError
	}

	// Проверка наличие заказа в БД
	existedWithdrawal, err := s.withdrawalsRepo.Get(ctx, order)
	if err == nil && existedWithdrawal != nil {
		return unprocessableEntityError
	}

	user, err := s.usersRepo.Get(ctx, withdrawal.UserID)
	if err != nil {
		return err
	}

	if user.CurrentPoints < withdrawal.Sum {
		return paymentRequiredError
	}

	//Добавляем списание и уменьшаем баланс паользователя в одной транзакции
	err = s.txManager.RunInTransaction(ctx, func(ctx context.Context) error {
		if err = s.withdrawalsRepo.Create(ctx, &withdrawal); err != nil {
			return fmt.Errorf("failed to create withdrawal: %w", err)
		}

		//Изменяем баланс пользователя
		user.CurrentPoints -= withdrawal.Sum
		user.WithdrawnPoints += withdrawal.Sum

		if err := s.usersRepo.Update(ctx, user); err != nil {
			return fmt.Errorf("failed to update profile: %w", err)
		}

		return nil
	})

	if err != nil {
		return customerrors.NewInternalServerError(err)
	}

	return nil
}

func (s *LoyaltyService) getUserOrders(ctx context.Context, login string) ([]models.Order, error) {

	orders, err := s.ordersRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var userOrders []models.Order
	for _, order := range orders {
		if order.UserID == login {
			userOrders = append(userOrders, order)
		}
	}

	return userOrders, nil
}

func (s *LoyaltyService) getUserWithdrawals(ctx context.Context, login string) ([]models.Withdrawal, error) {

	withdrawals, err := s.withdrawalsRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var userWithdrawals []models.Withdrawal
	for _, withdrawal := range withdrawals {
		if withdrawal.UserID == login {
			userWithdrawals = append(userWithdrawals, withdrawal)
		}
	}

	return userWithdrawals, nil
}

func (s *LoyaltyService) StartAccrualWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Accrual worker stopped")
			return
		case <-ticker.C:
			s.runAccrualUpdate(ctx)
		}
	}
}

func (s *LoyaltyService) runAccrualUpdate(ctx context.Context) {
	// Получаем всех пользователей, которым нужно обновить начисления
	users, err := s.usersRepo.GetAll(ctx)
	if err != nil {
		log.Printf("Failed to get users for accrual update: %v", err)
		return
	}

	for _, user := range users {
		// Используем отдельный контекст с таймаутом для каждого пользователя
		userCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if err := s.updateAccruals(userCtx, user.Login); err != nil {
			log.Printf("Failed to update accruals for user %s: %v", user.Login, err)
			continue
		}
	}
}

func (s *LoyaltyService) updateAccruals(ctx context.Context, userLogin string) error {
	orders, err := s.getUserOrders(ctx, userLogin)
	if err != nil {
		return err
	}

	user, err := s.usersRepo.Get(ctx, userLogin)
	if err != nil {
		return err
	}

	for _, order := range orders {
		if order.Status == models.StatusNew || order.Status == models.StatusProcessing {
			orderInfo, err := s.accrualClient.GetOrderInfo(ctx, order.Number)
			statusCode := http.StatusOK
			if err != nil {
				//Если не HTTPError - создаём HTTPError
				var httpErr *customerrors.HTTPError
				if !errors.As(err, &httpErr) {
					httpErr = &customerrors.HTTPError{
						Code: http.StatusInternalServerError, //Если изначально не HTTPError - запрос не был отправлен, значит InternalServerError
						Err:  err,
					}
				}

				statusCode = httpErr.Code
			}

			//Если TooManyRequests - преврать синхронизацию
			//orderInfo пуст - перейти к следующему заказу
			if statusCode == http.StatusTooManyRequests {
				return nil
			} else if orderInfo == nil {
				continue
			}

			updatedOrder := models.Order{
				UserID:     user.Login,
				Number:     orderInfo.Order,
				Accrual:    orderInfo.Accrual,
				Status:     models.Status(orderInfo.Status),
				UploadedAt: order.UploadedAt,
			}

			//В транзакции начислить баллы за заказы И обновить баланс пользователя
			err = s.txManager.RunInTransaction(ctx, func(ctx context.Context) error {
				if err = s.ordersRepo.Update(ctx, &updatedOrder); err != nil {
					return fmt.Errorf("failed to update order: %w", err)
				}

				//Проверяем что в теле ответа был Accrual (обратное возможно если status != "PROCESSED")
				if orderInfo.Accrual != 0 {
					//Изменяем баланс пользователя
					user.CurrentPoints += order.Accrual

					if err := s.usersRepo.Update(ctx, user); err != nil {
						return fmt.Errorf("failed to update user balance: %w", err)
					}
				}

				return nil
			})

			//Если внутренняя ошибка - значит что-то серьёзное, прекращаем обработку
			if err != nil {
				return customerrors.NewInternalServerError(err)
			}
		}
	}

	return nil
}
