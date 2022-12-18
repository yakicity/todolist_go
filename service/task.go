package service

import (
	"fmt"
	"net/http"
	"strconv"
	"database/sql"
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
	is_done_status := ctx.Query("is_done_status")
	// Get tasks in DB
	fmt.Println(is_done_status)
	
	var tasks []database.Task
    query := "SELECT id, title, created_at, is_done, description, priority, deadline FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ?"
    switch {
    case kw != "":
        err = db.Select(&tasks, query + " AND title LIKE ?", userID, "%" + kw + "%")
	case (is_done_status == "all"):
		err = db.Select(&tasks, query, userID)
	case (is_done_status == "t"):
		is_done_bool, err := strconv.ParseBool(is_done_status)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
		err = db.Select(&tasks, query + " AND is_done LIKE ?", userID, is_done_bool)
	case (is_done_status == "f"):
		is_done_bool, err := strconv.ParseBool(is_done_status)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
		err = db.Select(&tasks, query + " AND is_done LIKE ?", userID, is_done_bool)
    default:
		err = db.Select(&tasks, query, userID)
		fmt.Println("default case")
    }
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	var user database.User
	err = db.Get(&user, "SELECT id, name FROM users WHERE id =?", userID) // Use DB#Get for one entry
	if err != nil {
		// Error(http.StatusBadRequest, err.Error())(ctx)
		ctx.Redirect(http.StatusFound, "/")
		return
	}
	// Render tasks
	ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list","SessionUser": user.Name, "Tasks": tasks, "Kw": kw,"Value":is_done_status})
}

// ShowTask renders a task with given ID
func ShowTask(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	
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
	err = db.Get(&task, "SELECT id, title, created_at, is_done, description, priority, deadline FROM tasks INNER JOIN ownership ON task_id = id WHERE task_id =? AND ownership.user_id = ?", id,userID) // Use DB#Get for one entry
	if err != nil {
		// Error(http.StatusBadRequest, err.Error())(ctx)
		ctx.Redirect(http.StatusFound, "/list")
		return
	}

	// Render task
	ctx.HTML(http.StatusOK, "task.html", task)
}

func NewTaskForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registration"})
}

func RegisterTask(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
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
	deadlinedate, exist := ctx.GetPostForm("deadlinedate")
	deadlinetime, exist := ctx.GetPostForm("deadlinetime")
	priority, exist := ctx.GetPostForm("priority")
	deadline := deadlinedate + " " + deadlinetime + ":00"

	// Get DB connection
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    tx := db.MustBegin()
	var result sql.Result
	result, err = tx.Exec("INSERT INTO tasks (title, description, priority,deadline) VALUES (?,?,?,?)", title, description,priority,deadline)


	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	
    taskID, err := result.LastInsertId()
    if err != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
	
    _, err = tx.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", userID, taskID)
    if err != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    tx.Commit()
    ctx.Redirect(http.StatusFound, fmt.Sprintf("/task/%d", taskID))
}

func EditTaskForm(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
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
	// err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
	// if err != nil {
	// 	Error(http.StatusBadRequest, err.Error())(ctx)
	// 	return
	// }

	err = db.Get(&task, "SELECT id, title, created_at, is_done, description, priority, deadline FROM tasks INNER JOIN ownership ON task_id = id WHERE task_id =? AND ownership.user_id = ?", id,userID)
	if err != nil {
		// Error(http.StatusBadRequest, err.Error())(ctx)
		ctx.Redirect(http.StatusFound, "/list")
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

	deadlinedate, exist := ctx.GetPostForm("deadlinedate")
	deadlinetime, exist := ctx.GetPostForm("deadlinetime")
	priority, exist := ctx.GetPostForm("priority")

	deadline := deadlinedate + " " + deadlinetime + ":00"
	fmt.Println(deadline)
	
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	_, err = db.Exec("UPDATE tasks SET title=?,is_done=?,description=?,priority=?,deadline=? WHERE id=?", title, is_done_bool, description,priority,deadline, id)
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
	userID := sessions.Default(ctx).Get("user")
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
	// _, err = db.Exec("DELETE FROM tasks WHERE id=?", id)
	// if err != nil {
	// 	Error(http.StatusInternalServerError, err.Error())(ctx)
	// 	return
	// }
	_, err = db.Exec("DELETE tasks,ownership FROM tasks INNER JOIN ownership ON task_id = id WHERE task_id =? AND ownership.user_id = ?",id,userID)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/list")
		// Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Redirect to /list
	ctx.Redirect(http.StatusFound, "/list")
}

func ShareTaskForm(ctx *gin.Context){
	userID := sessions.Default(ctx).Get("user")
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

	err = db.Get(&task, "SELECT id, title, created_at, is_done, description, priority, deadline FROM tasks INNER JOIN ownership ON task_id = id WHERE task_id =? AND ownership.user_id = ?", id,userID)
	if err != nil {
		// Error(http.StatusBadRequest, err.Error())(ctx)
		ctx.Redirect(http.StatusFound, "/list")
		return
	}

	// Render edit form
	ctx.HTML(http.StatusOK, "form_share_task.html",
		gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task})
}

func UpdateShareTask(ctx *gin.Context){
	// Get task id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get share username 
	// requireなので絶対に存在する
	username, _ := ctx.GetPostForm("username")
	
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

    // ユーザの取得
    var user database.User
    err = db.Get(&user, "SELECT id, name FROM users WHERE name = ?", username)
    if err != nil {
		userID := sessions.Default(ctx).Get("user")
		// Get target task
		var task database.Task
		err = db.Get(&task, "SELECT id, title, created_at, is_done, description, priority, deadline FROM tasks INNER JOIN ownership ON task_id = id WHERE task_id =? AND ownership.user_id = ?", id,userID)
		if err != nil {
			ctx.Redirect(http.StatusFound, "/list")
			return
		}
		ctx.HTML(http.StatusBadRequest, "form_share_task.html", 
			gin.H{"Title": "", "Username": username, "Task": task, "Error": "No such user"})
        return
    }

	_, err = db.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", user.ID, id)
    if err != nil {
		userID := sessions.Default(ctx).Get("user")
		// Get target task
		var task database.Task
		err = db.Get(&task, "SELECT id, title, created_at, is_done, description, priority, deadline FROM tasks INNER JOIN ownership ON task_id = id WHERE task_id =? AND ownership.user_id = ?", id,userID)
		if err != nil {
			ctx.Redirect(http.StatusFound, "/list")
			return
		}
		ctx.HTML(http.StatusBadRequest, "form_share_task.html", 
		gin.H{"Title": "", "Username": username, "Task": task, "Error": "this task has already shared with the person"})
        return
    }
	// Render status
	path := "/list" // デフォルトではタスク一覧ページへ戻る
	// if id, err := result.LastInsertId(); err == nil {
	path = fmt.Sprintf("/task/%d", id) // 正常にIDを取得できた場合は /task/<id> へ戻る
	// }
	ctx.Redirect(http.StatusFound, path)
}

func DeleteShareTask(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
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
	
	// 共有しているかチェック→してなければ消えないようにする
	var duplicate int
	err = db.Get(&duplicate, "SELECT COUNT(*) FROM ownership WHERE task_id=?", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
    if duplicate < 2 {
		// Get target task
		var task database.Task
		err = db.Get(&task, "SELECT id, title, created_at, is_done, description, priority, deadline FROM tasks INNER JOIN ownership ON task_id = id WHERE task_id =? AND ownership.user_id = ?", id,userID)
		if err != nil {
			ctx.Redirect(http.StatusFound, "/list")
			return
		}
		ctx.HTML(http.StatusBadRequest, "form_share_task.html", gin.H{"Title": "", "Task": task, 
		"Error": "this task has not shared yet. you should delete task or share it with someone"})
        return
    }
	_, err = db.Exec("DELETE FROM ownership WHERE task_id =? AND user_id = ?",id,userID)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/list")
		return
	}
	// Redirect to /list
	ctx.Redirect(http.StatusFound, "/list")
}