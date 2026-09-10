package factories

/*
|--------------------------------------------------------------------------
| Factories
|--------------------------------------------------------------------------
|
| Implements Factories as part of the factories package in the Gofreight
| framework. Key symbols: PostSequence.
| 
| Symbols defined here include: PostFactory (exported value);
| CommentFactory (exported value).
| 
*/

import (
	"fmt"

	"blog/app/models"
	"github.com/lsgser/gofreight/gftest"
	"github.com/lsgser/gofreight/gftest/faker"
)

var PostFactory = gftest.NewFactory(models.Posts).Define(map[string]any{
	"title": faker.Lazy(func() any { return faker.Sentence() }),
	"body":  faker.Lazy(func() any { return faker.Paragraph() }),
})

func PostSequence(n int) map[string]any {
	return map[string]any{
		"title": fmt.Sprintf("Post #%d", n+1),
		"body":  fmt.Sprintf("Body content for post number %d.", n+1),
	}
}

var CommentFactory = gftest.NewFactory(models.Comments).Define(map[string]any{
	"body":   faker.Lazy(func() any { return faker.Sentence() }),
	"author": faker.Lazy(func() any { return faker.Name() }),
})
