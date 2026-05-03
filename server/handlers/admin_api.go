package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminStats 统计数据
type AdminStats struct {
	TotalActivities  int     `json:"totalActivities"`
	ActiveActivities int     `json:"activeActivities"`
	TotalSignups     int     `json:"totalSignups"`
	PendingSignups   int     `json:"pendingSignups"`
	TotalUsers       int     `json:"totalUsers"`
	ActiveUsers      int     `json:"activeUsers"`
	TotalRevenue     float64 `json:"totalRevenue"`
}

// AdminListActivities 获取活动列表（管理后台用）
func AdminListActivities(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
		category := c.DefaultQuery("category", "")

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		offset := (page - 1) * pageSize

		// 构建查询条件
		whereClause := ""
		args := []interface{}{pageSize, offset}
		if category != "" {
			whereClause = " WHERE a.category = $3"
			args = []interface{}{pageSize, offset, category}
		}

		// 获取总数
		var total int
		countQuery := `SELECT COUNT(*) FROM activities a` + whereClause
		if category != "" {
			db.QueryRow(countQuery, category).Scan(&total)
		} else {
			db.QueryRow(countQuery).Scan(&total)
		}

		// 获取活动列表
		query := `
			SELECT
				a.id, a.title, a.cover, a.date, a.location,
				a.max_participants, a.price, a.description, a.status, a.created_at, a.signup_end_time,
				COALESCE(a.category, 'activity'), COALESCE(a.rules, ''), COALESCE(a.route, ''), COALESCE(a.awards, ''),
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
			` + whereClause + `
			ORDER BY a.date DESC
			LIMIT $1 OFFSET $2
		`
		rows, err := db.Query(query, args...)

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
			Category        string     `json:"category"`
			Rules           string     `json:"rules"`
			Route           string     `json:"route"`
			Awards          string     `json:"awards"`
		}

		var activities []ActivityItem
		for rows.Next() {
			var a ActivityItem
			var signupEndTime sql.NullString
			err := rows.Scan(&a.ID, &a.Title, &a.Cover, &a.Date, &a.Location,
				&a.MaxParticipants, &a.Price, &a.Description, &a.Status, &a.CreatedAt, &signupEndTime,
				&a.Category, &a.Rules, &a.Route, &a.Awards,
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
			Category        string     `json:"category"`
			Rules           string     `json:"rules"`
			Route           string     `json:"route"`
			Awards          string     `json:"awards"`
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
				), 0),
				COALESCE(a.category, 'activity'), COALESCE(a.rules, ''), COALESCE(a.route, ''), COALESCE(a.awards, '')
			FROM activities a
			LEFT JOIN users u ON CASE WHEN a.created_by ~ '^[0-9]+$' THEN a.created_by::int ELSE 0 END = u.id
			WHERE a.id = $1
		`, id).Scan(
			&a.ID, &a.Title, &a.Cover, &a.Date, &a.Location,
			&a.MaxParticipants, &a.Price, &a.Description, &a.Status, &a.CreatedAt, &signupEndTime,
			&a.CreatedByID, &a.CreatedByName, &a.SignupCount,
			&a.Category, &a.Rules, &a.Route, &a.Awards)

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
			Category        string   `json:"category"`
			Rules           string   `json:"rules"`
			Route           string   `json:"route"`
			Awards          string   `json:"awards"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 默认分类
		category := input.Category
		if category == "" {
			category = "activity"
		}

		var id int
		err := db.QueryRow(`
			INSERT INTO activities (title, cover, date, location, max_participants, price, description, signup_end_time, created_by, category, rules, route, awards)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			RETURNING id
		`, input.Title, input.Cover, input.Date, input.Location,
			input.MaxParticipants, input.Price, input.Description, input.SignupEndTime, 1,
			category, input.Rules, input.Route, input.Awards).Scan(&id)

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
			Category        string  `json:"category"`
			Rules           string  `json:"rules"`
			Route           string  `json:"route"`
			Awards          string  `json:"awards"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec(`
			UPDATE activities
			SET title=$1, cover=$2, date=$3, location=$4, max_participants=$5,
			    price=$6, description=$7, status=$8, signup_end_time=$9,
			    category=$10, rules=$11, route=$12, awards=$13
			WHERE id=$14
		`, input.Title, input.Cover, input.Date, input.Location,
			input.MaxParticipants, input.Price, input.Description, input.Status, input.SignupEndTime,
			input.Category, input.Rules, input.Route, input.Awards, id)

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
	IdNumber         string `json:"idNumber"`
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
			SELECT s.id, s.user_id, s.name, s.phone, COALESCE(s.id_number, ''), s.emergency_contact, s.emergency_phone,
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
			rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Phone, &s.IdNumber, &s.EmergencyContact, &s.EmergencyPhone,
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
		db.QueryRow(`SELECT COUNT(*) FROM activities WHERE status = 1`).Scan(&stats.ActiveActivities)
		db.QueryRow(`SELECT COUNT(*) FROM signups WHERE status = 1`).Scan(&stats.TotalSignups)
		db.QueryRow(`SELECT COUNT(*) FROM signups WHERE status = 0`).Scan(&stats.PendingSignups)
		db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
		db.QueryRow(`SELECT COUNT(*) FROM users WHERE COALESCE(status, 1) = 1`).Scan(&stats.ActiveUsers)
		db.QueryRow(`SELECT COALESCE(SUM(amount), 0) FROM signups WHERE status = 1`).Scan(&stats.TotalRevenue)

		c.JSON(http.StatusOK, gin.H{"data": stats})
	}
}

// ============ 品牌管理 API ============

// AdminListBrands 品牌列表
func AdminListBrands(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT b.id, b.name, b.logo, COALESCE(b.description, ''), COUNT(bm.id) as model_count
			FROM brands b
			LEFT JOIN brand_models bm ON b.id = bm.brand_id
			GROUP BY b.id, b.name, b.logo, b.description
			ORDER BY b.id
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		type BrandItem struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Logo        string `json:"logo"`
			Description string `json:"description"`
			ModelCount  int    `json:"modelCount"`
		}

		var brands []BrandItem
		for rows.Next() {
			var b BrandItem
			if err := rows.Scan(&b.ID, &b.Name, &b.Logo, &b.Description, &b.ModelCount); err != nil {
				continue
			}
			brands = append(brands, b)
		}

		c.JSON(http.StatusOK, gin.H{"data": brands})
	}
}

// AdminCreateBrand 创建品牌
func AdminCreateBrand(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Name        string `json:"name"`
			Logo        string `json:"logo"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var id int
		err := db.QueryRow(`
			INSERT INTO brands (name, logo, description) VALUES ($1, $2, $3) RETURNING id
		`, input.Name, input.Logo, input.Description).Scan(&id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id}})
	}
}

// AdminUpdateBrand 更新品牌
func AdminUpdateBrand(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var input struct {
			Name        string `json:"name"`
			Logo        string `json:"logo"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec(`
			UPDATE brands SET name=$1, logo=$2, description=$3 WHERE id=$4
		`, input.Name, input.Logo, input.Description, id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
	}
}

// AdminDeleteBrand 删除品牌
func AdminDeleteBrand(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		_, err := db.Exec(`DELETE FROM brands WHERE id = $1`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
	}
}

// AdminGetBrandModels 获取品牌车型
func AdminGetBrandModels(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		brandID := c.Param("id")

		rows, err := db.Query(`
			SELECT id, name, price, cover, bike_type FROM brand_models WHERE brand_id = $1 ORDER BY price DESC
		`, brandID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		type ModelItem struct {
			ID       int     `json:"id"`
			Name     string  `json:"name"`
			Price    float64 `json:"price"`
			Cover    string  `json:"cover"`
			BikeType string  `json:"bikeType"`
		}

		var models []ModelItem
		for rows.Next() {
			var m ModelItem
			if err := rows.Scan(&m.ID, &m.Name, &m.Price, &m.Cover, &m.BikeType); err != nil {
				continue
			}
			models = append(models, m)
		}

		c.JSON(http.StatusOK, gin.H{"data": models})
	}
}

// AdminCreateModel 创建车型
func AdminCreateModel(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			BrandID  int     `json:"brandId"`
			Name     string  `json:"name"`
			Price    float64 `json:"price"`
			Cover    string  `json:"cover"`
			BikeType string  `json:"bikeType"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var id int
		err := db.QueryRow(`
			INSERT INTO brand_models (brand_id, name, price, cover, bike_type) VALUES ($1, $2, $3, $4, $5) RETURNING id
		`, input.BrandID, input.Name, input.Price, input.Cover, input.BikeType).Scan(&id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id}})
	}
}

// AdminUpdateModel 更新车型
func AdminUpdateModel(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var input struct {
			BrandID  int     `json:"brandId"`
			Name     string  `json:"name"`
			Price    float64 `json:"price"`
			Cover    string  `json:"cover"`
			BikeType string  `json:"bikeType"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec(`
			UPDATE brand_models SET brand_id=$1, name=$2, price=$3, cover=$4, bike_type=$5 WHERE id=$6
		`, input.BrandID, input.Name, input.Price, input.Cover, input.BikeType, id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
	}
}

// AdminDeleteModel 删除车型
func AdminDeleteModel(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		_, err := db.Exec(`DELETE FROM brand_models WHERE id = $1`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
	}
}

// AdminListBikes 用户自行车列表
func AdminListBikes(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT b.id, b.name, b.cover, b.bike_type, b.purchase_date, b.cost,
			       COALESCE(u.nickname, '') as user_name,
			       COALESCE(br.name, ''), COALESCE(bm.name, '')
			FROM bikes b
			LEFT JOIN users u ON b.user_id = u.id
			LEFT JOIN brands br ON b.brand_id = br.id
			LEFT JOIN brand_models bm ON b.model_id = bm.id
			ORDER BY b.created_at DESC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		type BikeItem struct {
			ID          int      `json:"id"`
			Name        string   `json:"name"`
			Cover       string   `json:"cover"`
			BikeType    string   `json:"bikeType"`
			PurchaseDate *string `json:"purchaseDate"`
			Cost        float64  `json:"cost"`
			UserName    string   `json:"userName"`
			BrandName   string   `json:"brandName"`
			ModelName   string   `json:"modelName"`
		}

		var bikes []BikeItem
		for rows.Next() {
			var b BikeItem
			var purchaseDate sql.NullString
			if err := rows.Scan(&b.ID, &b.Name, &b.Cover, &b.BikeType, &purchaseDate, &b.Cost,
				&b.UserName, &b.BrandName, &b.ModelName); err != nil {
				continue
			}
			if purchaseDate.Valid {
				b.PurchaseDate = &purchaseDate.String
			}
			bikes = append(bikes, b)
		}

		c.JSON(http.StatusOK, gin.H{"data": bikes})
	}
}

// ============ 用户管理 API ============

// AdminUserItem 用户项
type AdminUserItem struct {
	ID              int     `json:"id"`
	Openid          string  `json:"openid"`
	Nickname        string  `json:"nickname"`
	Avatar          string  `json:"avatar"`
	Phone           string  `json:"phone"`
	MembershipLevel int     `json:"membershipLevel"`
	TotalRides      int     `json:"totalRides"`
	SignupCount     int     `json:"signupCount"`
	Status          int     `json:"status"`
	CreatedAt       string  `json:"createdAt"`
	LastLoginTime   *string `json:"lastLoginTime,omitempty"`
}

// AdminListUsers 获取用户列表
func AdminListUsers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
		keyword := c.DefaultQuery("keyword", "")
		membershipLevel := c.DefaultQuery("membershipLevel", "")

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		offset := (page - 1) * pageSize

		// 构建查询条件
		whereClause := "WHERE 1=1"
		args := []interface{}{}
		argIdx := 1

		if keyword != "" {
			whereClause += fmt.Sprintf(" AND (u.nickname LIKE $%d OR u.phone LIKE $%d)", argIdx, argIdx)
			args = append(args, "%"+keyword+"%")
			argIdx++
		}

		if membershipLevel != "" {
			whereClause += fmt.Sprintf(" AND COALESCE(u.membership_level, 0) = $%d", argIdx)
			args = append(args, membershipLevel)
			argIdx++
		}

		// 获取总数
		var total int
		countQuery := `SELECT COUNT(*) FROM users u ` + whereClause
		if len(args) > 0 {
			db.QueryRow(countQuery, args...).Scan(&total)
		} else {
			db.QueryRow(countQuery).Scan(&total)
		}

		// 获取用户列表
		query := fmt.Sprintf(`
			SELECT u.id, u.openid, COALESCE(u.nickname, ''), COALESCE(u.avatar, ''),
			       COALESCE(u.phone, ''), COALESCE(u.membership_level, 0), COALESCE(u.total_rides, 0),
			       COALESCE(u.status, 1), u.created_at,
			       (SELECT COUNT(*) FROM signups WHERE user_id = u.id AND status = 1) as signup_count
			FROM users u
			%s
			ORDER BY u.created_at DESC
			LIMIT $%d OFFSET $%d
		`, whereClause, argIdx, argIdx+1)
		args = append(args, pageSize, offset)

		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var users []AdminUserItem
		for rows.Next() {
			var u AdminUserItem
			err := rows.Scan(&u.ID, &u.Openid, &u.Nickname, &u.Avatar, &u.Phone,
				&u.MembershipLevel, &u.TotalRides, &u.Status, &u.CreatedAt, &u.SignupCount)
			if err != nil {
				continue
			}
			users = append(users, u)
		}

		c.JSON(http.StatusOK, gin.H{
			"data":     users,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		})
	}
}

// AdminGetUser 获取用户详情
func AdminGetUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var u AdminUserItem
		err := db.QueryRow(`
			SELECT u.id, u.openid, COALESCE(u.nickname, ''), COALESCE(u.avatar, ''),
			       COALESCE(u.phone, ''), COALESCE(u.membership_level, 0), COALESCE(u.total_rides, 0),
			       COALESCE(u.status, 1), u.created_at,
			       (SELECT COUNT(*) FROM signups WHERE user_id = u.id AND status = 1) as signup_count
			FROM users u
			WHERE u.id = $1
		`, id).Scan(&u.ID, &u.Openid, &u.Nickname, &u.Avatar, &u.Phone,
			&u.MembershipLevel, &u.TotalRides, &u.Status, &u.CreatedAt, &u.SignupCount)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": u})
	}
}

// AdminUpdateUser 更新用户
func AdminUpdateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var input struct {
			Nickname        string `json:"nickname"`
			MembershipLevel int    `json:"membershipLevel"`
			Status          int    `json:"status"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec(`
			UPDATE users
			SET nickname = $1, membership_level = $2, status = $3
			WHERE id = $4
		`, input.Nickname, input.MembershipLevel, input.Status, id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
	}
}

// AdminDisableUser 禁用/启用用户
func AdminDisableUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var input struct {
			Status int `json:"status"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec(`UPDATE users SET status = $1 WHERE id = $2`, input.Status, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "操作成功"})
	}
}
