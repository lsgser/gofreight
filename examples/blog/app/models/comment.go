package models

/*
|--------------------------------------------------------------------------
| Comment
|--------------------------------------------------------------------------
|
| Implements Comment as part of the models package in the Gofreight
| framework. Key symbols: Comment, Validators, Save.
| 
| Symbols defined here include: Comment (exported type); Comments
| (exported value).
| 
*/

import (
	"context"

	"github.com/lsgser/gofreight/model"
)

type Comment struct {
	model.Record
	PostID int64  `db:"post_id" json:"post_id"`
	Body   string `db:"body" json:"body"`
	Author string `db:"author" json:"author"`
}

var Comments = model.NewRepository[Comment]("comments")

func (c *Comment) Validators() []model.Validator {
	return []model.Validator{
		model.Presence("Body"),
		model.Length("Body", 2, 0),
	}
}

func (c *Comment) Save(ctx context.Context) error {
	return Comments.Save(ctx, c)
}

func init() {
	Posts.Association(model.HasManyAssociation("Comments", "comments", "post_id"))
	Comments.Association(model.BelongsToAssociation("Post", "posts", "post_id"))
}
