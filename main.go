package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"

	"todolist.go/db"
	"todolist.go/service"
)

const port = 8000

func main() {
	// initialize DB connection
	dsn := db.DefaultDSN(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	if err := db.Connect(dsn); err != nil {
		log.Fatal(err)
	}

	// initialize Gin engine
	engine := gin.Default()
	engine.LoadHTMLGlob("views/*.html")

    // prepare session
    store := cookie.NewStore([]byte("my-secret"))
    engine.Use(sessions.Sessions("user-session", store))

	// routing
	engine.Static("/assets", "./assets")

	
	engine.GET("/", service.Home)
	engine.GET("/list", service.LoginCheck, service.TaskList)
	taskGroup := engine.Group("/task")
    taskGroup.Use(service.LoginCheck)
    {
        taskGroup.GET("/:id", service.ShowTask)// ":id" is a parameter
        taskGroup.GET("/new", service.NewTaskForm)
        taskGroup.POST("/new", service.RegisterTask)
        taskGroup.GET("/edit/:id", service.EditTaskForm)
        taskGroup.POST("/edit/:id", service.UpdateTask)
        taskGroup.GET("/delete/:id", service.DeleteTask)
		// タスクの共有
		taskGroup.GET("/share/:id", service.ShareTaskForm)
		taskGroup.POST("/share/:id", service.UpdateShareTask)
		taskGroup.GET("/share/delete/:id", service.DeleteShareTask)
    }	
	engine.GET("/user/new", service.NewUserForm)
	engine.POST("/user/new", service.RegisterUser)
	engine.GET("/login", service.LoginPage)
	engine.POST("/login", service.Login)	
	engine.GET("/logout", service.Logout)	
	userGroup := engine.Group("/user")
    userGroup.Use(service.LoginCheck)
    {
		userGroup.GET("/delete", service.DeleteUser)
		// ユーザー情報変更系
		userGroup.GET("/edit", service.EditUserForm)
		userGroup.GET("/edit/name", service.EditUserNameForm)
		userGroup.POST("/edit/name", service.UpdateUserName)
		userGroup.GET("/edit/password", service.EditUserPasswordForm)
		userGroup.POST("/edit/password", service.UpdateUserPassword)
    }	
	// start server
	engine.Run(fmt.Sprintf(":%d", port))
}
