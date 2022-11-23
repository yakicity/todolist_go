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
	is_done := ctx.Query("is_done")
	isnot_done := ctx.Query("isnot_done")
	SelectStatus := "完了条件なし"
	// Get tasks in DB
	fmt.Println(is_done)
	fmt.Println(isnot_done)
	
	var tasks []database.Task
    query := "SELECT id, title, created_at, is_done, description, priority, deadline FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ?"
    switch {
    case kw != "":
        err = db.Select(&tasks, query + " AND title LIKE ?", userID, "%" + kw + "%")
	case (is_done == "t" && isnot_done == "f"):
		err = db.Select(&tasks, query, userID)
		SelectStatus = "全タスク"
	case (is_done == "t" && isnot_done == "") || (is_done == "t" && isnot_done == "t"):
		is_done_bool, err := strconv.ParseBool(is_done)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
		err = db.Select(&tasks, query + " AND is_done LIKE ?", userID, is_done_bool)
		SelectStatus = "完了タスク"
	case (is_done == "" && isnot_done == "f") || (is_done == "f" && isnot_done == "f"):
		isnot_done_bool, err := strconv.ParseBool(isnot_done)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
		err = db.Select(&tasks, query + " AND is_done LIKE ?", userID, isnot_done_bool)
		SelectStatus = "未完了タスク"
    default:
		err = db.Select(&tasks, query, userID)
		SelectStatus = "全タスク"
		fmt.Println("default case")
    }
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Render tasks
	ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks, "Kw": kw, "SelectStatus": SelectStatus})
}


// type ownership struct {
// 	user_id    int
// 	task_id    int
// }

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
	// err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id) // Use DB#Get for one entry
	// if err != nil {
	// 	Error(http.StatusBadRequest, err.Error())(ctx)
	// 	return
	// }
	err = db.Get(&task, "SELECT id, title, created_at, is_done, description, priority, deadline FROM tasks INNER JOIN ownership ON task_id = id WHERE task_id =? AND ownership.user_id = ?", id,userID) // Use DB#Get for one entry
	if err != nil {
		// Error(http.StatusBadRequest, err.Error())(ctx)
		ctx.Redirect(http.StatusFound, "/list")
		return
	}

	// rows, err := db.Query("SELECT * FROM ownership WHERE task_id =?",id)
	// if err != nil {
	// 	fmt.Println("YYYYYYYYYYYY")
	// 	Error(http.StatusInternalServerError, err.Error())(ctx)
	// 	return
	// }
	// task_user_id := 0
	// for rows.Next() {
	// 	owner := ownership{}
	// 	rows.Scan(&owner.user_id, &owner.task_id)
	// 	fmt.Println(owner.user_id)
	// 	task_user_id = owner.user_id
	// }
	// u_id_string := strconv.Itoa(u_id)
	// u_id_u64,_ :=strconv.ParseUint(u_id_string, 10, 64)
	// if task_user_id - userID == 0{
	// 	fmt.Println("EEEEEEEEEEE")
	// 	ctx.Redirect(http.StatusFound, "/list")
	// }
	// fmt.Println(task_user_id)

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
	fmt.Println(deadline)
	if deadlinedate == "" || deadlinetime == "" {
		deadline = ""
		fmt.Println(deadline)
	}

	// Get DB connection
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    tx := db.MustBegin()
	var result sql.Result
	
	if (deadlinedate == "" || deadlinetime == ""){
		result, err = tx.Exec("INSERT INTO tasks (title, description, priority) VALUES (?,?,?)", title, description,priority)
	} else {
    	result, err = tx.Exec("INSERT INTO tasks (title, description, priority,deadline) VALUES (?,?,?,?)", title, description,priority,deadline)
	}
	// result, err := tx.Exec("INSERT INTO tasks (title, description, priority,deadline) VALUES (?,?,?,?)", title, description,priority,deadline)
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
	// db, err := database.GetConnection()
	// if err != nil {
	// 	Error(http.StatusInternalServerError, err.Error())(ctx)
	// 	return
	// }
	// // Create new data with given title on DB
	// result, err := db.Exec("INSERT INTO tasks (title, description) VALUES (?,?)", title, description)
	// if err != nil {
	// 	Error(http.StatusInternalServerError, err.Error())(ctx)
	// 	return
	// }
	// // Render status
	// path := "/list" // デフォルトではタスク一覧ページへ戻る
	// if id, err := result.LastInsertId(); err == nil {
	// 	path = fmt.Sprintf("/task/%d", id) // 正常にIDを取得できた場合は /task/<id> へ戻る
	// }
	// ctx.Redirect(http.StatusFound, path)
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

	err = db.Get(&task, "SELECT id, title, created_at, is_done, description, priority, deadline FROM tasks INNER JOIN ownership ON task_id = id WHERE task_id =? AND ownership.user_id = ?", id,userID) // Use DB#Get for one entry
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

