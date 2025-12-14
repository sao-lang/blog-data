package gnest

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ==========================================
// 1. 核心接口定义 (Core Interfaces)
// ==========================================

type argumentResolver func(c *gin.Context) (reflect.Value, error)

// 预分析结果：存储一个函数的所有参数解析器
type handlerMetadata struct {
	resolvers []argumentResolver
}

// DefaultExceptionFilter 是框架内置的兜底过滤器
type DefaultExceptionFilter struct{}

func (f *DefaultExceptionFilter) Catch(c *gin.Context, err error) {
	// 只有在业务没处理请求（没写入 Header）时才执行
	if !c.IsAborted() {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": 500,
			"message":    err.Error(),
			"error":      "Internal Server Error",
		})
		c.Abort()
	}
}

type CanActivate interface{ CanActivate(ctx *gin.Context) bool }
type NestInterceptor interface {
	Intercept(ctx *gin.Context, next func() interface{}) interface{}
}
type PipeTransform interface {
	Transform(value interface{}, targetType reflect.Type) (interface{}, error)
}
type ExceptionFilter interface {
	Catch(ctx *gin.Context, err error)
}

// --- 初始化阶段 ---
type OnModuleInit interface{ OnModuleInit() }                     // 所有依赖注入完成
type OnApplicationBootstrap interface{ OnApplicationBootstrap() } // 应用准备好接收请求前
// --- 终止阶段 ---
type OnModuleDestroy interface{ OnModuleDestroy() }                               // 收到信号，准备关闭（清理定时器等）
type BeforeApplicationShutdown interface{ BeforeApplicationShutdown(sig string) } // 停止接收连接，关闭 DB 前
type OnApplicationShutdown interface{ OnApplicationShutdown() }                   // 所有资源已释放，进程即将退出

// Render 用于渲染 HTML 模板
type Render struct {
	Name string      // 模板文件名
	Data interface{} // 模板数据
}

// RedirectResult 用于重定向
type RedirectResult struct {
	Code     int    // 状态码 (如 301, 302)
	Location string // 跳转目标地址
}

// DataResult 用于返回原始字节流 (如图片、验证码)
type DataResult struct {
	ContentType string
	Data        []byte
}

// FileResult 用于返回本地文件或触发下载
type FileResult struct {
	FilePath string
	FileName string // 如果不为空，则作为附件下载
}

// ==========================================
// 2. 核心 IoC 容器 (Container & DI)
// ==========================================

type GnestApp struct {
	Engine             *gin.Engine
	providers          map[reflect.Type]reflect.Value
	container          []interface{} // 用于扫描生命周期钩子
	globalGuards       []CanActivate
	globalInterceptors []NestInterceptor
	globalPipes        []PipeTransform
	globalFilters      []ExceptionFilter
	customDecorators   map[reflect.Type]func(c *gin.Context) interface{} // 补回：自定义参数装饰器
	validate           *validator.Validate                               // 增加：内置校验器
}

func New() *GnestApp {
	return &GnestApp{
		Engine:           gin.Default(),
		providers:        make(map[reflect.Type]reflect.Value),
		customDecorators: make(map[reflect.Type]func(c *gin.Context) interface{}),
		validate:         validator.New(),
		globalFilters:    []ExceptionFilter{&DefaultExceptionFilter{}},
	}
}

// Provide 注册并注入依赖
func (app *GnestApp) Provide(ps ...interface{}) *GnestApp {
	for _, p := range ps {
		v := reflect.ValueOf(p)
		app.providers[v.Type()] = v
		app.container = append(app.container, p)
	}
	// 递归注入并检测循环依赖
	for _, p := range app.container {
		app.inject(reflect.ValueOf(p), make(map[reflect.Type]bool))
	}
	return app
}

func (app *GnestApp) inject(v reflect.Value, path map[reflect.Type]bool) {
	// 只有指针能被 Elem()，且只有指针指向的结构体字段才能被 Set
	if v.Kind() != reflect.Ptr {
		return
	}

	el := v.Elem()
	if el.Kind() != reflect.Struct {
		return
	}

	t := el.Type()
	if path[t] {
		log.Fatalf("[Gnest Error] Circular dependency detected: %v", t)
	}
	path[t] = true
	defer delete(path, t)

	for i := 0; i < el.NumField(); i++ {
		f := el.Field(i)
		if f.CanSet() {
			if p, ok := app.providers[f.Type()]; ok {
				f.Set(p)
				// 递归注入
				app.inject(p, path)
			}
		}
	}
}

// AddCustomDecorator 注册自定义参数装饰器
func (app *GnestApp) AddCustomDecorator(targetType reflect.Type, factory func(c *gin.Context) interface{}) {
	app.customDecorators[targetType] = factory
}

// 全局增强器设置
func (app *GnestApp) UseGlobalGuards(gs ...CanActivate) *GnestApp {
	app.globalGuards = append(app.globalGuards, gs...)
	return app
}
func (app *GnestApp) UseGlobalPipes(ps ...PipeTransform) *GnestApp {
	app.globalPipes = append(app.globalPipes, ps...)
	return app
}
func (app *GnestApp) UseGlobalFilters(fs ...ExceptionFilter) *GnestApp {
	app.globalFilters = append(app.globalFilters, fs...)
	return app
}
func (app *GnestApp) Use(ms ...gin.HandlerFunc) *GnestApp {
	app.Engine.Use(ms...)
	return app
}
func (app *GnestApp) UseGlobalInterceptors(is ...NestInterceptor) *GnestApp {
	app.globalInterceptors = append(app.globalInterceptors, is...)
	return app
}

// ==========================================
// 3. 路由组与链式配置 (RouterGroup)
// ==========================================

type RouterGroup struct {
	app          *GnestApp
	ginGroup     *gin.RouterGroup
	guards       []CanActivate
	interceptors []NestInterceptor
	pipes        []PipeTransform
	filters      []ExceptionFilter
}

func (app *GnestApp) Group(path string) *RouterGroup {
	return &RouterGroup{app: app, ginGroup: app.Engine.Group(path)}
}

func (rg *RouterGroup) UseGuards(gs ...CanActivate) *RouterGroup {
	rg.guards = append(rg.guards, gs...)
	return rg
}
func (rg *RouterGroup) UseInterceptors(is ...NestInterceptor) *RouterGroup {
	rg.interceptors = append(rg.interceptors, is...)
	return rg
}
func (rg *RouterGroup) UsePipes(ps ...PipeTransform) *RouterGroup {
	rg.pipes = append(rg.pipes, ps...)
	return rg
}
func (rg *RouterGroup) UseFilters(fs ...ExceptionFilter) *RouterGroup {
	rg.filters = append(rg.filters, fs...)
	return rg
}

func (rg *RouterGroup) processError(c *gin.Context, err error, fs []ExceptionFilter) {
	// 严格按顺序执行：方法级 -> 组级 -> 全局(含内置兜底)
	for _, f := range fs {
		f.Catch(c, err)
		if c.IsAborted() {
			return
		}
	}
}

// ==========================================
// 4. 执行引擎 (Core Engine)
// ==========================================

func (rg *RouterGroup) Handle(method, path string, handler interface{}, methodEnhancers ...interface{}) {
	hVal := reflect.ValueOf(handler)
	hTyp := hVal.Type()

	// 1. 编译期提取增强器
	var mGuards []CanActivate
	var mInterceptors []NestInterceptor
	var mPipes []PipeTransform
	var mFilters []ExceptionFilter
	for _, e := range methodEnhancers {
		switch v := e.(type) {
		case CanActivate:
			mGuards = append(mGuards, v)
		case NestInterceptor:
			mInterceptors = append(mInterceptors, v)
		case PipeTransform:
			mPipes = append(mPipes, v)
		case ExceptionFilter:
			mFilters = append(mFilters, v)
		}
	}

	// 2. 预合并链条
	fGuards := concat(rg.app.globalGuards, rg.guards, mGuards)
	fInterceptors := concat(rg.app.globalInterceptors, rg.interceptors, mInterceptors)
	fPipes := concat(rg.app.globalPipes, rg.pipes, mPipes)
	fFilters := concat(mFilters, rg.filters, rg.app.globalFilters)

	// 3. 预设参数工厂
	factories := make([]argumentResolver, hTyp.NumIn())
	for i := 0; i < hTyp.NumIn(); i++ {
		factories[i] = rg.app.makeParamFactory(hTyp.In(i))
	}

	// 4. 运行时 Handler
	coreHandler := func(c *gin.Context) {
		// A. Panic 捕获与过滤器整合
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				rg.processError(c, err, fFilters)
			}
		}()

		// B. Guards 执行
		for _, g := range fGuards {
			if !g.CanActivate(c) {
				if !c.IsAborted() {
					c.AbortWithStatus(http.StatusForbidden)
				}
				return
			}
		}

		// C. 执行 Pipeline
		pipeline := func() interface{} {
			args := make([]reflect.Value, hTyp.NumIn())
			for i, factory := range factories {
				// A. 调用工厂获取原始值 (Context, DTO等)
				valRef, err := factory(c)
				if err != nil {
					return err
				}
				// 将 reflect.Value 转回 interface{} 交给 Pipe 处理
				var val interface{} = valRef.Interface()
				// B. 执行 Pipeline (Pipes 转换与校验)
				for _, p := range fPipes {
					val, err = p.Transform(val, hTyp.In(i))
					if err != nil {
						return err
					}
				}
				// C. 重新包装为 reflect.Value 准备反射调用
				args[i] = reflect.ValueOf(val)
			}

			// D. 执行真正的业务方法
			res := hVal.Call(args)
			if len(res) > 0 {
				return res[0].Interface()
			}
			return nil
		}

		// D. 拦截器洋葱模型执行
		result := rg.app.execInterceptors(c, fInterceptors, 0, pipeline)
		if err, ok := result.(error); ok {
			rg.processError(c, err, fFilters)
			return
		}
		rg.processResponse(c, result, fFilters)
	}

	rg.ginGroup.Handle(method, path, coreHandler)
}

func (rg *RouterGroup) GET(path string, h interface{}, m ...interface{}) {
	rg.Handle("GET", path, h, m...)
}
func (rg *RouterGroup) POST(path string, h interface{}, m ...interface{}) {
	rg.Handle("POST", path, h, m...)
}
func (rg *RouterGroup) PUT(path string, h interface{}, m ...interface{}) {
	rg.Handle("PUT", path, h, m...)
}
func (rg *RouterGroup) DELETE(path string, h interface{}, m ...interface{}) {
	rg.Handle("DELETE", path, h, m...)
}
func (rg *RouterGroup) OPTION(path string, h interface{}, m ...interface{}) {
	rg.Handle("OPTION", path, h, m...)
}
func (rg *RouterGroup) PATCH(path string, h interface{}, m ...interface{}) {
	rg.Handle("PATCH", path, h, m...)
}
func (rg *RouterGroup) TRACE(path string, h interface{}, m ...interface{}) {
	rg.TRACE("PATCH", path, h)
}

// SetMode 设置运行模式 (gin.ReleaseMode / gin.DebugMode)
func (app *GnestApp) SetMode(mode string) *GnestApp {
	gin.SetMode(mode)
	return app
}

// NoRoute 自定义 404 页面
func (app *GnestApp) NoRoute(h interface{}) *GnestApp {
	// 这里复用你的 Handle 逻辑即可，简单起见直接用 gin 原生
	if handler, ok := h.(gin.HandlerFunc); ok {
		app.Engine.NoRoute(handler)
	}
	return app
}

// SetHTMLTemplate 支持多模板引擎 (如果不用默认的 Glob)
func (app *GnestApp) SetHTMLTemplate(templ *template.Template) *GnestApp {
	app.Engine.SetHTMLTemplate(templ)
	return app
}

// Static 静态目录托管
func (app *GnestApp) Static(relativePaths string, root string) *GnestApp {
	app.Engine.Static(relativePaths, root)
	return app
}

// StaticFile 单个文件托管 (如 favicon.ico)
func (app *GnestApp) StaticFile(relativePath, filepath string) *GnestApp {
	app.Engine.StaticFile(relativePath, filepath)
	return app
}

// LoadHTMLGlob 配置模板路径
func (app *GnestApp) LoadHTMLGlob(pattern string) *GnestApp {
	app.Engine.LoadHTMLGlob(pattern)
	return app
}

func (app *GnestApp) GET(path string, handler interface{}, m ...interface{}) {
	// 借用根路径的 RouterGroup 来处理
	rg := &RouterGroup{app: app, ginGroup: &app.Engine.RouterGroup}
	rg.GET(path, handler, m...)
}

func (app *GnestApp) POST(path string, handler interface{}, m ...interface{}) {
	// 借用根路径的 RouterGroup 来处理
	rg := &RouterGroup{app: app, ginGroup: &app.Engine.RouterGroup}
	rg.POST(path, handler, m...)
}

func (app *GnestApp) PUT(path string, handler interface{}, m ...interface{}) {
	// 借用根路径的 RouterGroup 来处理
	rg := &RouterGroup{app: app, ginGroup: &app.Engine.RouterGroup}
	rg.PUT(path, handler, m...)
}

func (app *GnestApp) DELETE(path string, handler interface{}, m ...interface{}) {
	// 借用根路径的 RouterGroup 来处理
	rg := &RouterGroup{app: app, ginGroup: &app.Engine.RouterGroup}
	rg.DELETE(path, handler, m...)
}

func (app *GnestApp) OPTION(path string, handler interface{}, m ...interface{}) {
	// 借用根路径的 RouterGroup 来处理
	rg := &RouterGroup{app: app, ginGroup: &app.Engine.RouterGroup}
	rg.OPTION(path, handler, m...)
}

func (app *GnestApp) PATCH(path string, handler interface{}, m ...interface{}) {
	// 借用根路径的 RouterGroup 来处理
	rg := &RouterGroup{app: app, ginGroup: &app.Engine.RouterGroup}
	rg.PATCH(path, handler, m...)
}

func (app *GnestApp) TRACE(path string, handler interface{}, m ...interface{}) {
	// 借用根路径的 RouterGroup 来处理
	rg := &RouterGroup{app: app, ginGroup: &app.Engine.RouterGroup}
	rg.TRACE(path, handler, m...)
}

// ==========================================
// 5. 参数绑定与底层支持 (Underlying Support)
// ==========================================

func (app *GnestApp) makeParamFactory(t reflect.Type) argumentResolver {
	// --- 1. 基础类型处理 (Context, Req, Res) ---
	switch t.String() {
	case "*gin.Context":
		return func(c *gin.Context) (reflect.Value, error) { return reflect.ValueOf(c), nil }
	case "*http.Request":
		return func(c *gin.Context) (reflect.Value, error) { return reflect.ValueOf(c.Request), nil }
	}

	// --- 2. 文件处理 (@UploadedFile) ---
	if t.String() == "*multipart.FileHeader" {
		return func(c *gin.Context) (reflect.Value, error) {
			f, err := c.FormFile("file") // 这里可进一步优化为根据参数名取
			return reflect.ValueOf(f), err
		}
	}
	if t.String() == "[]*multipart.FileHeader" {
		return func(c *gin.Context) (reflect.Value, error) {
			form, err := c.MultipartForm()
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(form.File["files"]), nil
		}
	}

	// --- 3. 自定义装饰器处理 (如 @User) ---
	if factory, ok := app.customDecorators[t]; ok {
		return func(c *gin.Context) (reflect.Value, error) {
			return reflect.ValueOf(factory(c)), nil
		}
	}

	// --- 4. 依赖注入处理 (Provider) ---
	if p, ok := app.providers[t]; ok {
		return func(c *gin.Context) (reflect.Value, error) { return p, nil }
	}

	// --- 5. 核心：DTO 结构体智能绑定 (Body/Query/Param/Header) ---
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		return app.createStructResolver(t.Elem())
	}

	return func(c *gin.Context) (reflect.Value, error) {
		return reflect.Value{}, fmt.Errorf("unsupported type: %v", t)
	}
}

// 针对结构体 DTO 的预分析优化
func (app *GnestApp) createStructResolver(st reflect.Type) argumentResolver {
	// 在路由注册阶段，先扫描 Tag，决定运行时调用哪些绑定方法
	hasUriTag := false
	hasHeaderTag := false
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if field.Tag.Get("uri") != "" {
			hasUriTag = true
		}
		if field.Tag.Get("header") != "" {
			hasHeaderTag = true
		}
	}

	return func(c *gin.Context) (reflect.Value, error) {
		obj := reflect.New(st).Interface()

		// 只有存在相关 Tag 时才调用对应的绑定器，减少性能损耗
		if hasUriTag {
			_ = c.ShouldBindUri(obj)
		}
		if hasHeaderTag {
			_ = c.ShouldBindHeader(obj)
		}

		// 默认绑定 Body 或 Query
		if err := c.ShouldBind(obj); err != nil {
			return reflect.Value{}, err
		}

		// 自动校验
		if err := app.validate.Struct(obj); err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(obj), nil
	}
}

func (app *GnestApp) execInterceptors(c *gin.Context, is []NestInterceptor, i int, next func() interface{}) interface{} {
	if i >= len(is) {
		return next()
	}
	return is[i].Intercept(c, func() interface{} { return app.execInterceptors(c, is, i+1, next) })
}

func (rg *RouterGroup) processResponse(c *gin.Context, res interface{}, fs []ExceptionFilter) {
	if c.IsAborted() || c.Writer.Written() {
		return
	}

	if err, ok := res.(error); ok {
		rg.processError(c, err, fs)
		return
	}

	// 根据返回值的类型，自动映射 Gin 的响应方法
	switch v := res.(type) {
	case Render: // 上一步增加的 HTML 渲染
		c.HTML(http.StatusOK, v.Name, v.Data)
	case RedirectResult:
		code := v.Code
		if code == 0 {
			code = http.StatusFound
		}
		c.Redirect(code, v.Location)
	case DataResult:
		c.Data(http.StatusOK, v.ContentType, v.Data)
	case FileResult:
		if v.FileName != "" {
			c.FileAttachment(v.FilePath, v.FileName)
		} else {
			c.File(v.FilePath)
		}
	case string: // 返回纯字符串
		c.String(http.StatusOK, v)
	case []byte: // 返回原始字节
		c.Data(http.StatusOK, "application/octet-stream", v)
	case nil:
		return
	default:
		// 默认返回 JSON
		c.JSON(http.StatusOK, res)
	}
}

// 生命周期逻辑
func (app *GnestApp) ListenAndServe(addr string) {
	app.callHook("OnModuleInit")
	app.callHook("OnApplicationBootstrap")
	srv := &http.Server{Addr: addr, Handler: app.Engine}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()
	log.Printf("[Gnest] Server started on %s", addr)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	sig := <-stop
	// 3. 销毁序列
	app.callHook("OnModuleDestroy") // 补齐这个！
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	app.callHook("BeforeApplicationShutdown", sig.String())
	srv.Shutdown(ctx)
	app.callHook("OnApplicationShutdown")
	log.Println("[Gnest] Shutdown finished")
}

func (app *GnestApp) callHook(name string, args ...interface{}) {
	for _, inst := range app.container {
		v := reflect.ValueOf(inst)
		if m := v.MethodByName(name); m.IsValid() {
			in := make([]reflect.Value, len(args))
			for i, a := range args {
				in[i] = reflect.ValueOf(a)
			}
			m.Call(in)
		}
	}
}

func concat[T any](ss ...[]T) []T {
	var r []T
	for _, s := range ss {
		r = append(r, s...)
	}
	return r
}
