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

	
	// engine.GET("/list",service.TaskList)
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
    }	
	// ユーザ登録
	engine.GET("/user/new", service.NewUserForm)
	engine.POST("/user/new", service.RegisterUser)

	//ログイン
	engine.GET("/login", service.LoginPage)
	engine.POST("/login", service.Login)	

	engine.GET("/logout", service.Logout)	

	engine.GET("/user/delete", service.LoginCheck, service.DeleteUser)

	engine.GET("/user/edit", service.LoginCheck, service.EditUserForm)

	engine.GET("/user/edit/name", service.LoginCheck, service.EditUserNameForm)
	engine.POST("/user/edit/name", service.LoginCheck, service.UpdateUserName)
	engine.GET("/user/edit/password", service.LoginCheck, service.EditUserPasswordForm)
	engine.POST("/user/edit/password", service.LoginCheck, service.UpdateUserPassword)
	// start server
	engine.Run(fmt.Sprintf(":%d", port))
}
