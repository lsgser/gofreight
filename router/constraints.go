package router

import (
	"regexp"
)

// Where attaches regex constraints to route parameters (Laravel where()).
//
//	r.Get("/posts/:id", handler).Where(map[string]string{"id": "[0-9]+"})
func (reg *RouteRegistrar) Where(constraints map[string]string) *RouteRegistrar {
	if reg == nil || reg.index < 0 || reg.index >= len(reg.router.routes) {
		return reg
	}
	compiled, err := compileConstraints(constraints)
	if err != nil {
		panic(err)
	}
	route := &reg.router.routes[reg.index]
	if route.constraints == nil {
		route.constraints = compiled
	} else {
		for k, v := range compiled {
			route.constraints[k] = v
		}
	}
	return reg
}

// WhereParam constrains a single route parameter.
func (reg *RouteRegistrar) WhereParam(name, pattern string) *RouteRegistrar {
	return reg.Where(map[string]string{name: pattern})
}

// compileRouteConstraints validates and stores constraints on a route during registration.
func (r *Router) compileRouteConstraints(route *Route, raw map[string]string) error {
	compiled, err := compileConstraints(raw)
	if err != nil {
		return err
	}
	route.constraints = compiled
	return nil
}

// MustCompileConstraint compiles a constraint pattern for reuse.
func MustCompileConstraint(pattern string) *regexp.Regexp {
	re, err := regexp.Compile("^" + pattern + "$")
	if err != nil {
		panic(err)
	}
	return re
}
