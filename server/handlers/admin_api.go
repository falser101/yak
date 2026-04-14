package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminStats 统计数据
type AdminStats struct {
	TotalActivities int `json:"totalActivities"`
	TotalSignups    int `json:"totalSignups"`
	TotalUsers      int `json:"totalUsers"`
}

// AdminListActivities 获取活动列表（管理后台用）
func AdminListActivities(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		offset := (page - 1) * pageSize

		// 获取总数
		var total int
		db.QueryRow(`SELECT COUNT(*) FROM activities`).Scan(&total)

		// 获取活动列表
		rows, err := db.Query(`
			SELECT
				a.id, a.title, a.cover, a.date, a.location,
				a.max_participants, a.price, a.description, a.status, a.created_at, a.signup_end_time,
				COALESCE(u.id::text, '0')::int as created_by_id,
				COALESCE(u.nickname, '') as created_by_name,
				COALESCE(signup_count.cnt, 0) as signup_count
			FROM activities a
			LEFT JOIN users u ON CASE WHEN a.created_by ~ '^[0-9]+$' THEN a.created_by::int ELSE 0 END = u.id
			LEFT JOIN (
				SELECT activity_id, COUNT(*) as cnt
				FROM signups WHERE status = 1
				GROUP BY activity_id
			) signup_count ON a.id = signup_count.activity_id
			ORDER BY a.date DESC
			LIMIT $1 OFFSET $2
		`, pageSize, offset)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		type ActivityItem struct {
			ID              int        `json:"id"`
			Title           string     `json:"title"`
			Cover           string     `json:"cover"`
			Date            string     `json:"date"`
			Location        string     `json:"location"`
			MaxParticipants int        `json:"maxParticipants"`
			SignupCount     int        `json:"signupCount"`
			Price           float64    `json:"price"`
			Description     string     `json:"description"`
			Status          int        `json:"status"`
			SignupEndTime   *string    `json:"signupEndTime,omitempty"`
			CreatedAt       time.Time  `json:"createdAt"`
			CreatedByID     int        `json:"createdById"`
			CreatedByName   string     `json:"createdByName"`
		}

		var activities []ActivityItem
		for rows.Next() {
			var a ActivityItem
			var signupEndTime sql.NullString
			err := rows.Scan(&a.ID, &a.Title, &a.Cover, &a.Date, &a.Location,
				&a.MaxParticipants, &a.Price, &a.Description, &a.Status, &a.CreatedAt, &signupEndTime,
				&a.CreatedByID, &a.CreatedByName, &a.SignupCount)
			if err != nil {
				continue
			}
			if signupEndTime.Valid {
				a.SignupEndTime = &signupEndTime.String
			}
			activities = append(activities, a)
		}

		c.JSON(http.StatusOK, gin.H{
			"data":     activities,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		})
	}
}

// AdminGetActivity 获取活动详情
func AdminGetActivity(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		type ActivityDetail struct {
			ID              int        `json:"id"`
			Title           string     `json:"title"`
			Cover           string     `json:"cover"`
			Date            string     `json:"date"`
			Location        string     `json:"location"`
			MaxParticipants int        `json:"maxParticipants"`
			SignupCount     int        `json:"signupCount"`
			Price           float64    `json:"price"`
			Description     string     `json:"description"`
			Status          int        `json:"status"`
			SignupEndTime   *string    `json:"signupEndTime,omitempty"`
			CreatedAt       time.Time  `json:"createdAt"`
			CreatedByID     int        `json:"createdById"`
			CreatedByName   string     `json:"createdByName"`
		}

		var a ActivityDetail
		var signupEndTime sql.NullString
		err := db.QueryRow(`
			SELECT
				a.id, a.title, a.cover, a.date, a.location,
				a.max_participants, a.price, a.description, a.status, a.created_at, a.signup_end_time,
				COALESCE(u.id::text, '0')::int, COALESCE(u.nickname, ''),
				COALESCE((
					SELECT COUNT(*) FROM signups WHERE activity_id = a.id AND status = 1
				), 0)
			FROM activities a
			LEFT JOIN users u ON CASE WHEN a.created_by ~ '^[0-9]+$' THEN a.created_by::int ELSE 0 END = u.id
			WHERE a.id = $1
		`, id).Scan(
			&a.ID, &a.Title, &a.Cover, &a.Date, &a.Location,
			&a.MaxParticipants, &a.Price, &a.Description, &a.Status, &a.CreatedAt, &signupEndTime,
			&a.CreatedByID, &a.CreatedByName, &a.SignupCount)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "活动不存在"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if signupEndTime.Valid {
			a.SignupEndTime = &signupEndTime.String
		}

		c.JSON(http.StatusOK, gin.H{"data": a})
	}
}

// AdminCreateActivityJSON 创建活动
func AdminCreateActivityJSON(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Title           string   `json:"title"`
			Cover           string   `json:"cover"`
			Date            string   `json:"date"`
			Location        string   `json:"location"`
			MaxParticipants int      `json:"maxParticipants"`
			Price           float64  `json:"price"`
			Description     string   `json:"description"`
			SignupEndTime   *string  `json:"signupEndTime,omitempty"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var id int
		err := db.QueryRow(`
			INSERT INTO activities (title, cover, date, location, max_participants, price, description, signup_end_time, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`, input.Title, input.Cover, input.Date, input.Location,
			input.MaxParticipants, input.Price, input.Description, input.SignupEndTime, 1).Scan(&id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id}})
	}
}

// AdminUpdateActivityJSON 更新活动
func AdminUpdateActivityJSON(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var input struct {
			Title           string  `json:"title"`
			Cover           string  `json:"cover"`
			Date            string  `json:"date"`
			Location        string  `json:"location"`
			MaxParticipants int     `json:"maxParticipants"`
			Price           float64 `json:"price"`
			Description     string  `json:"description"`
			Status          int     `json:"status"`
			SignupEndTime   *string `json:"signupEndTime,omitempty"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec(`
			UPDATE activities
			SET title=$1, cover=$2, date=$3, location=$4, max_participants=$5,
			    price=$6, description=$7, status=$8, signup_end_time=$9
			WHERE id=$10
		`, input.Title, input.Cover, input.Date, input.Location,
			input.MaxParticipants, input.Price, input.Description, input.Status, input.SignupEndTime, id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
	}
}

// AdminDeleteActivityJSON 删除活动
func AdminDeleteActivityJSON(db *sql.DB) gin.HandlerFunc {
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

// AdminSignupItem 报名项
type AdminSignupItem struct {
	ID               int    `json:"id"`
	UserID           int    `json:"userId"`
	Name             string `json:"name"`
	Phone            string `json:"phone"`
	EmergencyContact string `json:"emergencyContact"`
	EmergencyPhone   string `json:"emergencyPhone"`
	Remark           string `json:"remark"`
	Status           int    `json:"status"`
	CreatedAt        string `json:"createdAt"`
	Nickname         string `json:"nickname"`
	Avatar           string `json:"avatar"`
}

// AdminGetSignups 获取活动报名列表
func AdminGetSignups(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		activityID := c.Param("id")

		rows, err := db.Query(`
			SELECT s.id, s.user_id, s.name, s.phone, s.emergency_contact, s.emergency_phone,
			       s.remark, s.status, s.created_at, COALESCE(u.nickname, ''), COALESCE(u.avatar, '')
			FROM signups s
			LEFT JOIN users u ON s.user_id = u.id
			WHERE s.activity_id = $1
			ORDER BY s.created_at DESC
		`, activityID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var signups []AdminSignupItem
		for rows.Next() {
			var s AdminSignupItem
			rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Phone, &s.EmergencyContact, &s.EmergencyPhone,
				&s.Remark, &s.Status, &s.CreatedAt, &s.Nickname, &s.Avatar)
			signups = append(signups, s)
		}

		c.JSON(http.StatusOK, gin.H{"data": signups})
	}
}

// AdminGetStats 获取统计数据
func AdminGetStats(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var stats AdminStats

		db.QueryRow(`SELECT COUNT(*) FROM activities`).Scan(&stats.TotalActivities)
		db.QueryRow(`SELECT COUNT(*) FROM signups WHERE status = 1`).Scan(&stats.TotalSignups)
		db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)

		c.JSON(http.StatusOK, gin.H{"data": stats})
	}
}
