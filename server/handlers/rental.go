package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ============ 租车车辆 API ============

// RentalBikeItem 租车车辆项
type RentalBikeItem struct {
	ID         int             `json:"id"`
	Name       string          `json:"name"`
	Cover      string          `json:"cover"`
	BikeType   string          `json:"bikeType"`
	Tag        string          `json:"tag"`
	PriceDay   float64         `json:"priceDay"`
	PriceHour  float64         `json:"priceHour"`
	PriceTeam  float64         `json:"priceTeam"`
	Deposit    float64         `json:"deposit"`
	Specs      map[string]any  `json:"specs"`
	Notes      string          `json:"notes"`
	Status     int             `json:"status"`
	BrandID    int            `json:"brandId"`
	BrandName  string         `json:"brandName"`
	BrandLogo  string         `json:"brandLogo"`
}

// GetRentalBikes 获取租车车辆列表
func GetRentalBikes(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		bikeType := c.DefaultQuery("bikeType", "")
		tag := c.DefaultQuery("tag", "")
		keyword := c.DefaultQuery("keyword", "")

		whereClause := "WHERE rb.status = 1"
		args := []interface{}{}
		argIdx := 1

		if bikeType != "" {
			whereClause += fmt.Sprintf(" AND rb.bike_type = $%d", argIdx)
			args = append(args, bikeType)
			argIdx++
		}
		if tag != "" {
			whereClause += fmt.Sprintf(" AND rb.tag = $%d", argIdx)
			args = append(args, tag)
			argIdx++
		}
		if keyword != "" {
			whereClause += fmt.Sprintf(" AND rb.name LIKE $%d", argIdx)
			args = append(args, "%"+keyword+"%")
			argIdx++
		}

		query := fmt.Sprintf(`
			SELECT rb.id, rb.name, rb.cover, rb.bike_type, rb.tag,
			       rb.price_day, rb.price_hour, rb.price_team, rb.deposit,
			       COALESCE(rb.specs, '{}'), COALESCE(rb.notes, ''), rb.status,
			       COALESCE(br.id, 0), COALESCE(br.name, ''), COALESCE(br.logo, '')
			FROM rental_bikes rb
			LEFT JOIN brands br ON rb.brand_id = br.id
			%s
			ORDER BY rb.id DESC
		`, whereClause)

		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var bikes []RentalBikeItem
		for rows.Next() {
			var b RentalBikeItem
			var specsJSON []byte
			var tag sql.NullString
			err := rows.Scan(&b.ID, &b.Name, &b.Cover, &b.BikeType, &tag,
				&b.PriceDay, &b.PriceHour, &b.PriceTeam, &b.Deposit,
				&specsJSON, &b.Notes, &b.Status,
				&b.BrandID, &b.BrandName, &b.BrandLogo)
			if err != nil {
				continue
			}
			b.Tag = tag.String
			// 解析 specs JSON
			if len(specsJSON) > 0 {
				json.Unmarshal(specsJSON, &b.Specs)
			}
			bikes = append(bikes, b)
		}

		c.JSON(http.StatusOK, gin.H{"data": bikes})
	}
}

// GetRentalBike 获取租车车辆详情
func GetRentalBike(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var b RentalBikeItem
		var specsJSON []byte
		var tag sql.NullString
		err := db.QueryRow(`
			SELECT rb.id, rb.name, rb.cover, rb.bike_type, rb.tag,
			       rb.price_day, rb.price_hour, rb.price_team, rb.deposit,
			       COALESCE(rb.specs, '{}'), COALESCE(rb.notes, ''), rb.status,
			       COALESCE(br.id, 0), COALESCE(br.name, ''), COALESCE(br.logo, '')
			FROM rental_bikes rb
			LEFT JOIN brands br ON rb.brand_id = br.id
			WHERE rb.id = $1
		`, id).Scan(&b.ID, &b.Name, &b.Cover, &b.BikeType, &tag,
			&b.PriceDay, &b.PriceHour, &b.PriceTeam, &b.Deposit,
			&specsJSON, &b.Notes, &b.Status,
			&b.BrandID, &b.BrandName, &b.BrandLogo)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "车辆不存在"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		b.Tag = tag.String
		if len(specsJSON) > 0 {
			json.Unmarshal(specsJSON, &b.Specs)
		}

		c.JSON(http.StatusOK, gin.H{"data": b})
	}
}

// ============ 租车订单 API ============

// RentalOrderItem 租车订单项
type RentalOrderItem struct {
	ID           int        `json:"id"`
	OrderNo     string     `json:"orderNo"`
	BikeID       int        `json:"bikeId"`
	BikeName     string     `json:"bikeName"`
	BikeCover    string     `json:"bikeCover"`
	BikeColor    string     `json:"bikeColor"`
	Package      string     `json:"package"`
	Quantity     int        `json:"quantity"`
	RentalDate   string     `json:"rentalDate"`
	Amount       float64    `json:"amount"`
	Deposit      float64    `json:"deposit"`
	Status       int        `json:"status"`
	ContactName  string     `json:"contactName"`
	ContactPhone string     `json:"contactPhone"`
	Remark       string     `json:"remark"`
	CreatedAt   string     `json:"createdAt"`
	PayTime      *string    `json:"payTime,omitempty"`
	PayMethod    int        `json:"payMethod"`
}

// generateOrderNo 生成订单号
func generateOrderNo() string {
	return fmt.Sprintf("RC%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
}

// CreateRentalOrder 创建租车订单
func CreateRentalOrder(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getBikeUserID(c, db)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}

		var input struct {
			BikeID      int     `json:"bikeId" binding:"required"`
			Package     string  `json:"package" binding:"required"`
			Quantity    int     `json:"quantity"`
			RentalDate  string  `json:"rentalDate"`
			ContactName string  `json:"contactName"`
			ContactPhone string `json:"contactPhone"`
			Remark      string  `json:"remark"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if input.Quantity < 1 {
			input.Quantity = 1
		}

		// 查询车辆信息
		var bikeName, bikeCover, bikeColor string
		var priceDay, priceHour, priceTeam, deposit float64
		err := db.QueryRow(`
			SELECT name, cover, bike_type, price_day, price_hour, price_team, deposit
			FROM rental_bikes WHERE id = $1 AND status = 1
		`, input.BikeID).Scan(&bikeName, &bikeCover, &bikeColor, &priceDay, &priceHour, &priceTeam, &deposit)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "车辆不存在或已下架"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 计算金额
		var amount float64
		switch input.Package {
		case "day":
			amount = priceDay * float64(input.Quantity)
		case "hour":
			amount = priceHour * float64(input.Quantity)
		case "team":
			amount = priceTeam
		default:
			amount = priceDay * float64(input.Quantity)
		}

		orderNo := generateOrderNo()

		// 处理空日期
		rentalDate := input.RentalDate
		if rentalDate == "" {
			rentalDate = "2099-12-31" // 默认日期
		}

		var id int
		err = db.QueryRow(`
			INSERT INTO rental_orders (order_no, user_id, bike_id, bike_name, bike_cover, bike_color,
				package, quantity, rental_date, amount, deposit, contact_name, contact_phone, remark, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, 0)
			RETURNING id
		`, orderNo, userID, input.BikeID, bikeName, bikeCover, bikeColor,
			input.Package, input.Quantity, rentalDate, amount, deposit,
			input.ContactName, input.ContactPhone, input.Remark).Scan(&id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": gin.H{
			"id":       id,
			"orderNo":  orderNo,
			"amount":   amount,
			"deposit":  deposit,
			"total":    amount + deposit,
		}})
	}
}

// GetMyRentalOrders 获取我的租车订单
func GetMyRentalOrders(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getBikeUserID(c, db)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}

		statusFilter := c.DefaultQuery("status", "")

		query := `
			SELECT id, order_no, bike_id, bike_name, bike_cover, bike_color,
			       package, quantity, rental_date, amount, deposit, status,
			       contact_name, contact_phone, remark, created_at, pay_time, pay_method
			FROM rental_orders WHERE user_id = $1
		`
		args := []interface{}{userID}
		if statusFilter != "" {
			s, _ := strconv.Atoi(statusFilter)
			query += " AND status = $2"
			args = append(args, s)
		}
		query += " ORDER BY created_at DESC"

		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var orders []RentalOrderItem
		for rows.Next() {
			var o RentalOrderItem
			var rentalDate, createdAt sql.NullString
			var payTime sql.NullString
			err := rows.Scan(&o.ID, &o.OrderNo, &o.BikeID, &o.BikeName, &o.BikeCover, &o.BikeColor,
				&o.Package, &o.Quantity, &rentalDate, &o.Amount, &o.Deposit, &o.Status,
				&o.ContactName, &o.ContactPhone, &o.Remark, &createdAt, &payTime, &o.PayMethod)
			if err != nil {
				continue
			}
			o.RentalDate = rentalDate.String
			o.CreatedAt = createdAt.String
			if payTime.Valid {
				o.PayTime = &payTime.String
			}
			orders = append(orders, o)
		}

		c.JSON(http.StatusOK, gin.H{"data": orders})
	}
}

// CancelRentalOrder 取消租车订单
func CancelRentalOrder(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getBikeUserID(c, db)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}

		id := c.Param("id")

		var ownerID int
		err := db.QueryRow(`SELECT user_id FROM rental_orders WHERE id = $1`, id).Scan(&ownerID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
			return
		}
		if ownerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
			return
		}

		_, err = db.Exec(`UPDATE rental_orders SET status = 2 WHERE id = $1 AND status = 0`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "取消成功"})
	}
}

// ============ 支付 API ============

// PayAtStore 到店支付
func PayAtStore(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getBikeUserID(c, db)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}

		var input struct {
			OrderType string `json:"orderType"` // "rental" 或 "signup"
			OrderID   int    `json:"orderId"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if input.OrderType == "rental" {
			// 租车订单 - 状态变为待支付（到店支付）
			var ownerID int
			err := db.QueryRow(`SELECT user_id FROM rental_orders WHERE id = $1`, input.OrderID).Scan(&ownerID)
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
				return
			}
			if ownerID != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
				return
			}

			_, err = db.Exec(`
				UPDATE rental_orders SET status = 0, pay_method = 3 WHERE id = $1
			`, input.OrderID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "已选择到店支付，请到店付款"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的订单类型"})
		}
	}
}

// ============ Admin 租车订单管理 ============

// AdminRentalOrderItem 后台租车订单项
type AdminRentalOrderItem struct {
	ID           int        `json:"id"`
	OrderNo     string     `json:"orderNo"`
	UserID       int        `json:"userId"`
	Nickname     string     `json:"nickname"`
	BikeID       int        `json:"bikeId"`
	BikeName     string     `json:"bikeName"`
	BikeCover    string     `json:"bikeCover"`
	Package      string     `json:"package"`
	Quantity     int        `json:"quantity"`
	RentalDate   string     `json:"rentalDate"`
	Amount       float64    `json:"amount"`
	Deposit      float64    `json:"deposit"`
	Status       int        `json:"status"`
	ContactName  string     `json:"contactName"`
	ContactPhone string     `json:"contactPhone"`
	Remark       string     `json:"remark"`
	CreatedAt   string     `json:"createdAt"`
	PayTime      *string   `json:"payTime,omitempty"`
	PayMethod    int        `json:"payMethod"`
}

// AdminListRentalOrders 获取租车订单列表
func AdminListRentalOrders(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
		status := c.DefaultQuery("status", "")

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		offset := (page - 1) * pageSize

		whereClause := "WHERE 1=1"
		args := []interface{}{}
		argIdx := 1

		if status != "" {
			whereClause += fmt.Sprintf(" AND ro.status = $%d", argIdx)
			args = append(args, status)
			argIdx++
		}

		var total int
		countQuery := `SELECT COUNT(*) FROM rental_orders ro ` + whereClause
		if len(args) > 0 {
			db.QueryRow(countQuery, args...).Scan(&total)
		} else {
			db.QueryRow(countQuery).Scan(&total)
		}

		query := fmt.Sprintf(`
			SELECT ro.id, ro.order_no, COALESCE(ro.user_id, 0), COALESCE(u.nickname, ''),
			       ro.bike_id, ro.bike_name, ro.bike_cover, ro.package, ro.quantity,
			       COALESCE(ro.rental_date::text, ''), ro.amount, ro.deposit, ro.status,
			       ro.contact_name, ro.contact_phone, COALESCE(ro.remark, ''),
			       ro.created_at::text, COALESCE(ro.pay_time::text, ''), ro.pay_method
			FROM rental_orders ro
			LEFT JOIN users u ON ro.user_id = u.id
			%s
			ORDER BY ro.created_at DESC
			LIMIT $%d OFFSET $%d
		`, whereClause, argIdx, argIdx+1)
		args = append(args, pageSize, offset)

		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var orders []AdminRentalOrderItem
		for rows.Next() {
			var o AdminRentalOrderItem
			var payTime sql.NullString
			err := rows.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.Nickname,
				&o.BikeID, &o.BikeName, &o.BikeCover, &o.Package, &o.Quantity,
				&o.RentalDate, &o.Amount, &o.Deposit, &o.Status,
				&o.ContactName, &o.ContactPhone, &o.Remark,
				&o.CreatedAt, &payTime, &o.PayMethod)
			if err != nil {
				continue
			}
			if payTime.Valid && payTime.String != "" {
				o.PayTime = &payTime.String
			}
			orders = append(orders, o)
		}

		c.JSON(http.StatusOK, gin.H{
			"data":     orders,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		})
	}
}

// AdminConfirmPayment 确认收款
func AdminConfirmPayment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var input struct {
			Status int `json:"status"` // 1=已支付
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 更新租车订单状态
		_, err := db.Exec(`
			UPDATE rental_orders
			SET status = $1, pay_time = CURRENT_TIMESTAMP
			WHERE id = $2
		`, input.Status, id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "操作成功"})
	}
}

// AdminRentalStats 租车统计
func AdminRentalStats(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var stats struct {
			TotalOrders   int     `json:"totalOrders"`
			PendingOrders int     `json:"pendingOrders"`
			TotalRevenue  float64 `json:"totalRevenue"`
		}

		db.QueryRow(`SELECT COUNT(*) FROM rental_orders`).Scan(&stats.TotalOrders)
		db.QueryRow(`SELECT COUNT(*) FROM rental_orders WHERE status = 0`).Scan(&stats.PendingOrders)
		db.QueryRow(`SELECT COALESCE(SUM(amount + deposit), 0) FROM rental_orders WHERE status = 1`).Scan(&stats.TotalRevenue)

		c.JSON(http.StatusOK, gin.H{"data": stats})
	}
}

// AdminRentalBikes 后台租车车辆列表
func AdminListRentalBikes(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT rb.id, rb.name, rb.cover, rb.bike_type, rb.tag,
			       rb.price_day, rb.price_hour, rb.price_team, rb.deposit,
			       COALESCE(rb.specs, '{}'), COALESCE(rb.notes, ''), rb.status,
			       COALESCE(br.id, 0), COALESCE(br.name, '')
			FROM rental_bikes rb
			LEFT JOIN brands br ON rb.brand_id = br.id
			ORDER BY rb.id DESC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var bikes []RentalBikeItem
		for rows.Next() {
			var b RentalBikeItem
			var specsJSON []byte
			var tag sql.NullString
			err := rows.Scan(&b.ID, &b.Name, &b.Cover, &b.BikeType, &tag,
				&b.PriceDay, &b.PriceHour, &b.PriceTeam, &b.Deposit,
				&specsJSON, &b.Notes, &b.Status,
				&b.BrandID, &b.BrandName)
			if err != nil {
				continue
			}
			b.Tag = tag.String
			if len(specsJSON) > 0 {
				json.Unmarshal(specsJSON, &b.Specs)
			}
			bikes = append(bikes, b)
		}

		c.JSON(http.StatusOK, gin.H{"data": bikes})
	}
}

// AdminCreateRentalBike 创建租车车辆
func AdminCreateRentalBike(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			BrandID   int             `json:"brandId"`
			ModelID   int             `json:"modelId"`
			Name      string          `json:"name"`
			Cover     string          `json:"cover"`
			BikeType  string          `json:"bikeType"`
			Tag       string          `json:"tag"`
			PriceDay  float64         `json:"priceDay"`
			PriceHour float64         `json:"priceHour"`
			PriceTeam float64         `json:"priceTeam"`
			Deposit   float64         `json:"deposit"`
			Specs     map[string]any  `json:"specs"`
			Notes     string          `json:"notes"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		specsJSON, _ := json.Marshal(input.Specs)

		var id int
		err := db.QueryRow(`
			INSERT INTO rental_bikes (brand_id, model_id, name, cover, bike_type, tag,
				price_day, price_hour, price_team, deposit, specs, notes, status)
			VALUES ($1, NULLIF($2, 0), $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 1)
			RETURNING id
		`, input.BrandID, input.ModelID, input.Name, input.Cover, input.BikeType,
			input.Tag, input.PriceDay, input.PriceHour, input.PriceTeam,
			input.Deposit, specsJSON, input.Notes).Scan(&id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id}})
	}
}

// AdminUpdateRentalBike 更新租车车辆
func AdminUpdateRentalBike(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var input struct {
			BrandID   int             `json:"brandId"`
			ModelID   int             `json:"modelId"`
			Name      string          `json:"name"`
			Cover     string          `json:"cover"`
			BikeType  string          `json:"bikeType"`
			Tag       string          `json:"tag"`
			PriceDay  float64         `json:"priceDay"`
			PriceHour float64         `json:"priceHour"`
			PriceTeam float64         `json:"priceTeam"`
			Deposit   float64         `json:"deposit"`
			Specs     map[string]any  `json:"specs"`
			Notes     string          `json:"notes"`
			Status    int             `json:"status"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		specsJSON, _ := json.Marshal(input.Specs)

		_, err := db.Exec(`
			UPDATE rental_bikes SET brand_id=$1, model_id=$2, name=$3, cover=$4,
				bike_type=$5, tag=$6, price_day=$7, price_hour=$8, price_team=$9,
				deposit=$10, specs=$11, notes=$12, status=$13
			WHERE id=$14
		`, input.BrandID, input.ModelID, input.Name, input.Cover, input.BikeType,
			input.Tag, input.PriceDay, input.PriceHour, input.PriceTeam,
			input.Deposit, specsJSON, input.Notes, input.Status, id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
	}
}

// AdminDeleteRentalBike 删除租车车辆
func AdminDeleteRentalBike(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		_, err := db.Exec(`DELETE FROM rental_bikes WHERE id = $1`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
	}
}
