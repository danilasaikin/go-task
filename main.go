package main

import (
	"database/sql"
	_ "encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	_ "log"
	"os"
	"os/signal"
	"syscall"
)

type News struct {
	Id         int64   `reform:"id,pk"`
	Title      string  `reform:"title"`
	Content    string  `reform:"content"`
	Categories []int64 `reform:"categories"`
}

func SetupDB() (*sql.DB, error) {
	dsn := "root:testtest@tcp(localhost:3306)/News"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Проверка подключения к базе данных
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// EditNews Обработчик для ручки POST /edit/:Id
func EditNews(c *fiber.Ctx) error {
	db := c.Locals("db").(*sql.DB)

	var reqData News
	if err := c.BodyParser(&reqData); err != nil {
		return err
	}

	id := c.Params("id")

	// Выполним SQL-запрос для обновления новости
	_, err := db.Exec("UPDATE News SET Title = ?, Content = ? WHERE Id = ?", reqData.Title, reqData.Content, id)
	if err != nil {
		return err
	}

	// Возвращаем обновленную новость в формате JSON
	return c.JSON(reqData)
}

// Обработчик для ручки GET /list
func GetNewsList(c *fiber.Ctx) error {
	db, err := SetupDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Выполним SQL-запрос для получения списка всех новостей
	rows, err := db.Query("SELECT * FROM News")
	if err != nil {
		return err
	}
	defer rows.Close()

	var newsList []News
	for rows.Next() {
		var news News
		if err := rows.Scan(&news.Id, &news.Title, &news.Content); err != nil {
			return err
		}
		newsList = append(newsList, news)
	}

	// Возвращаем список новостей в формате JSON
	return c.JSON(newsList)
}

// Функция для обновления новости по Id
func UpdateNewsByID(db *sql.DB, id int64, title, content string) error {
	_, err := db.Exec("UPDATE News SET Title=?, Content=? WHERE Id=?", title, content, id)
	return err
}

// GetAllNews Функция для получения списка всех новостей
func GetAllNews(db *sql.DB) ([]News, error) {
	rows, err := db.Query("SELECT * FROM News")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var newsList []News
	for rows.Next() {
		var news News
		if err := rows.Scan(&news.Id, &news.Title, &news.Content); err != nil {
			return nil, err
		}
		newsList = append(newsList, news)
	}

	return newsList, nil
}
func main() {
	// Подключение к базе данных MySQL
	db, err := SetupDB()
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	// Создание экземпляра веб-приложения Fiber
	app := fiber.New()

	// Установка соединения с базой данных в контексте запроса
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", db)
		return c.Next()
	})

	// Маршруты
	app.Post("/edit/:Id", EditNews)
	app.Get("/list", GetNewsList)

	// Запуск сервера
	go func() {
		if err := app.Listen(":3000"); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Ожидание сигнала завершения приложения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// Закрытие соединения с базой данных
	if err := db.Close(); err != nil {
		log.Fatal("Error closing database connection:", err)
	}
}
