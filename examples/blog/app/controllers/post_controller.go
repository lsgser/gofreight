package controllers

/*
|--------------------------------------------------------------------------
| Post Controller
|--------------------------------------------------------------------------
|
| Implements Post Controller as part of the controllers package in the
| Gofreight framework. Key symbols: PostController, RegisterPostRoutes,
| Index, Show, Create, Update.
| 
| Symbols defined here include: PostController (exported type).
| 
*/

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"blog/app/models"
	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/router"
)

type PostController struct{}

func RegisterPostRoutes(r *router.Router) {
	c := PostController{}
	r.Resources("posts", router.ResourceHandlers{
		Index:   controller.Handler(c.Index),
		Show:    controller.Handler(c.Show),
		Create:  controller.Handler(c.Create),
		Update:  controller.Handler(c.Update),
		Destroy: controller.Handler(c.Destroy),
	})
}

func (c PostController) Index(base controller.Base) error {
	posts, err := models.Posts.Query(context.Background()).Scope("recent").With("Comments").Get()
	if err != nil {
		return err
	}

	if wantsJSON(base.Request) {
		base.RenderJSON(posts)
		return nil
	}

	return base.RenderView("posts/index", map[string]any{"Posts": posts})
}

func (c PostController) Show(base controller.Base) error {
	id, err := strconv.ParseInt(base.Param("id"), 10, 64)
	if err != nil {
		base.NotFound("invalid id")
		return nil
	}

	post, err := models.Posts.Find(context.Background(), id)
	if err != nil {
		base.NotFound("Post not found")
		return nil
	}

	if wantsJSON(base.Request) {
		base.RenderJSON(post)
		return nil
	}

	return base.RenderView("posts/show", map[string]any{"Post": post})
}

func (c PostController) Create(base controller.Base) error {
	var post models.Post
	if err := json.NewDecoder(base.Request.Body).Decode(&post); err != nil {
		base.Unprocessable(map[string][]string{"base": {"invalid JSON"}})
		return nil
	}

	if err := post.Save(context.Background()); err != nil {
		if strings.Contains(err.Error(), "validation failed") {
			base.Unprocessable(map[string][]string{"base": {err.Error()}})
			return nil
		}
		return err
	}

	base.Status = http.StatusCreated
	base.RenderJSON(post)
	return nil
}

func (c PostController) Update(base controller.Base) error {
	id, err := strconv.ParseInt(base.Param("id"), 10, 64)
	if err != nil {
		base.NotFound("invalid id")
		return nil
	}

	post, err := models.Posts.Find(context.Background(), id)
	if err != nil {
		base.NotFound("Post not found")
		return nil
	}

	if err := json.NewDecoder(base.Request.Body).Decode(post); err != nil {
		base.Unprocessable(map[string][]string{"base": {"invalid JSON"}})
		return nil
	}
	post.ID = id

	if err := post.Save(context.Background()); err != nil {
		if strings.Contains(err.Error(), "validation failed") {
			base.Unprocessable(map[string][]string{"base": {err.Error()}})
			return nil
		}
		return err
	}

	base.RenderJSON(post)
	return nil
}

func (c PostController) Destroy(base controller.Base) error {
	id, err := strconv.ParseInt(base.Param("id"), 10, 64)
	if err != nil {
		base.NotFound("invalid id")
		return nil
	}

	if err := models.Posts.Delete(context.Background(), id); err != nil {
		return err
	}

	base.Status = http.StatusNoContent
	base.Response.WriteHeader(http.StatusNoContent)
	return nil
}

func wantsJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "application/json")
}
