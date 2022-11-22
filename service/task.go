package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	database "todolist.go/db"
)

// TaskList renders list of tasks in DB
func TaskList(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Get query parameter
	kw := ctx.Query("kw")
	is_done := ctx.Query("is_done")
	isnot_done := ctx.Query("isnot_done")
	SelectStatus := "完了条件なし"
	// Get tasks in DB
	fmt.Println(is_done)
	fmt.Println(isnot_done)
	var tasks []database.Task
    query := "SELECT id, title, created_at, is_done FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ?"
    switch {
    case kw != "":
        err = db.Select(&tasks, query + " AND title LIKE ?", userID, "%" + kw + "%")
	case (is_done == "t" && isnot_done == "f"):
		err = db.Select(&tasks, query, userID)
		SelectStatus = "完了未完了全て"
	case (is_done == "t" && isnot_done == "") || (is_done == "t" && isnot_done == "t"):
		is_done_bool, err := strconv.ParseBool(is_done)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
		err = db.Select(&tasks, query + " AND is_done LIKE ?", userID, is_done_bool)
		SelectStatus = "完了のみ"
	case (is_done == "" && isnot_done == "f") || (is_done == "f" && isnot_done == "f"):
		isnot_done_bool, err := strconv.ParseBool(isnot_done)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
		err = db.Select(&tasks, query + " AND is_done LIKE ?", userID, isnot_done_bool)
		SelectStatus = "未完了のみ"
    default:
		fmt.Println("default case")
    }
	// switch {
	// case kw != "":
	// 	err = db.Select(&tasks, "SELECT * FROM tasks WHERE title LIKE ?", "%"+kw+"%")
	// case (is_done == "t" && isnot_done == "f"):
	// 	err = db.Select(&tasks, "SELECT * FROM tasks")
	// 	SelectStatus = "完了未完了全て"
	// case (is_done == "t" && isnot_done == "") || (is_done == "t" && isnot_done == "t"):
	// 	is_done_bool, err := strconv.ParseBool(is_done)
	// 	if err != nil {
	// 		Error(http.StatusInternalServerError, err.Error())(ctx)
	// 		return
	// 	}
	// 	err = db.Select(&tasks, "SELECT * FROM tasks WHERE is_done LIKE ?", is_done_bool)
	// 	SelectStatus = "完了のみ"
	// case (is_done == "" && isnot_done == "f") || (is_done == "f" && isnot_done == "f"):
	// 	isnot_done_bool, err := strconv.ParseBool(isnot_done)
	// 	if err != nil {
	// 		Error(http.StatusInternalServerError, err.Error())(ctx)
	// 		return
	// 	}
	// 	err = db.Select(&tasks, "SELECT * FROM tasks WHERE is_done LIKE ?", isnot_done_bool)
	// 	SelectStatus = "未完了のみ"
	// default:
	// 	fmt.Println("default case")
	// 	// err = db.Select(&tasks, "SELECT * FROM tasks")
	// }
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Render tasks
	ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks, "Kw": kw, "SelectStatus": SelectStatus})
}

// ShowTask renders a task with given ID
func ShowTask(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// parse ID given as a parameter
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get a task with given ID
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Render task
	ctx.HTML(http.StatusOK, "task.html", task)
}

func NewTaskForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registration"})
}

func RegisterTask(ctx *gin.Context) {
	// Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist {
		Error(http.StatusBadRequest, "No title is given")(ctx)
		return
	}
	// Get task description
	description, exist := ctx.GetPostForm("description")
	if description == "" {
		description = "there is no description"
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Create new data with given title on DB
	result, err := db.Exec("INSERT INTO tasks (title, description) VALUES (?,?)", title, description)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Render status
	path := "/list" // デフォルトではタスク一覧ページへ戻る
	if id, err := result.LastInsertId(); err == nil {
		path = fmt.Sprintf("/task/%d", id) // 正常にIDを取得できた場合は /task/<id> へ戻る
	}
	ctx.Redirect(http.StatusFound, path)
}

func EditTaskForm(ctx *gin.Context) {
	// Get task id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Get target task
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Render edit form
	ctx.HTML(http.StatusOK, "form_edit_task.html",
		gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task})
}

func UpdateTask(ctx *gin.Context) {

	// Get task id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist {
		Error(http.StatusBadRequest, "No title is given")(ctx)
		return
	}
	// Get task is_done
	is_done, exist := ctx.GetPostForm("is_done")
	is_done_bool, err := strconv.ParseBool(is_done)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Get task description
	description, exist := ctx.GetPostForm("description")
	if description == "" {
		description = "there is no description"
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Create new data with given title on DB
	_, err = db.Exec("UPDATE tasks SET title=?,is_done=?,description=? WHERE id=?", title, is_done_bool, description, id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Render status
	path := "/list" // デフォルトではタスク一覧ページへ戻る
	// if id, err := result.LastInsertId(); err == nil {
	path = fmt.Sprintf("/task/%d", id) // 正常にIDを取得できた場合は /task/<id> へ戻る
	// }
	ctx.Redirect(http.StatusFound, path)
}

func DeleteTask(ctx *gin.Context) {
	// Get task id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Delete the task from DB
	_, err = db.Exec("DELETE FROM tasks WHERE id=?", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Redirect to /list
	ctx.Redirect(http.StatusFound, "/list")
}
