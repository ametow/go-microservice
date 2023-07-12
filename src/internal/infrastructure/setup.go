package infrastructure

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"yandex-team.ru/bstask/internal/handlers/courier"
	"yandex-team.ru/bstask/internal/handlers/misc"
	"yandex-team.ru/bstask/internal/handlers/order"
	courierRepo "yandex-team.ru/bstask/internal/pkg/repository/courier"
	orderRepo "yandex-team.ru/bstask/internal/pkg/repository/order"
	courierService "yandex-team.ru/bstask/internal/usecase/courier"
	orderService "yandex-team.ru/bstask/internal/usecase/order"
)

func Setup() *echo.Echo {
	db, err := ConnectDb(Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
		Password: viper.GetString("db.password"),
	})
	if err != nil {
		logrus.Fatalf("failed to initialize db: %s", err.Error())
	}

	app := echo.New()

	// Задание 3 (rate limited to 10 rps)
	app.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(10)))

	courierRepo := courierRepo.NewRepo(db)
	cService := courierService.NewCourierService(courierRepo)
	courierHandler := courier.NewHandler(cService)
	courierHandler.Init(app)

	orderRepo := orderRepo.NewRepo(db)
	oService := orderService.NewOrderService(&orderRepo)
	orderHandler := order.NewHandler(oService)
	orderHandler.Init(app)

	misc.NewHandler(app)

	return app
}
