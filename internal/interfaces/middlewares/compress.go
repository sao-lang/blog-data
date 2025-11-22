package middlewares

import (
	"compress/gzip"
	"io"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// gzipWriterWrapper 是一个实现 io.Writer 接口的结构体，
// 用于将数据写入 gzip.Writer，并将压缩后的数据写入 ResponseWriter。
type gzipWriterWrapper struct {
	gin.ResponseWriter
	gzipWriter *gzip.Writer
}

// Write 实现了 io.Writer 接口的 Write 方法，
// 将数据写入 gzip.Writer，并将压缩后的数据写入 ResponseWriter。
func (w *gzipWriterWrapper) Write(data []byte) (int, error) {
	return w.gzipWriter.Write(data)
}

// WriteString 实现了 io.StringWriter 接口的 WriteString 方法，
// 将字符串写入 gzip.Writer，并将压缩后的数据写入 ResponseWriter。
func (w *gzipWriterWrapper) WriteString(s string) (int, error) {
	return io.WriteString(w.gzipWriter, s)
}

var gzipPool = sync.Pool{New: func() interface{} {
	return gzip.NewWriter(nil)
}}

func Compress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ginWriter := c.Writer
		writer := c.Writer

		// 检查客户端是否支持 gzip 压缩
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// 添加压缩响应头
		writer.Header().Set("Content-Encoding", "gzip")
		writer.Header().Set("Vary", "Accept-Encoding")

		// 创建 gzip.Writer
		gzipWriter := gzipPool.Get().(*gzip.Writer)
		gzipWriter.Reset(ginWriter)

		// 修改 gin.Context.Writer 为 gzipWriter
		c.Writer = &gzipWriterWrapper{
			ResponseWriter: ginWriter,
			gzipWriter:     gzipWriter,
		}

		defer func() {
			// 恢复 gin.Context.Writer
			c.Writer = ginWriter

			// 关闭 gzipWriter 并放回池中
			gzipWriter.Close()
			gzipPool.Put(gzipWriter)
		}()

		// 执行下一个处理程序
		c.Next()
	}
}
