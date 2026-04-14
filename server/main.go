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

	// 加载模板
	templatePath, _ := filepath.Abs("templates")
	r.LoadHTMLGlob(filepath.Join(templatePath, "*.html"))

	// 跨域配置 (仅 API)
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

	// 后台管理路由
	admin := r.Group("/admin")
	{
		admin.GET("/login", handlers.LoginPage)
		admin.POST("/login", handlers.LoginPost)
		admin.GET("/logout", handlers.Logout)

		// 需要登录验证的路由
		admin.Use(handlers.AuthMiddleware())
		{
			admin.GET("/", handlers.AdminIndex)
			admin.GET("/activities", handlers.AdminActivities(db))
			admin.GET("/activities/new", handlers.AdminNewActivity)
			admin.POST("/activities", handlers.AdminCreateActivity(db))
			admin.GET("/activities/:id/edit", handlers.AdminEditActivity(db))
			admin.POST("/activities/:id", handlers.AdminUpdateActivity(db))
			admin.DELETE("/activities/:id", handlers.AdminDeleteActivity(db))
			admin.GET("/signups/:id", handlers.AdminSignups(db))

			// 管理后台 JSON API（与小程序 API 分离）
			admin.GET("/api/activities", handlers.AdminListActivities(db))
			admin.GET("/api/activities/:id", handlers.AdminGetActivity(db))
			admin.POST("/api/activities", handlers.AdminCreateActivityJSON(db))
			admin.PUT("/api/activities/:id", handlers.AdminUpdateActivityJSON(db))
			admin.DELETE("/api/activities/:id", handlers.AdminDeleteActivityJSON(db))
			admin.GET("/api/activities/:id/signups", handlers.AdminGetSignups(db))
			admin.GET("/api/stats", handlers.AdminGetStats(db))
		}
	}

	// API 路由 (小程序使用)
	api := r.Group("/api")
	{
		// 认证相关
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

		// 活动相关
		api.GET("/activities", handlers.GetActivities(db))
		api.GET("/activities/:id", handlers.GetActivity(db))
		api.POST("/activities", handlers.CreateActivity(db))
		api.POST("/upload", handlers.UploadImage)

		// 报名相关
		api.POST("/activities/:id/signup", handlers.Signup(db))
		api.GET("/activities/:id/signups", handlers.GetSignups(db))

		// 我的报名
		api.GET("/my/signups", handlers.GetMySignups(db))
	}

	log.Println("服务器启动：http://localhost:8080")
	log.Println("后台管理：http://localhost:8080/admin/login")
	log.Println("API 文档：http://localhost:8080/api/activities")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
