package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/lib/pq"
)

var store = sessions.NewCookieStore([]byte("yak-admin-secret-key-change-in-prod"))

// 微信配置（从环境变量读取）
var (
	WechatAppID     = getEnv("WECHAT_APPID", "wx8f1bbd8ea3190d2c")
	WechatSecret    = getEnv("WECHAT_SECRET", "1d6fc393c5a0aa1f726459b10e89c2dd")
	WechatLoginURL  = "https://api.weixin.qq.com/sns/jscode2session"
	sessionKeyStore = make(map[int]string) // userID -> sessionKey
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
type User struct {
	ID               int       `json:"id"`
	Openid           string    `json:"openid"`
	Nickname         string    `json:"nickname"`
	Avatar           string    `json:"avatar"`
	Phone            string    `json:"phone"`
	CreatedAt        time.Time `json:"createdAt"`
	RzStatus         int       `json:"rzStatus"`          // 0:未认证, 1:认证中, 2:已认证
	RzRealName       string    `json:"rzRealName"`        // 真实姓名
	RzGender         int       `json:"rzGender"`          // 性别: 0:未知, 1:男, 2:女
	RzEmergencyName  string    `json:"rzEmergencyName"`   // 紧急联系人
	RzEmergencyPhone string    `json:"rzEmergencyPhone"`  // 紧急联系人电话
}

// ============ 微信登录相关 ============

// WechatSession 微信返回的 session
type WechatSession struct {
	Openid     string `json:"openid"`
	SessionKey string `json:"session_key"`
	Unionid    string `json:"unionid"`
	Errcode    int    `json:"errcode"`
	Errmsg     string `json:"errmsg"`
}

// Login 小程序登录接口
func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Code string `json:"code"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
			return
		}

		code := input.Code

		// 开发环境：直接返回测试用户
		if code == "dev_test" {
			user := getTestUser(db)
			c.SetCookie("session", fmt.Sprintf("%d", user.ID), 86400*7, "/", "", false, true)
			c.JSON(http.StatusOK, gin.H{
				"user":  user,
				"isDev": true,
			})
			return
		}

		// 生产环境：真实微信登录
		user, err := loginWithWechat(db, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.SetCookie("session", fmt.Sprintf("%d", user.ID), 86400*7, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"user": user})
	}
}

// loginWithWechat 微信登录
func loginWithWechat(db *sql.DB, code string) (*User, error) {
	// 请求微信接口
	url := fmt.Sprintf("%s?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		WechatLoginURL, WechatAppID, WechatSecret, code)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求微信接口失败：%v", err)
	}
	defer resp.Body.Close()

	var ws WechatSession
	if err := json.NewDecoder(resp.Body).Decode(&ws); err != nil {
		return nil, fmt.Errorf("解析微信响应失败：%v", err)
	}

	if ws.Errcode != 0 {
		return nil, fmt.Errorf("微信登录失败：%s", ws.Errmsg)
	}

	// 查询或创建用户
	user, err := getOrCreateUser(db, ws.Openid)
	if err != nil {
		return nil, err
	}

	// 存储 session_key 用于后续解密手机号
	sessionKeyStore[user.ID] = ws.SessionKey

	return user, nil
}

// getOrCreateUser 查询或创建用户
func getOrCreateUser(db *sql.DB, openid string) (*User, error) {
	var user User
	var nickname, avatar, phone, rzRealName, rzEmergencyName, rzEmergencyPhone sql.NullString
	var rzStatus, rzGender sql.NullInt64
	var rzVerifiedAt sql.NullTime
	err := db.QueryRow(`
		SELECT id, openid, nickname, avatar, phone, created_at,
		       COALESCE(rz_status, 0), COALESCE(rz_real_name, ''), COALESCE(rz_gender, 0),
		       COALESCE(rz_emergency_name, ''), COALESCE(rz_emergency_phone, ''), rz_verified_at
		FROM users WHERE openid = $1
	`, openid).Scan(&user.ID, &user.Openid, &nickname, &avatar, &phone, &user.CreatedAt,
		&rzStatus, &rzRealName, &rzGender, &rzEmergencyName, &rzEmergencyPhone, &rzVerifiedAt)

	if err == nil {
		user.Nickname = nickname.String
		user.Avatar = avatar.String
		user.Phone = phone.String
		user.RzStatus = int(rzStatus.Int64)
		user.RzRealName = rzRealName.String
		user.RzGender = int(rzGender.Int64)
		user.RzEmergencyName = rzEmergencyName.String
		user.RzEmergencyPhone = rzEmergencyPhone.String
		return &user, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// 创建新用户
	var nickname2, avatar2, phone2 sql.NullString
	err = db.QueryRow(`
		INSERT INTO users (openid, nickname, avatar)
		VALUES ($1, $2, $3)
		RETURNING id, openid, nickname, avatar, phone, created_at
	`, openid, "微信用户", "").Scan(
		&user.ID, &user.Openid, &nickname2, &avatar2, &phone2, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	user.Nickname = nickname2.String
	user.Avatar = avatar2.String
	user.Phone = phone2.String
	user.RzStatus = 0

	return &user, nil
}

// getTestUser 获取测试用户
func getTestUser(db *sql.DB) *User {
	var user User
	var nickname, avatar, phone, rzRealName, rzEmergencyName, rzEmergencyPhone sql.NullString
	var rzStatus, rzGender sql.NullInt64
	err := db.QueryRow(`
		SELECT id, openid, nickname, avatar, phone, created_at,
		       COALESCE(rz_status, 0), COALESCE(rz_real_name, ''), COALESCE(rz_gender, 0),
		       COALESCE(rz_emergency_name, ''), COALESCE(rz_emergency_phone, '')
		FROM users WHERE openid = $1
	`, "dev_test_user").Scan(
		&user.ID, &user.Openid, &nickname, &avatar, &phone, &user.CreatedAt,
		&rzStatus, &rzRealName, &rzGender, &rzEmergencyName, &rzEmergencyPhone)

	if err != nil {
		// 测试用户不存在，创建一个
		user = User{
			ID:       1,
			Openid:   "dev_test_user",
			Nickname: "测试用户",
			RzStatus: 2, // 默认已认证，方便测试
		}
	} else {
		user.Nickname = nickname.String
		user.Avatar = avatar.String
		user.Phone = phone.String
		user.RzStatus = int(rzStatus.Int64)
		user.RzRealName = rzRealName.String
		user.RzGender = int(rzGender.Int64)
		user.RzEmergencyName = rzEmergencyName.String
		user.RzEmergencyPhone = rzEmergencyPhone.String
	}
	return &user
}

// GetCurrentUser 获取当前登录用户
func GetCurrentUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("session")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		userID, _ := strconv.Atoi(cookie)
		var user User
		err = db.QueryRow(`
			SELECT id, openid, nickname, avatar, phone, created_at
			FROM users WHERE id = $1
		`, userID).Scan(
			&user.ID, &user.Openid, &user.Nickname, &user.Avatar, &user.Phone, &user.CreatedAt)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			c.Abort()
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("user", &user)
		c.Next()
	}
}

// ============ 活动相关 ============

// GetActivities 获取活动列表（支持分页）
func GetActivities(db *sql.DB) gin.HandlerFunc {
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
		err := db.QueryRow(`SELECT COUNT(*) FROM activities`).Scan(&total)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 获取当前用户 ID（用于判断是否已报名）
		currentUserID := 0
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && len(authHeader) > 7 {
			currentUserID, _ = strconv.Atoi(authHeader[7:])
		} else {
			cookie, _ := c.Cookie("session")
			if cookie != "" {
				currentUserID, _ = strconv.Atoi(cookie)
			}
		}

		// 获取活动列表（不含最近报名用户）
		rows, err := db.Query(`
			SELECT
				a.id, a.title, a.cover, a.date, a.location,
				a.max_participants, a.price, a.description, a.status, a.created_at, a.signup_end_time,
				CASE WHEN a.created_by ~ '^[0-9]+$' THEN a.created_by::int ELSE 0 END as created_by_id,
				COALESCE(u.nickname, '') as created_by_name,
				COALESCE(signup_count.cnt, 0) as signup_count,
				CASE WHEN $3 > 0 AND EXISTS(
					SELECT 1 FROM signups WHERE activity_id = a.id AND user_id = $3 AND status = 1
				) THEN true ELSE false END as is_signed_up
			FROM activities a
			LEFT JOIN users u ON CASE WHEN a.created_by ~ '^[0-9]+$' THEN a.created_by::int ELSE 0 END = u.id
			LEFT JOIN (
				SELECT activity_id, COUNT(*) as cnt
				FROM signups WHERE status = 1
				GROUP BY activity_id
			) signup_count ON a.id = signup_count.activity_id
			ORDER BY a.date DESC
			LIMIT $1 OFFSET $2
		`, pageSize, offset, currentUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		type ActivityItem struct {
			ID              int            `json:"id"`
			Title           string         `json:"title"`
			Cover           string         `json:"cover"`
			Date            string         `json:"date"`
			Location        string         `json:"location"`
			MaxParticipants int            `json:"maxParticipants"`
			SignupCount     int            `json:"signupCount"`
			Price           float64        `json:"price"`
			Description     string         `json:"description"`
			Status          int            `json:"status"`
			SignupEndTime   *time.Time     `json:"signupEndTime,omitempty"`
			CreatedAt       time.Time      `json:"createdAt"`
			CreatedByID     int            `json:"createdById"`
			CreatedByName   string         `json:"createdByName"`
			IsSignedUp      bool           `json:"isSignedUp"`
			RecentSignups   []SignupAvatar `json:"recentSignups"`
		}

		var activities []ActivityItem
		var activityIDs []int
		for rows.Next() {
			var a ActivityItem
			err := rows.Scan(&a.ID, &a.Title, &a.Cover, &a.Date, &a.Location,
				&a.MaxParticipants, &a.Price, &a.Description, &a.Status, &a.CreatedAt, &a.SignupEndTime,
				&a.CreatedByID, &a.CreatedByName, &a.SignupCount, &a.IsSignedUp)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			activities = append(activities, a)
			activityIDs = append(activityIDs, a.ID)
		}

		// 获取最近报名用户（每个活动最近 1 个）
		if len(activityIDs) > 0 {
			// 转换为 int32 数组
			activityIDArray := make([]int32, len(activityIDs))
			for i, id := range activityIDs {
				activityIDArray[i] = int32(id)
			}

			signupRows, err := db.Query(`
				SELECT s.activity_id, s.user_id, COALESCE(u.avatar, '') as avatar, COALESCE(u.nickname, '') as nickname
				FROM (
					SELECT DISTINCT ON (activity_id) activity_id, user_id, created_at
					FROM signups
					WHERE activity_id = ANY($1) AND status = 1
					ORDER BY activity_id, created_at DESC
				) s
				LEFT JOIN users u ON s.user_id = u.id
			`, pq.Int32Array(activityIDArray))
			if err == nil {
				defer signupRows.Close()
				signupMap := make(map[int][]SignupAvatar)
				for signupRows.Next() {
					var activityID, userID int
					var avatar, nickname string
					signupRows.Scan(&activityID, &userID, &avatar, &nickname)
					signupMap[activityID] = append(signupMap[activityID], SignupAvatar{
						UserID:   userID,
						Avatar:   avatar,
						Nickname: nickname,
					})
				}
				// 将最近报名用户填充到活动列表中
				for i := range activities {
					if signups, ok := signupMap[activities[i].ID]; ok {
						activities[i].RecentSignups = signups
					}
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"data":     activities,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		})
	}
}

// GetActivity 获取活动详情
func GetActivity(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// 获取当前用户ID（从 header 或 cookie）
		currentUserID := 0
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && len(authHeader) > 7 {
			currentUserID, _ = strconv.Atoi(authHeader[7:])
		} else {
			cookie, _ := c.Cookie("session")
			if cookie != "" {
				currentUserID, _ = strconv.Atoi(cookie)
			}
		}

		type ActivityDetail struct {
			ID              int            `json:"id"`
			Title           string         `json:"title"`
			Cover           string         `json:"cover"`
			Date            string         `json:"date"`
			Location        string         `json:"location"`
			MaxParticipants int            `json:"maxParticipants"`
			Participants    int            `json:"participants"`
			Price           float64        `json:"price"`
			Description     string         `json:"description"`
			Status          int            `json:"status"`
			SignupEndTime   *time.Time     `json:"signupEndTime,omitempty"`
			CreatedAt       time.Time      `json:"createdAt"`
			CreatedByID     int            `json:"createdById"`
			CreatedByName   string         `json:"createdByName"`
			IsSignedUp      bool           `json:"isSignedUp"`
			CurrentUser     *UserInfo      `json:"currentUser,omitempty"`
			AllSignups      []SignupDetail `json:"allSignups,omitempty"`
		}

		var a ActivityDetail
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
			&a.MaxParticipants, &a.Price, &a.Description, &a.Status, &a.CreatedAt, &a.SignupEndTime,
			&a.CreatedByID, &a.CreatedByName, &a.Participants)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "活动不存在"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 检查当前用户是否已报名
		if currentUserID > 0 {
			var count int
			db.QueryRow(`SELECT COUNT(*) FROM signups WHERE activity_id = $1 AND user_id = $2 AND status = 1`, id, currentUserID).Scan(&count)
			a.IsSignedUp = count > 0
		}

		// 获取所有报名用户列表
		signupRows, err := db.Query(`
			SELECT s.user_id, s.name, s.phone, s.created_at,
			       COALESCE(u.nickname, '') as nickname, COALESCE(u.avatar, '') as avatar
			FROM signups s
			LEFT JOIN users u ON s.user_id = u.id
			WHERE s.activity_id = $1 AND s.status = 1
			ORDER BY s.created_at DESC
		`, id)
		if err == nil {
			defer signupRows.Close()
			for signupRows.Next() {
				var s SignupDetail
				signupRows.Scan(&s.UserID, &s.Name, &s.Phone, &s.CreatedAt, &s.Nickname, &s.Avatar)
				a.AllSignups = append(a.AllSignups, s)
			}
		}

		c.JSON(http.StatusOK, gin.H{"data": a})
	}
}

type UserInfo struct {
	ID       int    `json:"id"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
}

// 报名用户头像
type SignupAvatar struct {
	UserID   int    `json:"userId"`
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
}

// 报名详情
type SignupDetail struct {
	UserID    int    `json:"userId"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar"`
	Nickname  string `json:"nickname"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"createdAt"`
}

// CreateActivity 创建活动
func CreateActivity(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Title           string     `json:"title"`
			Cover           string     `json:"cover"`
			Date            string     `json:"date"`
			Location        string     `json:"location"`
			MaxParticipants int        `json:"maxParticipants"`
			Price           float64    `json:"price"`
			Description     string     `json:"description"`
			SignupEndTime   *time.Time `json:"signupEndTime,omitempty"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 获取当前用户 ID
		var userID int
		cookie, _ := c.Cookie("session")
		if cookie != "" {
			userID, _ = strconv.Atoi(cookie)
		}

		var id int
		err := db.QueryRow(`
			INSERT INTO activities (title, cover, date, location, max_participants, price, description, signup_end_time, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`, input.Title, input.Cover, input.Date, input.Location,
			input.MaxParticipants, input.Price, input.Description, input.SignupEndTime, userID).Scan(&id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id}})
	}
}

// Signup 报名活动（带事务和锁）
func Signup(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		activityID := c.Param("id")

		var input struct {
			Name             string `json:"name"`
			Phone            string `json:"phone"`
			EmergencyContact string `json:"emergencyContact"`
			EmergencyPhone   string `json:"emergencyPhone"`
			Remark           string `json:"remark"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 获取当前用户 ID（从 header 或 cookie）
		userID := 0
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && len(authHeader) > 7 {
			// Bearer token 格式
			userID, _ = strconv.Atoi(authHeader[7:])
		} else {
			cookie, err := c.Cookie("session")
			if err == nil {
				userID, _ = strconv.Atoi(cookie)
			}
		}

		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}

		aid, _ := strconv.Atoi(activityID)

		// 开启事务
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
			return
		}

		// 加锁查询剩余名额
		var currentCount, maxParticipants int
		// 先获取活动的最大人数
		err = tx.QueryRow(`SELECT max_participants FROM activities WHERE id = $1 FOR UPDATE`, aid).Scan(&maxParticipants)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "活动不存在"})
			return
		}

		// 再计算当前报名人数
		err = tx.QueryRow(`SELECT COUNT(*) FROM signups WHERE activity_id = $1 AND status = 1`, aid).Scan(&currentCount)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询名额失败"})
			return
		}

		if currentCount >= maxParticipants {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "报名已满"})
			return
		}

		// 检查是否已报名
		var exists bool
		tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM signups WHERE activity_id = $1 AND user_id = $2 AND status = 1)`, aid, userID).Scan(&exists)
		if exists {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "您已报名"})
			return
		}

		// 插入报名记录
		_, err = tx.Exec(`
			INSERT INTO signups (activity_id, user_id, name, phone, emergency_contact, emergency_phone, remark, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 1)
		`, aid, userID, input.Name, input.Phone, input.EmergencyContact, input.EmergencyPhone, input.Remark)

		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "报名失败：" + err.Error()})
			return
		}

		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"message": "报名成功"})
	}
}

// GetSignups 获取活动报名列表
func GetSignups(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		activityID := c.Param("id")

		rows, err := db.Query(`
			SELECT s.id, s.user_id, s.name, s.phone, s.emergency_contact, s.emergency_phone, s.remark, s.status, s.created_at,
			       u.nickname, u.avatar
			FROM signups s
			LEFT JOIN users u ON s.user_id = u.id
			WHERE s.activity_id = $1 AND s.status = 1
			ORDER BY s.created_at DESC
		`, activityID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		type SignupItem struct {
			ID               int       `json:"id"`
			UserID           int       `json:"userId"`
			Name             string    `json:"name"`
			Phone            string    `json:"phone"`
			EmergencyContact string    `json:"emergencyContact"`
			EmergencyPhone   string    `json:"emergencyPhone"`
			Remark           string    `json:"remark"`
			Status           int       `json:"status"`
			CreatedAt        time.Time `json:"createdAt"`
			Nickname         string    `json:"nickname"`
			Avatar           string    `json:"avatar"`
		}

		var signups []SignupItem
		for rows.Next() {
			var s SignupItem
			err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Phone, &s.EmergencyContact,
				&s.EmergencyPhone, &s.Remark, &s.Status, &s.CreatedAt, &s.Nickname, &s.Avatar)
			if err != nil {
				continue
			}
			signups = append(signups, s)
		}

		c.JSON(http.StatusOK, gin.H{"data": signups})
	}
}

// GetMySignups 获取我的报名
func GetMySignups(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserIDFromContext(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}

		rows, err := db.Query(`
			SELECT a.id, a.title, a.date, a.location, a.price, a.status, a.cover,
			       s.name, s.phone, s.emergency_contact, s.emergency_phone, s.remark, s.status as signup_status, s.created_at
			FROM signups s
			JOIN activities a ON s.activity_id = a.id
			WHERE s.user_id = $1
			ORDER BY s.created_at DESC
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		type MySignup struct {
			ActivityID       int       `json:"activityId"`
			Title            string    `json:"title"`
			Date             string    `json:"date"`
			Location         string    `json:"location"`
			Price            float64   `json:"price"`
			ActivityStatus   int       `json:"activityStatus"`
			Cover            string    `json:"cover"`
			Name             string    `json:"name"`
			Phone            string    `json:"phone"`
			EmergencyContact string    `json:"emergencyContact"`
			EmergencyPhone   string    `json:"emergencyPhone"`
			Remark           string    `json:"remark"`
			SignupStatus     int       `json:"signupStatus"`
			CreatedAt        time.Time `json:"createdAt"`
		}

		var signups []MySignup
		for rows.Next() {
			var s MySignup
			err := rows.Scan(&s.ActivityID, &s.Title, &s.Date, &s.Location, &s.Price,
				&s.ActivityStatus, &s.Cover, &s.Name, &s.Phone, &s.EmergencyContact,
				&s.EmergencyPhone, &s.Remark, &s.SignupStatus, &s.CreatedAt)
			if err != nil {
				continue
			}
			signups = append(signups, s)
		}

		c.JSON(http.StatusOK, gin.H{"data": signups})
	}
}

// unused import error fix
var _ = errors.New("")
