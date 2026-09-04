package factories

import (
	"fmt"

	"blog/app/models"
	"github.com/lsgser/gofreight/gftest"
)

// PostFactory creates test posts.
var PostFactory = gftest.NewFactory(models.Posts).Define(map[string]any{
	"title": "Test Post",
	"body":  "This is a test post body with enough content.",
})

// PostSequence creates posts with sequenced titles.
func PostSequence(n int) map[string]any {
	return map[string]any{
		"title": fmt.Sprintf("Post #%d", n+1),
		"body":  fmt.Sprintf("Body content for post number %d.", n+1),
	}
}

// CommentFactory creates test comments.
var CommentFactory = gftest.NewFactory(models.Comments).Define(map[string]any{
	"body":   "Great post!",
	"author": "Test User",
})
