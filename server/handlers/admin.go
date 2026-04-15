package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// 转换日期格式为 datetime-local 需要的格式 (YYYY-MM-DDTHH:mm)
func formatDateTimeLocal(dateStr string) string {
	// 数据库返回格式：2006-01-02 15:04:05
	// datetime-local 需要格式：2006-01-02T15:04
	if len(dateStr) < 16 {
		return dateStr
	}
	// 将空格替换为 T，并只保留到分钟
	return strings.Replace(dateStr[:16], " ", "T", 1)
}

func stringPtr(s string) *string {
	return &s
}

// Activity 活动模型
type Activity struct {
	ID              int      `json:"id"`
	Title           string   `json:"title"`
	Cover           string   `json:"cover"`
	Date            string   `json:"date"`
	Location        string   `json:"location"`
	MaxParticipants int      `json:"maxParticipants"`
	Participants    int      `json:"participants"`
	Price           float64  `json:"price"`
	Description     string   `json:"description"`
	Status          int      `json:"status"`
	SignupEndTime   *string  `json:"signupEndTime,omitempty"`
	CreatedAt       string   `json:"createdAt"`
	CreatedBy       string   `json:"createdBy"`
}

// Login 登录页面
func LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"error": c.Query("error"),
	})
}

// LoginPost 登录处理
func LoginPost(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// 简单验证，生产环境应该查数据库并哈希密码
	if username == "admin" && password == "admin123" {
		session, _ := store.Get(c.Request, "admin-session")
		session.Values["authenticated"] = true
		session.Values["username"] = username
		session.Save(c.Request, c.Writer)
		c.Redirect(http.StatusFound, "/admin/activities")
	} else {
		c.Redirect(http.StatusFound, "/admin/login?error=用户名或密码错误")
	}
}

// LoginAPI JSON 登录接口（供 Vue 前端使用）
func LoginAPI(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "请输入用户名和密码"})
		return
	}

	if input.Username == "admin" && input.Password == "admin123" {
		session, _ := store.Get(c.Request, "admin-session")
		session.Values["authenticated"] = true
		session.Values["username"] = input.Username
		session.Save(c.Request, c.Writer)
		c.JSON(http.StatusOK, gin.H{"success": true})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "用户名或密码错误"})
	}
}

// Logout 登出
func Logout(c *gin.Context) {
	session, _ := store.Get(c.Request, "admin-session")
	session.Values["authenticated"] = false
	session.Save(c.Request, c.Writer)
	c.Redirect(http.StatusFound, "/admin/login")
}

// AuthMiddleware 登录验证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, _ := store.Get(c.Request, "admin-session")
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}
		c.Set("username", session.Values["username"])
		c.Next()
	}
}

// AdminIndex 后台首页
func AdminIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_index.html", gin.H{
		"username": c.GetString("username"),
	})
}

// AdminActivities 活动管理列表
func AdminActivities(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT id, title, cover, date, location, max_participants, participants,
			       price, description, status, created_at, created_by
			FROM activities
			ORDER BY date DESC
		`)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var activities []Activity
		for rows.Next() {
			var a Activity
			var cover, createdBy sql.NullString
			err := rows.Scan(&a.ID, &a.Title, &cover, &a.Date, &a.Location,
				&a.MaxParticipants, &a.Participants, &a.Price, &a.Description,
				&a.Status, &a.CreatedAt, &a.CreatedBy)
			if err != nil {
				continue
			}
			a.Cover = cover.String
			a.CreatedBy = createdBy.String
			activities = append(activities, a)
		}

		c.HTML(http.StatusOK, "activities.html", gin.H{
			"username":   c.GetString("username"),
			"activities": activities,
		})
	}
}

// AdminNewActivity 新建活动页面
func AdminNewActivity(c *gin.Context) {
	c.HTML(http.StatusOK, "activity_form.html", gin.H{
		"username": c.GetString("username"),
		"action":   "new",
	})
}

// AdminCreateActivity 创建活动
func AdminCreateActivity(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		title := c.PostForm("title")
		location := c.PostForm("location")
		date := c.PostForm("date")
		maxParticipants := c.PostForm("maxParticipants")
		price := c.PostForm("price")
		description := c.PostForm("description")
		signupEndTime := c.PostForm("signupEndTime")

		maxP, _ := strconv.Atoi(maxParticipants)
		priceF, _ := strconv.ParseFloat(price, 64)

		var signupEndTimeVal interface{}
		if signupEndTime == "" {
			signupEndTimeVal = nil
		} else {
			signupEndTimeVal = signupEndTime
		}

		_, err := db.Exec(`
			INSERT INTO activities (title, cover, date, location, max_participants, price, description, signup_end_time, created_by, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0, 0)
		`, title, "", date, location, maxP, priceF, description, signupEndTimeVal)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusFound, "/admin/activities")
	}
}

// AdminEditActivity 编辑活动页面
func AdminEditActivity(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var a Activity
		var cover, createdBy, signupEndTime, dateStr sql.NullString
		err := db.QueryRow(`
			SELECT id, title, cover, date, location, max_participants, participants,
			       price, description, status, created_at, created_by, signup_end_time
			FROM activities WHERE id = $1
		`, id).Scan(&a.ID, &a.Title, &cover, &dateStr, &a.Location,
			&a.MaxParticipants, &a.Participants, &a.Price, &a.Description,
			&a.Status, &a.CreatedAt, &createdBy, &signupEndTime)

		if err == sql.ErrNoRows {
			c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "活动不存在"})
			return
		}

		// 转换日期格式为 datetime-local 需要的格式 (YYYY-MM-DDTHH:mm)
		if dateStr.Valid {
			a.Date = formatDateTimeLocal(dateStr.String)
		}
		a.Cover = cover.String
		a.CreatedBy = createdBy.String
		if signupEndTime.Valid {
			a.SignupEndTime = stringPtr(formatDateTimeLocal(signupEndTime.String))
		}
		c.HTML(http.StatusOK, "activity_form.html", gin.H{
			"username": c.GetString("username"),
			"action":   "edit",
			"activity": a,
		})
	}
}

// AdminUpdateActivity 更新活动
func AdminUpdateActivity(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		title := c.PostForm("title")
		location := c.PostForm("location")
		date := c.PostForm("date")
		maxParticipants := c.PostForm("maxParticipants")
		price := c.PostForm("price")
		description := c.PostForm("description")
		status := c.PostForm("status")
		signupEndTime := c.PostForm("signupEndTime")

		maxP, _ := strconv.Atoi(maxParticipants)
		priceF, _ := strconv.ParseFloat(price, 64)
		statusI, _ := strconv.Atoi(status)

		var signupEndTimeVal interface{}
		if signupEndTime == "" {
			signupEndTimeVal = nil
		} else {
			signupEndTimeVal = signupEndTime
		}

		_, err := db.Exec(`
			UPDATE activities
			SET title=$1, location=$2, date=$3, max_participants=$4, price=$5,
			    description=$6, status=$7, signup_end_time=$8
			WHERE id=$9
		`, title, location, date, maxP, priceF, description, statusI, signupEndTimeVal, id)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusFound, "/admin/activities")
	}
}

// AdminDeleteActivity 删除活动
func AdminDeleteActivity(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM activities WHERE id = $1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
	}
}

// AdminSignups 报名管理
func AdminSignups(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		activityID := c.Param("id")

		// 获取活动信息
		var activity Activity
		err := db.QueryRow(`
			SELECT id, title, location, date, max_participants, participants
			FROM activities WHERE id = $1
		`, activityID).Scan(&activity.ID, &activity.Title, &activity.Location,
			&activity.Date, &activity.MaxParticipants, &activity.Participants)

		if err == sql.ErrNoRows {
			c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "活动不存在"})
			return
		}

		// 获取报名列表（包含用户真实信息）
		rows, err := db.Query(`
			SELECT s.id, s.user_id, s.name, s.phone, s.emergency_contact, s.emergency_phone, s.status, s.created_at,
			       COALESCE(u.rz_real_name, '') as real_name, COALESCE(u.phone, '') as user_phone
			FROM signups s
			LEFT JOIN users u ON s.user_id = u.id
			WHERE s.activity_id = $1
			ORDER BY s.created_at DESC
		`, activityID)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var signups []gin.H
		for rows.Next() {
			var id, userID, status int
			var name, phone, emergencyContact, emergencyPhone, createdAt, realName, userPhone string
			rows.Scan(&id, &userID, &name, &phone, &emergencyContact, &emergencyPhone, &status, &createdAt, &realName, &userPhone)
			signups = append(signups, gin.H{
				"id":              id,
				"userId":          userID,
				"name":            name,
				"phone":           phone,
				"emergencyContact": emergencyContact,
				"emergencyPhone":   emergencyPhone,
				"status":          status,
				"createdAt":       createdAt,
				"realName":        realName,
				"userPhone":       userPhone,
			})
		}

		c.HTML(http.StatusOK, "signups.html", gin.H{
			"username":  c.GetString("username"),
			"activity":  activity,
			"signups":   signups,
		})
	}
}
