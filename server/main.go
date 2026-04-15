package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"yak-server/handlers"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// 连接数据库
	connStr := "host=localhost port=5432 user=yak password=yak123456 dbname=yak sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	log.Println("数据库连接成功")

	// 初始化 MinIO
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "localhost:9000"
	}
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	if minioAccessKey == "" {
		minioAccessKey = "yak"
	}
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		minioSecretKey = "yak123456"
	}
	minioBucket := os.Getenv("MINIO_BUCKET")
	if minioBucket == "" {
		minioBucket = "yak-uploads"
	}
	if err := handlers.InitMinIO(minioEndpoint, minioAccessKey, minioSecretKey, minioBucket); err != nil {
		log.Printf("警告: MinIO 初始化失败: %v，图片上传功能将不可用\n", err)
	} else {
		log.Println("MinIO 初始化成功")
	}

	// 设置 Gin 模式
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	// 跨域配置
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Vue SPA 静态文件路径
	adminFrontendPath, _ := filepath.Abs("admin-frontend/dist")

	// 后台管理 Vue SPA 路由（必须放在 API 路由前面）
	r.GET("/admin/*path", func(c *gin.Context) {
		indexPath := filepath.Join(adminFrontendPath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
		} else {
			c.String(http.StatusNotFound, "Vue 前端未构建，请先运行: cd admin-frontend && npm install && npm run build")
		}
	})

	// 后台管理 JSON API
	admin := r.Group("/api/admin")
	{
		admin.POST("/login", handlers.LoginAPI) // 登录接口不需要认证
	}

	// 需要认证的 API 路由
	adminAuth := r.Group("/api/admin")
	adminAuth.Use(handlers.AuthMiddleware())
	{
		adminAuth.GET("/stats", handlers.AdminGetStats(db))
		adminAuth.GET("/activities", handlers.AdminListActivities(db))
		adminAuth.GET("/activities/:id", handlers.AdminGetActivity(db))
		adminAuth.POST("/activities", handlers.AdminCreateActivityJSON(db))
		adminAuth.PUT("/activities/:id", handlers.AdminUpdateActivityJSON(db))
		adminAuth.DELETE("/activities/:id", handlers.AdminDeleteActivityJSON(db))
		adminAuth.GET("/activities/:id/signups", handlers.AdminGetSignups(db))

		adminAuth.GET("/brands", handlers.AdminListBrands(db))
		adminAuth.POST("/brands", handlers.AdminCreateBrand(db))
		adminAuth.PUT("/brands/:id", handlers.AdminUpdateBrand(db))
		adminAuth.DELETE("/brands/:id", handlers.AdminDeleteBrand(db))
		adminAuth.GET("/brands/:id/models", handlers.AdminGetBrandModels(db))

		adminAuth.POST("/models", handlers.AdminCreateModel(db))
		adminAuth.PUT("/models/:id", handlers.AdminUpdateModel(db))
		adminAuth.DELETE("/models/:id", handlers.AdminDeleteModel(db))

		adminAuth.GET("/bikes", handlers.AdminListBikes(db))
	}

	// 小程序 API 路由
	api := r.Group("/api")
	{
		api.POST("/auth/login", handlers.Login(db))
		api.GET("/auth/userinfo", handlers.GetCurrentUser(db), func(c *gin.Context) {
			if user, exists := c.Get("user"); exists {
				c.JSON(http.StatusOK, gin.H{"user": user})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			}
		})
		api.GET("/auth/rz_status", handlers.GetRzStatus(db))
		api.POST("/auth/rz", handlers.SubmitRz(db))
		api.POST("/auth/decrypt_phone", handlers.DecryptPhone(db))

		api.GET("/activities", handlers.GetActivities(db))
		api.GET("/activities/:id", handlers.GetActivity(db))
		api.POST("/activities", handlers.CreateActivity(db))
		api.POST("/upload", handlers.UploadImage)

		api.POST("/activities/:id/signup", handlers.Signup(db))
		api.GET("/activities/:id/signups", handlers.GetSignups(db))

		api.GET("/my/signups", handlers.GetMySignups(db))

		api.GET("/brands", handlers.GetBrands(db))
		api.GET("/brands/:id", handlers.GetBrand(db))
		api.GET("/brands/:id/models", handlers.GetBrandModels(db))

		api.GET("/bikes", handlers.GetMyBikes(db))
		api.POST("/bikes", handlers.CreateBike(db))
		api.PUT("/bikes/:id", handlers.UpdateBike(db))
		api.DELETE("/bikes/:id", handlers.DeleteBike(db))
	}

	log.Println("服务器启动：http://localhost:8080")
	log.Println("后台管理：http://localhost:8080/admin")
	log.Println("API 文档：http://localhost:8080/api/activities")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
