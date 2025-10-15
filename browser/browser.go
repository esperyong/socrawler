package browser

import (
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/headless_browser"
)

// NewCleanBrowser 创建一个不加载任何cookies的干净浏览器实例
func NewCleanBrowser(headless bool) *headless_browser.Browser {
	opts := []headless_browser.Option{
		headless_browser.WithHeadless(headless),
	}

	logrus.Debugf("Created clean browser instance (headless: %v)", headless)
	return headless_browser.New(opts...)
}
