package controller

import (
	"html/template"
	"net/http"

	"github.com/raminderis/lenslocked/views"
)

func StaticHandler(tpl views.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, nil)
	}
}

func FAQ(tpl views.Template) http.HandlerFunc {
	questions := []struct {
		Question string
		Answer   template.HTML
	}{
		{
			Question: "Is there a free version?",
			Answer:   "Yes we do offer free version.",
		},
		{
			Question: "What are your support hours?",
			Answer:   "we have support staff emails 24/7, slower on weekends.",
		},
		{
			Question: "how do i contact support?",
			Answer:   `email us at <a href="mailto:someem@live.com">someemail@live.com</a>`,
		},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, questions)
	}
}
