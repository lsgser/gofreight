package generator

import (
	"os"
	"path/filepath"
	"strings"
)

func installAuthStarter(appPath string) error {
	module := filepath.Base(appPath)
	if module == "." {
		module = "app"
	}
	data := struct{ Module string }{Module: module}

	files := map[string]string{
		"app/controllers/auth_controller.go": authControllerTmpl,
		"app/views/auth/login.gft":           authLoginViewTmpl,
		"app/views/auth/register.gft":        authRegisterViewTmpl,
		"routes/auth.go":                     authRoutesTmpl,
		"app/auth/users.go":                authUsersHelperTmpl,
	}
	for rel, tmpl := range files {
		path := filepath.Join(appPath, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := writeTemplate(path, tmpl, data); err != nil {
			return err
		}
	}
	return appendAuthRoutesRegister(appPath)
}

func appendAuthRoutesRegister(appPath string) error {
	path := filepath.Join(appPath, "routes/register.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := string(data)
	if strings.Contains(content, "Auth(r)") {
		return nil
	}
	needle := "\tWeb(r)\n"
	if !strings.Contains(content, needle) {
		return nil
	}
	content = strings.Replace(content, needle, needle+"\tAuth(r)\n", 1)
	return os.WriteFile(path, []byte(content), 0644)
}

const authControllerTmpl = `package controllers

import (
	appauth "{{.Module}}/app/auth"
	"github.com/lsgser/gofreight/auth"
	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/router"
)

type AuthController struct{}

func (c AuthController) ShowLogin(base controller.Base) error {
	return base.RenderView("auth/login", base.ViewData(nil))
}

func (c AuthController) ShowRegister(base controller.Base) error {
	return base.RenderView("auth/register", base.ViewData(nil))
}

func (c AuthController) Login(base controller.Base) error {
	return auth.Login(auth.DefaultLoginConfig(appauth.FindUserByEmail))(base)
}

func (c AuthController) Register(base controller.Base) error {
	if err := base.Request.ParseForm(); err != nil {
		return err
	}
	email := base.Request.FormValue("email")
	password := base.Request.FormValue("password")
	if err := appauth.RegisterUser(base.Request.Context(), email, password); err != nil {
		base.Unprocessable(map[string][]string{"email": {err.Error()}})
		return nil
	}
	return auth.Login(auth.DefaultLoginConfig(appauth.FindUserByEmail))(base)
}

func (c AuthController) Logout(base controller.Base) error {
	return auth.Logout("current_user_id", "/login")(base)
}

func RegisterAuthRoutes(r *router.Router) {
	c := AuthController{}
	r.Get("/login", controller.Handler(c.ShowLogin), "login")
	r.Post("/login", controller.Handler(c.Login))
	r.Get("/register", controller.Handler(c.ShowRegister), "register")
	r.Post("/register", controller.Handler(c.Register))
	r.Post("/logout", controller.Handler(c.Logout))
}
`

const authLoginViewTmpl = `#layout "layouts/application"

<h1>Login</h1>

#form action="/login" method="POST"
  #field "email" label="Email" type="email"
  #field "password" label="Password" type="password"
  #token
  <button type="submit">Sign in</button>
#endform

<p><a href="/register">Create an account</a></p>
`

const authRegisterViewTmpl = `#layout "layouts/application"

<h1>Register</h1>

#form action="/register" method="POST"
  #field "email" label="Email" type="email"
  #field "password" label="Password" type="password"
  #token
  <button type="submit">Create account</button>
#endform

<p><a href="/login">Already have an account?</a></p>
`
const authRoutesTmpl = `package routes

import (
	"{{.Module}}/app/controllers"
	"github.com/lsgser/gofreight/router"
)

func Auth(r *router.Router) {
	controllers.RegisterAuthRoutes(r)
}
`

const authUsersHelperTmpl = `package auth

import (
	"context"
	"fmt"

	"{{.Module}}/app/models"
	"github.com/lsgser/gofreight/auth"
)

func FindUserByEmail(email string) (*auth.User, error) {
	user, err := models.Users.FindBy(context.Background(), "email", email)
	if err != nil || user == nil {
		return nil, fmt.Errorf("not found")
	}
	return &auth.User{ID: user.ID, Email: user.Email, Role: user.Role, Password: user.Password}, nil
}

func RegisterUser(ctx context.Context, email, password string) error {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	user := &models.User{Email: email, Password: hash, Role: "user"}
	return user.Save(ctx)
}
`
