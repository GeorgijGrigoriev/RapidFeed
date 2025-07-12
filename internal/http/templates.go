package http

import (
	"github.com/GeorgijGrigoriev/RapidFeed"
	"html/template"
)

func PrepareTemplate(name ...string) *template.Template {
	tmpl := template.New("").Funcs(template.FuncMap{
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
	})

	tmpl, err := tmpl.ParseFS(RapidFeed.HTMLTemplates, name...)
	if err != nil {
		panic(err)
	}

	return tmpl
}
