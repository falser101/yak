package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetRzStatus 获取当前用户实名认证状态
func GetRzStatus(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserIDFromContext(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}

		var rzStatus sql.NullInt64
		var phone, rzRealName, rzEmergencyName, rzEmergencyPhone sql.NullString
		var rzGender sql.NullInt64

		err := db.QueryRow(`
			SELECT COALESCE(rz_status, 0), COALESCE(phone, ''),
			       COALESCE(rz_real_name, ''), COALESCE(rz_gender, 0),
			       COALESCE(rz_emergency_name, ''), COALESCE(rz_emergency_phone, '')
			FROM users WHERE id = $1
		`, userID).Scan(&rzStatus, &phone, &rzRealName, &rzGender, &rzEmergencyName, &rzEmergencyPhone)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"rzStatus":         int(rzStatus.Int64),
				"phone":            phone.String,
				"rzRealName":       rzRealName.String,
				"rzGender":         int(rzGender.Int64),
				"rzEmergencyName":  rzEmergencyName.String,
				"rzEmergencyPhone": rzEmergencyPhone.String,
			},
		})
	}
}

// SubmitRz 提交实名认证
func SubmitRz(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserIDFromContext(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}

		var input struct {
			RealName       string `json:"realName" binding:"required"`
			IdCard        string `json:"idCard" binding:"required"`
			Gender        int    `json:"gender" binding:"required"` // 1:男, 2:女
			EmergencyName  string `json:"emergencyName" binding:"required"`
			EmergencyPhone string `json:"emergencyPhone" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 简单验证证件号格式（18位）
		if len(input.IdCard) != 18 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "证件号格式不正确"})
			return
		}

		// 更新用户实名信息，状态设为2（已认证）
		_, err := db.Exec(`
			UPDATE users
			SET rz_status = 2,
			    rz_real_name = $1,
			    rz_id_card = $2,
			    rz_gender = $3,
			    rz_emergency_name = $4,
			    rz_emergency_phone = $5,
			    rz_verified_at = CURRENT_TIMESTAMP
			WHERE id = $6
		`, input.RealName, input.IdCard, input.Gender, input.EmergencyName, input.EmergencyPhone, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "实名认证成功",
			"data": gin.H{
				"rzStatus":         2,
				"rzRealName":       input.RealName,
				"rzGender":         input.Gender,
				"rzEmergencyName":  input.EmergencyName,
				"rzEmergencyPhone": input.EmergencyPhone,
			},
		})
	}
}

// getUserIDFromContext 从上下文获取用户ID
func getUserIDFromContext(c *gin.Context) int {
	userID := 0
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && len(authHeader) > 7 {
		userID, _ = strconv.Atoi(authHeader[7:])
	} else {
		cookie, _ := c.Cookie("session")
		if cookie != "" {
			userID, _ = strconv.Atoi(cookie)
		}
	}
	return userID
}

// DecryptPhone 解密微信手机号
func DecryptPhone(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserIDFromContext(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}

		var input struct {
			EncryptData string `json:"encryptData"`
			IV          string `json:"iv"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}

		sessionKey, ok := sessionKeyStore[userID]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session已过期，请重新登录"})
			return
		}

		phone, err := decryptWechatPhone(input.EncryptData, input.IV, sessionKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "解密失败"})
			return
		}

		// 更新数据库中的 phone
		db.Exec("UPDATE users SET phone = $1 WHERE id = $2", phone, userID)

		c.JSON(http.StatusOK, gin.H{"phone": phone})
	}
}

// decryptWechatPhone 解密微信手机号
func decryptWechatPhone(encryptedDataStr, ivStr, sessionKeyStr string) (string, error) {
	// Base64 decode
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedDataStr)
	if err != nil {
		return "", errors.New("encryptedData base64 decode error")
	}
	iv, err := base64.StdEncoding.DecodeString(ivStr)
	if err != nil {
		return "", errors.New("iv base64 decode error")
	}
	sessionKey, err := base64.StdEncoding.DecodeString(sessionKeyStr)
	if err != nil {
		return "", errors.New("sessionKey base64 decode error")
	}

	// AES-256-CBC decryption
	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return "", errors.New("aes cipher error")
	}
	blockSize := block.BlockSize()
	if len(encryptedData) < blockSize {
		return "", errors.New("ciphertext too short")
	}

	// CBC mode
	if len(encryptedData)%blockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encryptedData, encryptedData)

	// PKCS7 unpad
	encryptedData = pkcs7Unpad(encryptedData, blockSize)
	if encryptedData == nil {
		return "", errors.New("pkcs7 unpad error")
	}

	// Parse JSON
	var result struct {
		PhoneNumber string `json:"phoneNumber"`
	}
	if err := json.Unmarshal(encryptedData, &result); err != nil {
		return "", errors.New("json parse error: " + err.Error())
	}

	return result.PhoneNumber, nil
}

// pkcs7Unpad removes PKCS7 padding
func pkcs7Unpad(data []byte, blockSize int) []byte {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil
	}
	pad := int(data[len(data)-1])
	if pad <= 0 || pad > blockSize {
		return nil
	}
	for i := len(data) - pad; i < len(data); i++ {
		if i < 0 || data[i] != byte(pad) {
			return nil
		}
	}
	return data[:len(data)-pad]
}
