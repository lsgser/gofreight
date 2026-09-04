package models

import (
	"context"

	"github.com/lsgser/gofreight/model"
)

type Post struct {
	model.Record
	Title string `db:"title" json:"title"`
	Body  string `db:"body" json:"body"`
	cbs   *model.Callbacks
}

var Posts = model.NewRepository[Post]("posts").
	Scope("recent", func(q *model.Query[Post]) *model.Query[Post] {
		return q.Latest()
	})

func (p *Post) Callbacks() *model.Callbacks {
	if p.cbs == nil {
		p.cbs = model.NewCallbacks()
	}
	return p.cbs
}

func (p *Post) Validators() []model.Validator {
	return []model.Validator{
		model.Presence("Title"),
		model.Length("Title", 3, 255),
		model.Presence("Body"),
		model.Length("Body", 10, 0),
	}
}

func (p *Post) Save(ctx context.Context) error {
	return Posts.Save(ctx, p)
}

func (p *Post) Delete(ctx context.Context) error {
	return Posts.Delete(ctx, p.ID)
}
