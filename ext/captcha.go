package ext

import (
	"fmt"

	"github.com/cjungo/cjungo"
	"github.com/mojocn/base64Captcha"
	"github.com/rs/zerolog"
)

type CaptchaController struct {
	logger *zerolog.Logger
}

func NewCaptchaController(
	logger *zerolog.Logger,
) *CaptchaController {
	return &CaptchaController{
		logger: logger,
	}
}

func (controller *CaptchaController) GenerateMath(ctx cjungo.HttpContext) error {
	store := base64Captcha.DefaultMemStore
	driver := &base64Captcha.DriverMath{
		Height:          100,
		Width:           400,
		NoiseCount:      4,
		ShowLineOptions: base64Captcha.OptionShowHollowLine | base64Captcha.OptionShowSlimeLine | base64Captcha.OptionShowSineLine,
		Fonts: []string{
			"wqy-microhei.ttc",
			"Comismsh.ttf",
			"ApothecaryFont.ttf",
			"3Dumb.ttf",
			"DENNEthree-dee.ttf",
			"DeborahFancyDress.ttf",
			"Flim-Flam.ttf",
		},
	}
	driver = driver.ConvertFonts()
	c := base64Captcha.NewCaptcha(driver, store)
	id, b64s, answer, err := c.Generate()
	if err != nil {
		return ctx.RespBad(err)
	}
	controller.logger.Info().
		Str("id", id).
		Str("answer", answer).
		Msg("[CAPTCHA]")
	return ctx.Resp(map[string]any{
		"id":    id,
		"image": b64s,
	})
}

func (controller *CaptchaController) Verify(id string, answer string, clear bool) error {
	store := base64Captcha.DefaultMemStore
	if store.Verify(id, answer, clear) {
		return nil
	} else {
		return fmt.Errorf("验证码有误")
	}
}
