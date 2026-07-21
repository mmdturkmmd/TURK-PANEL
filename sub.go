// Package sub provides subscription route registration for the Spider Panel.
// In Railway single-port mode, subscription routes are mounted inside the main
// web server — this package never starts its own HTTP listener.
package sub

import (
	"context"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/mhsanaei/3x-ui/v2/logger"
	"github.com/mhsanaei/3x-ui/v2/web/locale"
	"github.com/mhsanaei/3x-ui/v2/web/middleware"
	"github.com/mhsanaei/3x-ui/v2/web/service"

	"github.com/gin-gonic/gin"
)

// SettingProvider is the minimal dependency surface the sub package needs
// from the web service. Injecting an interface keeps the `sub` package free
// of an import cycle (sub -> web/service -> sub).
type SettingProvider interface {
	GetSubEnable() (bool, error)
	GetSubDomain() (string, error)
	GetSubPath() (string, error)
	GetSubJsonPath() (string, error)
	GetSubClashPath() (string, error)
	GetSubJsonEnable() (bool, error)
	GetSubClashEnable() (bool, error)
	GetSubEncrypt() (bool, error)
	GetSubShowInfo() (bool, error)
	GetRemarkModel() (string, error)
	GetSubUpdates() (string, error)
	GetSubJsonFragment() (string, error)
	GetSubJsonNoises() (string, error)
	GetSubJsonMux() (string, error)
	GetSubJsonRules() (string, error)
	GetSubTitle() (string, error)
	GetSubSupportUrl() (string, error)
	GetSubProfileUrl() (string, error)
	GetSubAnnounce() (string, error)
	GetSubEnableRouting() (bool, error)
	GetSubRoutingRules() (string, error)
}

// Server holds the subscription route state. It does NOT own a listener.
type Server struct {
	sub *SUBController
	ctx context.Context
}

// NewServer creates a subscription router helper with a cancellable context.
func NewServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel // retained for future graceful teardown hooks
	return &Server{ctx: ctx}
}

// RegisterRoutes mounts subscription routes on an existing Gin engine.
// htmlFS and assetsFS are the embedded filesystems passed from the web package
// to avoid importing web from sub (import cycle prevention).
func (s *Server) RegisterRoutes(router *gin.Engine, htmlFS fs.FS, assetsFS fs.FS) error {
	setting := service.SettingService{}
	if err := s.register(router, &setting, htmlFS, assetsFS); err != nil {
		return err
	}
	return nil
}

func (s *Server) register(router *gin.Engine, setting SettingProvider, htmlFS fs.FS, assetsFS fs.FS) error {
	subEnable, err := setting.GetSubEnable()
	if err != nil {
		return err
	}
	if !subEnable {
		return nil
	}

	subDomain, err := setting.GetSubDomain()
	if err != nil {
		return err
	}
	if subDomain != "" {
		router.Use(middleware.DomainValidatorMiddleware(subDomain))
	}

	LinksPath, err := setting.GetSubPath()
	if err != nil {
		return err
	}

	JsonPath, err := setting.GetSubJsonPath()
	if err != nil {
		return err
	}

	ClashPath, err := setting.GetSubClashPath()
	if err != nil {
		return err
	}

	subJsonEnable, err := setting.GetSubJsonEnable()
	if err != nil {
		return err
	}

	subClashEnable, err := setting.GetSubClashEnable()
	if err != nil {
		return err
	}

	basePath := LinksPath
	if basePath != "/" && !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}
	router.Use(func(c *gin.Context) {
		c.Set("base_path", basePath)
	})

	Encrypt, err := setting.GetSubEncrypt()
	if err != nil {
		return err
	}

	ShowInfo, err := setting.GetSubShowInfo()
	if err != nil {
		return err
	}

	RemarkModel, err := setting.GetRemarkModel()
	if err != nil {
		RemarkModel = "-ieo"
	}

	SubUpdates, err := setting.GetSubUpdates()
	if err != nil {
		SubUpdates = "10"
	}

	SubJsonFragment, err := setting.GetSubJsonFragment()
	if err != nil {
		SubJsonFragment = ""
	}

	SubJsonNoises, err := setting.GetSubJsonNoises()
	if err != nil {
		SubJsonNoises = ""
	}

	SubJsonMux, err := setting.GetSubJsonMux()
	if err != nil {
		SubJsonMux = ""
	}

	SubJsonRules, err := setting.GetSubJsonRules()
	if err != nil {
		SubJsonRules = ""
	}

	SubTitle, err := setting.GetSubTitle()
	if err != nil {
		SubTitle = ""
	}

	SubSupportUrl, err := setting.GetSubSupportUrl()
	if err != nil {
		SubSupportUrl = ""
	}

	SubProfileUrl, err := setting.GetSubProfileUrl()
	if err != nil {
		SubProfileUrl = ""
	}

	SubAnnounce, err := setting.GetSubAnnounce()
	if err != nil {
		SubAnnounce = ""
	}

	SubEnableRouting, err := setting.GetSubEnableRouting()
	if err != nil {
		return err
	}

	SubRoutingRules, err := setting.GetSubRoutingRules()
	if err != nil {
		SubRoutingRules = ""
	}

	router.Use(locale.LocalizerMiddleware())

	i18nWebFunc := func(key string, params ...string) string {
		return locale.I18n(locale.Web, key, params...)
	}
	router.SetFuncMap(map[string]any{"i18n": i18nWebFunc})

	// NOTE: Do NOT call engine.SetHTMLTemplate() here. The main web server
	// (web/web.go) already loads the full template set including subpage.html.
	// Re-setting templates on the shared engine would wipe the panel UI.

	// Assets: prefer disk (debug), fallback to embedded
	var fs2 http.FileSystem
	if _, err := os.Stat("web/assets"); err == nil {
		fs2 = http.FS(os.DirFS("web/assets"))
	} else if assetsFS != nil {
		if subFS, err := fs.Sub(assetsFS, "assets"); err == nil {
			fs2 = http.FS(subFS)
		} else {
			logger.Error("sub: failed to mount embedded assets:", err)
		}
	}

	if fs2 != nil {
		// Skip /assets — web.go initRouter already mounts it.
		// Only mount at custom basePath (subscription Path).
		if LinksPath != "/" {
			linksPathForAssets := strings.TrimRight(LinksPath, "/") + "/assets"
			router.StaticFS(linksPathForAssets, fs2)

			router.Use(func(c *gin.Context) {
				path := c.Request.URL.Path
				pathPrefix := strings.TrimRight(LinksPath, "/") + "/"
				if strings.HasPrefix(path, pathPrefix) && strings.Contains(path, "/assets/") {
					assetsIndex := strings.Index(path, "/assets/")
					if assetsIndex != -1 {
						assetPath := path[assetsIndex+8:]
						if assetPath != "" {
							c.FileFromFS(assetPath, fs2)
							c.Abort()
							return
						}
					}
				}
				c.Next()
			})
		}
	}

	g := router.Group("/")
	s.sub = NewSUBController(
		g, LinksPath, JsonPath, ClashPath, subJsonEnable, subClashEnable, Encrypt, ShowInfo, RemarkModel, SubUpdates,
		SubJsonFragment, SubJsonNoises, SubJsonMux, SubJsonRules, SubTitle, SubSupportUrl,
		SubProfileUrl, SubAnnounce, SubEnableRouting, SubRoutingRules)

	return nil
}

// Stop is a no-op for the single-port mode (no listener to close).
// Retained for API symmetry with the web Server.
func (s *Server) Stop() error {
	s.ctx.Done()
	return nil
}

// GetCtx returns the server's context.
func (s *Server) GetCtx() context.Context {
	return s.ctx
}
