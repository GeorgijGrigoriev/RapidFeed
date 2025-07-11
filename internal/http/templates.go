package http

import (
	"github.com/GeorgijGrigoriev/RapidFeed"
	"html/template"
)

func PrepareTemplate() *template.Template {
	return template.Must(template.New("template").Funcs(template.FuncMap{
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
	}).Parse(RapidFeed.IndexTemplate))
}
