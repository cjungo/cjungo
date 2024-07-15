package ext

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cjungo/cjungo"
	"github.com/h2non/filetype"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type StorageConf struct {
	dig.In
	PathPrefix       string
	Dir              string
	UploadMiddleware []echo.MiddlewareFunc
	IndexMiddleware  []echo.MiddlewareFunc
	QueryMiddleware  []echo.MiddlewareFunc
}

func StorageFor(
	router cjungo.HttpRouter,
	logger *zerolog.Logger,
	conf *StorageConf,
) *StorageController {
	controller := &StorageController{
		pathPrefix: conf.PathPrefix,
		dir:        conf.Dir,
		logger:     logger,
	}
	router.GET(conf.PathPrefix, controller.Index, conf.IndexMiddleware...)
	router.GET(fmt.Sprintf("%s/:filename", conf.PathPrefix), controller.Query, conf.QueryMiddleware...)
	router.POST(conf.PathPrefix, controller.Upload, conf.UploadMiddleware...)
	router.POST(fmt.Sprintf("%s/:dir", conf.PathPrefix), controller.Upload, conf.UploadMiddleware...)

	logger.Info().
		Str("action", "StorageFor").
		Str("prefix", conf.PathPrefix).
		Str("dir", conf.Dir).
		Msg("[STORAGE]")

	return controller
}

type StorageController struct {
	pathPrefix string
	dir        string
	logger     *zerolog.Logger
}

func (controller *StorageController) Upload(ctx cjungo.HttpContext) error {
	fh, err := ctx.FormFile("file")
	if err != nil {
		return err
	}
	f, err := fh.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	dstItems := []string{controller.dir}
	mid := ctx.Param("dir")
	if len(mid) > 0 {
		dstItems = append(dstItems, mid)
	}
	dstItems = append(dstItems, fh.Filename)

	dstPath := filepath.Join(dstItems...)
	dstDir := filepath.Dir(dstPath)
	if !cjungo.IsDirExist(dstDir) {
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return err
		}
	}
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, f); err != nil {
		return err
	}

	return ctx.RespOk()
}

func (controller *StorageController) Index(ctx cjungo.HttpContext) error {
	result, err := cjungo.GlobDir(controller.dir)
	if err != nil {
		return err
	}
	filenames := make([]string, len(result))
	for i, filename := range result {
		p, err := filepath.Rel(controller.dir, filename)
		if err != nil {
			return err
		}
		filenames[i] = strings.ReplaceAll(p, "\\", "/")
	}
	return ctx.Resp(filenames)
}

func (controller *StorageController) Query(ctx cjungo.HttpContext) error {
	filename := ctx.Param("filename")
	path := filepath.Join(controller.dir, filename)
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("不可访问目录")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	ext := strings.Trim(filepath.Ext(path), ".")
	controller.logger.Info().
		Str("ext", ext).
		Msg("[STORAGE]")

	return ctx.Stream(200, filetype.GetType(ext).MIME.Value, f)
}
