package upload

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wencai/easyhr/internal/common/response"
)

// Handler 图片上传处理器
type Handler struct {
	uploadDir string
	baseURL   string // e.g. https://your-bucket.oss-cn-beijing.aliyuncs.com
}

// NewHandler 创建 Handler
func NewHandler(uploadDir, baseURL string) *Handler {
	return &Handler{uploadDir: uploadDir, baseURL: baseURL}
}

// UploadImage handles POST /upload/image
// Accepts multipart/form-data with field name "file"
// Returns the uploaded file URL
func (h *Handler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "missing file")
		return
	}

	// 限制文件大小 5MB
	if file.Size > 5*1024*1024 {
		response.Error(c, http.StatusBadRequest, 40001, "文件大小不能超过5MB")
		return
	}

	// 验证文件类型
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		response.Error(c, http.StatusBadRequest, 40002, "仅支持 jpg/jpeg/png/gif/webp 格式")
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String()[:8], time.Now().Unix(), ext)
	dst := filepath.Join(h.uploadDir, filename)

	// 确保目录存在
	os.MkdirAll(h.uploadDir, 0755)

	// 保存文件
	src, err := file.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "open file failed")
		return
	}
	defer src.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "create file failed")
		return
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, src); err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "save file failed")
		return
	}

	// 返回 URL（实际生产中应替换为 OSS 签名 URL）
	url := fmt.Sprintf("%s/%s", h.baseURL, filename)
	response.Success(c, gin.H{"url": url})
}
