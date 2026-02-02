package http

import (
	"net/http"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/ui"
	"github.com/gofiber/template/html/v2"
)

func initTemplateEngine() *html.Engine {
	engine := html.NewFileSystem(http.FS(ui.HTMLTemplates), ".html")

	engine.AddFuncMap(tmplFuncMap())

	return engine
}

func tmplFuncMap() map[string]any {
	return map[string]any{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"seq": func(start, end int) []int {
			if start > end {
				start, end = end, start
			}
			seq := make([]int, 0, end-start+1)
			for i := start; i <= end; i++ {
				seq = append(seq, i)
			}

			return seq
		},
		"max": func(a, b int) int {
			if a > b {
				return a
			}

			return b
		},
		"min": func(a, b int) int {
			if a > b {
				return b
			}

			return a
		},
	}
}
