// proxy/proxy.go
package proxy

import (
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// ServiceProxy belirli bir servis için proxy işlemini gerçekleştirir
type ServiceProxy struct {
	ServiceURL string
	Name       string
}

// NewServiceProxy yeni bir servis proxy'si oluşturur
func NewServiceProxy(serviceURL, name string) *ServiceProxy {
	return &ServiceProxy{
		ServiceURL: serviceURL,
		Name:       name,
	}
}

// ReverseProxy reverse proxy işlemini gerçekleştirir
func (s *ServiceProxy) ReverseProxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		remote, err := url.Parse(s.ServiceURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Proxy target parsing error"})
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(remote)

		// Orijinal isteğin yolunu düzenle
		// Örneğin: /api/data-collector/pollution -> /pollution
		path := c.Param("proxyPath")
		
		// Hedef URL oluştur
		c.Request.URL.Path = path
		c.Request.URL.RawPath = path
		
		// Hedef sunucuya isteği yönlendir
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// ForwardRequest isteği belirtilen servise yönlendirir
func (s *ServiceProxy) ForwardRequest(method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	// Hedef URL oluştur
	targetURL := strings.TrimRight(s.ServiceURL, "/") + "/" + strings.TrimLeft(path, "/")
	
	// HTTP isteği oluştur
	req, err := http.NewRequest(method, targetURL, body)
	if err != nil {
		return nil, err
	}
	
	// Header'ları ekle
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	// İsteği gönder
	client := &http.Client{}
	return client.Do(req)
}