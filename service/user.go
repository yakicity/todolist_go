package service

import (
	"crypto/sha256"
	"net/http"
	"encoding/hex"
	"regexp"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	database "todolist.go/db"
)

func NewUserForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "new_user_form.html", gin.H{"Title": "Register user"})
}

func hash(pw string) []byte {
	const salt = "todolist.go#"
	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte(pw))
	return h.Sum(nil)
}

func RegisterUser(ctx *gin.Context) {
	// フォームデータの受け取り
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	passwordForConfirm := ctx.PostForm("passwordForConfirm")
	switch {
	case username == "":
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Usernane is not provided", "Username": username})
	case password == "":
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is not provided", "Password": password})
	case passwordForConfirm == "":
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password for confirmation is not provided", "PasswordForConfirm": passwordForConfirm})
	}

	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 重複チェック
	var duplicate int
	err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	if duplicate > 0 {

		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username})
		return
	}

	// パスワード二回入力があっているか確認
	if password != passwordForConfirm {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password for confirmation is not correct ", "Username": username})
		return
	}
	if !matchPassword(password) {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password does not meet the conditions", "Username": username})
		return
	}

	// DB への保存
	result, err := db.Exec("INSERT INTO users(name, password) VALUES (?, ?)", username, hash(password))
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 保存状態の確認
	id, _ := result.LastInsertId()
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// ctx.JSON(http.StatusOK, user)
	// ctx.HTML(http.StatusOK, "index.html", gin.H{"Title": "Task list"})
	ctx.Redirect(http.StatusFound, "/login")
}

func matchPassword(password string) bool {
	if len(password) < 8 { // 6文字以上か判定
		return false
	}
	if !(regexp.MustCompile("^[0-9a-zA-Z!-/:-@[-`{-~]+$").Match([]byte(password))) { // 英数字記号以外を使っているか判定
		return false
	}
	reg := []*regexp.Regexp{
		regexp.MustCompile(`[[:alpha:]]`), // 英字が含まれるか判定
		regexp.MustCompile(`[[:digit:]]`), // 数字が含まれるか判定
		// regexp.MustCompile([[:punct:]]), // 記号が含まれるか判定
	}
	for _, r := range reg {
		if r.FindString(password) == "" {
			return false
		}
	}
	return true
}


//ログイン

func LoginPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", gin.H{"Title": "Register user"})
}


const userkey = "user"
 
func Login(ctx *gin.Context) {
    username := ctx.PostForm("username")
    password := ctx.PostForm("password")
 
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
 
    // ユーザの取得
    var user database.User
    err = db.Get(&user, "SELECT id, name, password FROM users WHERE name = ?", username)
    if err != nil {
        ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
        return
    }
 
    // パスワードの照合
    if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
        ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "Incorrect password"})
        return
    }
 
    // セッションの保存
    session := sessions.Default(ctx)
    session.Set(userkey, user.ID)
    session.Save()
 
    ctx.Redirect(http.StatusFound, "/list")
}

// ログインチェック
func LoginCheck(ctx *gin.Context) {
    if sessions.Default(ctx).Get(userkey) == nil {
        // ログイン状態
		ctx.Redirect(http.StatusFound, "/login")
        ctx.Abort()
    } else {
		// 非ログイン
        ctx.Next()
    }
}

func Logout(ctx *gin.Context) {
    session := sessions.Default(ctx)
    session.Clear()
    session.Options(sessions.Options{MaxAge: -1})
    session.Save()
    ctx.Redirect(http.StatusFound, "/login")
}

func EditUserForm(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var user database.User
	fmt.Println(userID)
	err = db.Get(&user, "SELECT id, name FROM users WHERE id =?", userID) // Use DB#Get for one entry
	if err != nil {
		// Error(http.StatusBadRequest, err.Error())(ctx)
		ctx.Redirect(http.StatusFound, "/")
		return
	}
	// Render edit form
	ctx.HTML(http.StatusOK, "form_edit_user.html",
		gin.H{"Title": fmt.Sprintf("Edit user %d", user.ID), "User": user})
}

func EditUserNameForm(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var user database.User
	fmt.Println(userID)
	err = db.Get(&user, "SELECT id, name FROM users WHERE id =?", userID) // Use DB#Get for one entry
	if err != nil {
		// Error(http.StatusBadRequest, err.Error())(ctx)
		ctx.Redirect(http.StatusFound, "/")
		return
	}
	fmt.Println("YYYYYYYYYYYY")
	// Render edit form
	ctx.HTML(http.StatusOK, "form_edit_user_name.html",
		gin.H{"Title": fmt.Sprintf("Edit user %d", user.ID), "User": user})
}

func UpdateUserName(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")

    username := ctx.PostForm("username")
    password := ctx.PostForm("password")

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 重複チェック
	var duplicate int
	err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	if duplicate > 0 {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username})
		return
	}
	if !matchPassword(password) {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password does not meet the conditions", "Username": username})
		return
	}
	fmt.Println("EEEEEEEEEEEE")
	// DB への保存
	_, err = db.Exec("UPDATE users SET name=? WHERE id=? AND password=?", username, userID,hash(password))
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	fmt.Println("EEEEEEEEEEEE")
	ctx.Redirect(http.StatusFound, "/")
}

func EditUserPasswordForm(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var user database.User
	fmt.Println(userID)
	err = db.Get(&user, "SELECT id, name FROM users WHERE id =?", userID) // Use DB#Get for one entry
	if err != nil {
		// Error(http.StatusBadRequest, err.Error())(ctx)
		ctx.Redirect(http.StatusFound, "/")
		return
	}
	fmt.Println("YYYYYYYYYYYY")
	// Render edit form
	ctx.HTML(http.StatusOK, "form_edit_user_password.html",
		gin.H{"Title": fmt.Sprintf("Edit user %d", user.ID), "User": user})
}

func UpdateUserPassword(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")

    oldpassword := ctx.PostForm("oldpassword")
	password := ctx.PostForm("password")
	passwordForConfirm := ctx.PostForm("passwordForConfirm")

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	if password != passwordForConfirm {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password does not meet the conditions"})
		return
	}
	if !matchPassword(password) {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password does not meet the conditions"})
		return
	}
	// DB への保存
	_, err = db.Exec("UPDATE users SET password=? WHERE id=? AND password=?", hash(password), userID,hash(oldpassword))
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.Redirect(http.StatusFound, "/")
}


func DeleteUser(ctx *gin.Context){
	//ログインしてる前提で
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	_, err = db.Exec("DELETE tasks, ownership FROM tasks INNER JOIN ownership ON task_id = id WHERE ownership.user_id = ?",userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	_, err = db.Exec("DELETE FROM users WHERE id=?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// logout
	Logout(ctx)
}