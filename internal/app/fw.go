package app

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// --- 核心类型 ---
type MyContext struct {
	*gin.Context
	RequestContainer *Container
}

type (
	NextFunc        func() (any, error)
	MiddlewareFunc  func(*MyContext)
	GuardFunc       func(*MyContext) (bool, error)
	PipeFunc        func(ctx *MyContext, val any) (any, error)
	InterceptorFunc func(ctx *MyContext, data any, next NextFunc) (any, error)
	FilterFunc      func(ctx *MyContext, err error)
)

type (
	OnModuleInit  interface{ OnModuleInit() }
	BeforeRequest interface{ BeforeRequest(*MyContext) }
	AfterRequest  interface{ AfterRequest(*MyContext) }
)

type Body[T any] struct{ Value T }
type Query[T any] struct{ Value T }
type Headers struct{ Value http.Header }
type Cookies map[string]string
type Files struct{ Files []*multipart.FileHeader }

// --- DI 容器 ---
type Container struct {
	parent     *Container
	singletons map[reflect.Type]any
}

func NewContainer(parent ...*Container) *Container {
	c := &Container{singletons: map[reflect.Type]any{}}
	if len(parent) > 0 {
		c.parent = parent[0]
	}
	return c
}

func (c *Container) Provide(v any) {
	t := reflect.TypeOf(v)
	c.singletons[t] = v
	if h, ok := v.(OnModuleInit); ok {
		h.OnModuleInit()
	}
}

func (c *Container) Resolve(t reflect.Type) reflect.Value {
	if v, ok := c.singletons[t]; ok {
		return reflect.ValueOf(v)
	}
	if c.parent != nil {
		if v := c.parent.Resolve(t); v.IsValid() {
			return v
		}
	}
	if t.Kind() == reflect.Ptr {
		val := reflect.New(t.Elem()).Interface()
		c.singletons[t] = val
		return reflect.ValueOf(val)
	}
	return reflect.Value{}
}

// --- 路由与配置 ---
type RouteMeta struct {
	middlewares  []MiddlewareFunc
	guards       []GuardFunc
	pipes        []PipeFunc
	interceptors []InterceptorFunc
	filters      []FilterFunc
}

type App struct {
	engine    *gin.Engine
	container *Container
	meta      RouteMeta
}

type RouterGroup struct {
	prefix string
	app    *App
	meta   RouteMeta
}

func NewApp() *App {
	return &App{engine: gin.Default(), container: NewContainer()}
}

func (a *App) Use(m ...MiddlewareFunc) *App {
	a.meta.middlewares = append(a.meta.middlewares, m...)
	return a
}
func (a *App) UseGuards(gs ...GuardFunc) *App { a.meta.guards = append(a.meta.guards, gs...); return a }
func (a *App) UsePipes(p ...PipeFunc) *App    { a.meta.pipes = append(a.meta.pipes, p...); return a }
func (a *App) UseInterceptors(i ...InterceptorFunc) *App {
	a.meta.interceptors = append(a.meta.interceptors, i...)
	return a
}
func (a *App) UseFilters(f ...FilterFunc) *App {
	a.meta.filters = append(a.meta.filters, f...)
	return a
}
func (a *App) Group(prefix string) *RouterGroup { return &RouterGroup{prefix: prefix, app: a} }

func (g *RouterGroup) Use(m ...MiddlewareFunc) *RouterGroup {
	g.meta.middlewares = append(g.meta.middlewares, m...)
	return g
}
func (g *RouterGroup) UseGuards(gs ...GuardFunc) *RouterGroup {
	g.meta.guards = append(g.meta.guards, gs...)
	return g
}
func (g *RouterGroup) UsePipes(p ...PipeFunc) *RouterGroup {
	g.meta.pipes = append(g.meta.pipes, p...)
	return g
}
func (g *RouterGroup) UseInterceptors(i ...InterceptorFunc) *RouterGroup {
	g.meta.interceptors = append(g.meta.interceptors, i...)
	return g
}
func (g *RouterGroup) UseFilters(f ...FilterFunc) *RouterGroup {
	g.meta.filters = append(g.meta.filters, f...)
	return g
}

func (g *RouterGroup) GET(path string, handler any)  { g.Handle("GET", path, handler) }
func (g *RouterGroup) POST(path string, handler any) { g.Handle("POST", path, handler) }

// --- 执行引擎 ---
func (g *RouterGroup) Handle(method, path string, handler any) {
	hType := reflect.TypeOf(handler)
	g.app.engine.Handle(method, g.prefix+path, func(c *gin.Context) {
		ctx := &MyContext{Context: c, RequestContainer: NewContainer(g.app.container)}

		// 自动化激活：预实例化 Handler 需要的 Service
		for i := 0; i < hType.NumIn(); i++ {
			t := hType.In(i)
			if t.Kind() == reflect.Ptr && t != reflect.TypeOf(&MyContext{}) {
				instance := ctx.RequestContainer.Resolve(t)
				ctx.RequestContainer.singletons[t] = instance.Interface()
			}
		}

		allM := append(g.app.meta.middlewares, g.meta.middlewares...)
		allG := append(g.app.meta.guards, g.meta.guards...)
		allI := append(g.app.meta.interceptors, g.meta.interceptors...)
		allF := append(g.meta.filters, g.app.meta.filters...)

		for _, m := range allM {
			m(ctx)
			if ctx.IsAborted() {
				return
			}
		}

		g.triggerHooks(ctx, "BeforeRequest")

		for _, guard := range allG {
			if ok, err := guard(ctx); !ok || err != nil {
				g.runFilters(ctx, err, allF)
				return
			}
		}

		coreLogic := func() (any, error) {
			args, err := g.resolveParams(ctx, handler)
			if err != nil {
				return nil, err
			}
			res := reflect.ValueOf(handler).Call(args)
			var out any
			var e error
			if len(res) > 0 {
				out = res[0].Interface()
			}
			if len(res) > 1 && !res[1].IsNil() {
				e = res[1].Interface().(error)
			}
			return out, e
		}

		chain := coreLogic
		for i := len(allI) - 1; i >= 0; i-- {
			idx := i
			next := chain
			chain = func() (any, error) { return allI[idx](ctx, nil, next) }
		}

		result, finalErr := chain()
		g.triggerHooks(ctx, "AfterRequest")

		if finalErr != nil {
			g.runFilters(ctx, finalErr, allF)
			return
		}
		if !ctx.IsAborted() && result != nil {
			ctx.JSON(http.StatusOK, gin.H{"data": result})
		}
	})
}

func (g *RouterGroup) triggerHooks(ctx *MyContext, hookName string) {
	for _, v := range ctx.RequestContainer.singletons {
		if h, ok := v.(BeforeRequest); ok && hookName == "BeforeRequest" {
			h.BeforeRequest(ctx)
		}
		if h, ok := v.(AfterRequest); ok && hookName == "AfterRequest" {
			h.AfterRequest(ctx)
		}
	}
}

func (g *RouterGroup) runFilters(ctx *MyContext, err error, filters []FilterFunc) {
	if len(filters) > 0 {
		filters[0](ctx, err)
		return
	}
	ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
}

func (g *RouterGroup) resolveParams(ctx *MyContext, handler any) ([]reflect.Value, error) {
	ht := reflect.TypeOf(handler)
	args := make([]reflect.Value, 0, ht.NumIn())
	for i := 0; i < ht.NumIn(); i++ {
		pt := ht.In(i)
		var val any
		typeName := pt.Name()
		switch {
		case pt == reflect.TypeOf(&MyContext{}):
			args = append(args, reflect.ValueOf(ctx))
			continue
		case strings.Contains(typeName, "Body"):
			v := reflect.New(pt.Field(0).Type).Interface()
			if err := ctx.ShouldBindJSON(v); err != nil {
				return nil, err
			}
			val = reflect.ValueOf(v).Elem().Interface()
		case strings.Contains(typeName, "Query"):
			v := reflect.New(pt.Field(0).Type).Interface()
			if err := ctx.ShouldBindQuery(v); err != nil {
				return nil, err
			}
			val = reflect.ValueOf(v).Elem().Interface()
		case pt == reflect.TypeOf(Headers{}):
			args = append(args, reflect.ValueOf(Headers{Value: ctx.Request.Header}))
			continue
		case pt == reflect.TypeOf(Cookies{}):
			m := make(Cookies)
			for _, c := range ctx.Request.Cookies() {
				m[c.Name] = c.Value
			}
			args = append(args, reflect.ValueOf(m))
			continue
		case pt == reflect.TypeOf(Files{}):
			form, _ := ctx.MultipartForm()
			args = append(args, reflect.ValueOf(Files{Files: form.File["files"]}))
			continue
		default:
			args = append(args, ctx.RequestContainer.Resolve(pt))
			continue
		}
		allP := append(g.app.meta.pipes, g.meta.pipes...)
		for _, pipe := range allP {
			var err error
			val, err = pipe(ctx, val)
			if err != nil {
				return nil, err
			}
		}
		wrapper := reflect.New(pt).Elem()
		wrapper.Field(0).Set(reflect.ValueOf(val))
		args = append(args, wrapper)
	}
	return args, nil
}

// --- 1. 定义 Service ---
type TraceService struct{}

func (s *TraceService) BeforeRequest(ctx *MyContext) {
	fmt.Println("Step 3: [BeforeRequest Hook] 注入请求上下文成功")
}

func (s *TraceService) AfterRequest(ctx *MyContext) {
	fmt.Println("Step 11: [AfterRequest Hook] 资源回收成功")
}

func main() {
	app := NewApp()
	// 全局注入单例
	app.container.Provide(&TraceService{})

	// --- 核心修复：步骤 1 中间件 ---
	app.Use(func(ctx *MyContext) {
		fmt.Println("Step 1: [Global Middleware] 接入请求")

		// ★ 关键点：在钩子触发前，手动从容器中 Resolve 该服务
		// 这步操作会将 TraceService 实例化并放入当前请求的 RequestContainer 中
		// 这样随后的 Step 3 就能通过反射找到它并执行了
		ctx.RequestContainer.Resolve(reflect.TypeOf(&TraceService{}))
	})

	// 步骤 12: 过滤器
	app.UseFilters(func(ctx *MyContext, err error) {
		fmt.Printf("Step 12: [Filter] 捕获异常: %v\n", err)
	})

	v1 := app.Group("/api")

	v1.Use(func(ctx *MyContext) {
		fmt.Println("Step 2: [Route Middleware] 路由检查")
	}).UseGuards(func(ctx *MyContext) (bool, error) {
		fmt.Println("Step 4 & 5: [Guards] 权限校验通过")
		return true, nil
	}).UseInterceptors(func(ctx *MyContext, _ any, next NextFunc) (any, error) {
		fmt.Println("Step 6: [Interceptor Before] 开启监控")
		res, err := next()
		fmt.Println("Step 10: [Interceptor After] 包装响应")
		return res, err
	}).UsePipes(func(ctx *MyContext, val any) (any, error) {
		fmt.Println("Step 8: [Pipes] 数据清洗完成")
		return val, nil
	})

	// 步骤 9: Handler
	v1.POST("/trace", func(ctx *MyContext, body Body[struct {
		Msg string `json:"msg"`
	}], svc *TraceService) (any, error) {
		fmt.Println("Step 7: [Param Resolve] (由框架静默完成)")
		fmt.Println("Step 9: [Handler] 业务逻辑执行")

		if body.Value.Msg == "error" {
			return nil, errors.New("模拟报错")
		}
		return "Done", nil
	})

	fmt.Println(">> 12层全链路演示启动...")
	app.engine.Run(":8080")
}
