package widget

import (
	"html/template"

	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena-go/i18n-modular/i18nmod"
	"github.com/moisespsena-go/path-helpers"
	"github.com/ecletus/admin"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/roles"
)

const FS_NAME = "widgets"

var (
	PKG                    = path_helpers.GetCalledDir()
	I18NGROUP              = i18nmod.PkgToGroup(PKG)
	registeredWidgets      []*Widget
	registeredWidgetsGroup []*WidgetsGroup
)

// Config widget config
type Config struct {
	Admin         *admin.Admin
	PreviewAssets []string
	AssetFS       assetfs.Interface
	RootAssetFS   assetfs.Interface
}

// New new widgets container
func New(config *Config) *Widgets {
	AssetFS := config.AssetFS

	if AssetFS == nil {
		AssetFS = config.RootAssetFS.NameSpace(FS_NAME)
	}

	widgets := &Widgets{Config: config, funcMaps: template.FuncMap{}, AssetFS: AssetFS}
	return widgets
}

// Widgets widgets container
type Widgets struct {
	funcMaps              template.FuncMap
	Config                *Config
	Resource              *admin.Resource
	AssetFS               assetfs.Interface
	WidgetSettingResource *admin.Resource
}

// RegisterWidget register a new widget
func (widgets *Widgets) RegisterWidget(w *Widget) {
	registeredWidgets = append(registeredWidgets, w)
}

// RegisterWidgetsGroup register widgets group
func (widgets *Widgets) RegisterWidgetsGroup(group *WidgetsGroup) {
	registeredWidgetsGroup = append(registeredWidgetsGroup, group)
}

// RegisterFuncMap register view funcs, it could be used when render templates
func (widgets *Widgets) RegisterFuncMap(name string, fc interface{}) {
	widgets.funcMaps[name] = fc
}

// ConfigureQorResourceBeforeInitialize a method used to config Widget for qor admin
func (widgets *Widgets) ConfigureResourceBeforeInitialize(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		// set resources
		widgets.Resource = res

		// set setting resource
		if widgets.WidgetSettingResource == nil {
			widgets.WidgetSettingResource = res.GetAdmin().NewResource(&QorWidgetSetting{}, &admin.Config{Name: res.Name})
		}

		res.Name = widgets.WidgetSettingResource.Name

		for funcName, fc := range funcMap {
			res.GetAdmin().RegisterFuncMap(funcName, fc)
		}

		// configure routes
		controller := widgetController{Widgets: widgets}
		router := widgets.WidgetSettingResource.Router
		orouter := widgets.WidgetSettingResource.ItemRouter
		router.Get("/", admin.NewHandler(controller.Index, &admin.RouteConfig{Resource: widgets.WidgetSettingResource}))
		router.Get("/new", admin.NewHandler(controller.New, &admin.RouteConfig{Resource: widgets.WidgetSettingResource}))
		router.Get("/!setting", admin.NewHandler(controller.Setting, &admin.RouteConfig{Resource: widgets.WidgetSettingResource}))
		orouter.Get("/", admin.NewHandler(controller.Edit, &admin.RouteConfig{Resource: widgets.WidgetSettingResource}))
		orouter.Get("/!preview", admin.NewHandler(controller.Preview, &admin.RouteConfig{Resource: widgets.WidgetSettingResource}))
		orouter.Get("/edit", admin.NewHandler(controller.Edit, &admin.RouteConfig{Resource: widgets.WidgetSettingResource}))
		orouter.Put("/", admin.NewHandler(controller.Update, &admin.RouteConfig{Resource: widgets.WidgetSettingResource}))
		router.Post("/", admin.NewHandler(controller.Update, &admin.RouteConfig{Resource: widgets.WidgetSettingResource}))
		router.Get("/inline-edit", admin.NewHandler(controller.InlineEdit, &admin.RouteConfig{Resource: widgets.WidgetSettingResource}))
	}
}

// Widget widget struct
type Widget struct {
	Name          string
	PreviewIcon   string
	Group         string
	Templates     []string
	Setting       *admin.Resource
	Permission    *roles.Permission
	InlineEditURL func(*Context) string
	Context       func(context *Context, setting interface{}) *Context
}

// WidgetsGroup widgets Group
type WidgetsGroup struct {
	Name    string
	Widgets []string
}

// GetWidget get widget by name
func GetWidget(name string) *Widget {
	for _, w := range registeredWidgets {
		if w.Name == name {
			return w
		}
	}

	for _, g := range registeredWidgetsGroup {
		if g.Name == name {
			for _, widgetName := range g.Widgets {
				return GetWidget(widgetName)
			}
		}
	}
	return nil
}
