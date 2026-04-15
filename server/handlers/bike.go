package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ============ 品牌相关 API ============

// GetBrands 获取品牌列表
func GetBrands(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT b.id, b.name, b.logo, COALESCE(b.description, ''),
			       COUNT(bm.id) as model_count
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

// GetBrand 获取品牌详情（含车型列表）
func GetBrand(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var brand struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Logo        string `json:"logo"`
			Description string `json:"description"`
		}
		err := db.QueryRow(`
			SELECT id, name, logo, COALESCE(description, '')
			FROM brands WHERE id = $1
		`, id).Scan(&brand.ID, &brand.Name, &brand.Logo, &brand.Description)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "品牌不存在"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 获取车型列表
		rows, err := db.Query(`
			SELECT id, name, price, cover, bike_type
			FROM brand_models WHERE brand_id = $1
			ORDER BY price DESC
		`, id)
		if err == nil {
			defer rows.Close()
			type ModelItem struct {
				ID      int     `json:"id"`
				Name    string  `json:"name"`
				Price   float64 `json:"price"`
				Cover   string  `json:"cover"`
				BikeType string `json:"bikeType"`
			}
			var models []ModelItem
			for rows.Next() {
				var m ModelItem
				if err := rows.Scan(&m.ID, &m.Name, &m.Price, &m.Cover, &m.BikeType); err != nil {
					continue
				}
				models = append(models, m)
			}
			c.JSON(http.StatusOK, gin.H{"data": gin.H{
				"brand":  brand,
				"models": models,
			}})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": gin.H{
			"brand":  brand,
			"models": []interface{}{},
		}})
	}
}

// GetBrandModels 获取某品牌下的车型
func GetBrandModels(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		brandID := c.Param("id")

		rows, err := db.Query(`
			SELECT id, name, price, cover, bike_type
			FROM brand_models WHERE brand_id = $1
			ORDER BY price DESC
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

// ============ 用户自行车相关 API ============

// getUserIDFromContext 从 context 获取用户ID
func getBikeUserID(c *gin.Context, db *sql.DB) int {
	userID := 0
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && len(authHeader) > 7 {
		userID, _ = strconv.Atoi(authHeader[7:])
	}
	if userID == 0 {
		cookie, _ := c.Cookie("session")
		if cookie != "" {
			userID, _ = strconv.Atoi(cookie)
		}
	}
	return userID
}

// GetMyBikes 获取我的自行车
func GetMyBikes(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getBikeUserID(c, db)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}

		rows, err := db.Query(`
			SELECT b.id, b.name, b.cover, b.bike_type, b.purchase_date, b.cost,
			       COALESCE(br.id, 0), COALESCE(br.name, ''), COALESCE(br.logo, ''),
			       COALESCE(bm.id, 0), COALESCE(bm.name, ''), COALESCE(bm.price, 0)
			FROM bikes b
			LEFT JOIN brands br ON b.brand_id = br.id
			LEFT JOIN brand_models bm ON b.model_id = bm.id
			WHERE b.user_id = $1
			ORDER BY b.created_at DESC
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		type BikeItem struct {
			ID            int      `json:"id"`
			Name          string   `json:"name"`
			Cover         string   `json:"cover"`
			BikeType      string   `json:"bikeType"`
			PurchaseDate  *string  `json:"purchaseDate,omitempty"`
			Cost          float64  `json:"cost"`
			BrandID       int      `json:"brandId"`
			BrandName     string   `json:"brandName"`
			BrandLogo     string   `json:"brandLogo"`
			ModelID       int      `json:"modelId"`
			ModelName     string   `json:"modelName"`
			ModelPrice    float64  `json:"modelPrice"`
		}

		var bikes []BikeItem
		for rows.Next() {
			var b BikeItem
			var purchaseDate sql.NullString
			if err := rows.Scan(&b.ID, &b.Name, &b.Cover, &b.BikeType, &purchaseDate, &b.Cost,
				&b.BrandID, &b.BrandName, &b.BrandLogo, &b.ModelID, &b.ModelName, &b.ModelPrice); err != nil {
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

// CreateBike 添加自行车
func CreateBike(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getBikeUserID(c, db)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}

		var input struct {
			Name         string  `json:"name"`
			Cover        string  `json:"cover"`
			BikeType     string  `json:"bikeType"`
			BrandID      *int    `json:"brandId"`
			ModelID      *int    `json:"modelId"`
			PurchaseDate *string `json:"purchaseDate"`
			Cost         float64 `json:"cost"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		brandID := 0
		modelID := 0
		if input.BrandID != nil {
			brandID = *input.BrandID
		}
		if input.ModelID != nil {
			modelID = *input.ModelID
		}

		var id int
		err := db.QueryRow(`
			INSERT INTO bikes (user_id, brand_id, model_id, name, cover, bike_type, purchase_date, cost)
			VALUES ($1, NULLIF($2, 0), NULLIF($3, 0), $4, $5, $6, $7, $8)
			RETURNING id
		`, userID, brandID, modelID, input.Name, input.Cover, input.BikeType, input.PurchaseDate, input.Cost).Scan(&id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id}})
	}
}

// UpdateBike 更新自行车
func UpdateBike(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getBikeUserID(c, db)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}

		id := c.Param("id")

		var input struct {
			Name         string  `json:"name"`
			Cover        string  `json:"cover"`
			BikeType     string  `json:"bikeType"`
			BrandID      *int    `json:"brandId"`
			ModelID      *int    `json:"modelId"`
			PurchaseDate *string `json:"purchaseDate"`
			Cost         float64 `json:"cost"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 验证所属权
		var ownerID int
		err := db.QueryRow(`SELECT user_id FROM bikes WHERE id = $1`, id).Scan(&ownerID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "自行车不存在"})
			return
		}
		if ownerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
			return
		}

		brandID := 0
		modelID := 0
		if input.BrandID != nil {
			brandID = *input.BrandID
		}
		if input.ModelID != nil {
			modelID = *input.ModelID
		}

		_, err = db.Exec(`
			UPDATE bikes SET name=$1, cover=$2, bike_type=$3,
				brand_id=NULLIF($4, 0), model_id=NULLIF($5, 0),
				purchase_date=$6, cost=$7
			WHERE id = $8
		`, input.Name, input.Cover, input.BikeType, brandID, modelID, input.PurchaseDate, input.Cost, id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
	}
}

// DeleteBike 删除自行车
func DeleteBike(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getBikeUserID(c, db)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}

		id := c.Param("id")

		// 验证所属权
		var ownerID int
		err := db.QueryRow(`SELECT user_id FROM bikes WHERE id = $1`, id).Scan(&ownerID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "自行车不存在"})
			return
		}
		if ownerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
			return
		}

		_, err = db.Exec(`DELETE FROM bikes WHERE id = $1`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
	}
}
