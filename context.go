package widget

import (
	"fmt"
	"reflect"

	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena/template/html/template"
	"github.com/ecletus/admin"
	"github.com/ecletus/core"
	"github.com/ecletus/core/utils"
)

// Context widget context
type Context struct {
	Context          *core.Context
	Widgets          *Widgets
	AvailableWidgets []string
	Options          map[string]interface{}
	InlineEdit       bool
	SourceType       string
	SourceID         string
	FuncMaps         template.FuncMap
	WidgetSetting    QorWidgetSettingInterface
}

// Get get option with name
func (context Context) Get(name string) (interface{}, bool) {
	if value, ok := context.Options[name]; ok {
		return value, true
	}

	return nil, false
}

// Set set option by name
func (context *Context) Set(name string, value interface{}) {
	if context.Options == nil {
		context.Options = map[string]interface{}{}
	}
	context.Options[name] = value
}

// GetDB set option by name
func (context *Context) GetDB() *aorm.DB {
	return context.Context.DB()
}

// Clone clone a context
func (context *Context) Clone() *Context {
	clone := *context
	return &clone
}

// Render render widget based on context
func (context *Context) Render(widgetName string, widgetGroupName string) template.HTML {
	return context.RenderWidget(widgetName, widgetGroupName, true)
}

// Render render widget based on context
func (context *Context) RenderWidget(widgetName string, widgetGroupName string, enabled bool) template.HTML {
	var (
		visibleScopes         []string
		widgets               = context.Widgets
		widgetSettingResource = widgets.WidgetSettingResource
		clone                 = context.Clone()
	)

	for _, scope := range registeredScopes {
		if scope.Visible(context) {
			visibleScopes = append(visibleScopes, scope.ToParam())
		}
	}

	if setting := context.findWidgetSetting(widgetName, append(visibleScopes, "default"), widgetGroupName); setting != nil && (!enabled || setting.GetEnabled()) {
		clone.WidgetSetting = setting
		adminContext := admin.Context{Admin: context.Widgets.Config.Admin, Context: clone.Context}

		var (
			widgetObj     = GetWidget(setting.GetSerializableArgumentKind())
			widgetSetting = widgetObj.Context(clone, setting.GetSerializableArgument(setting))
		)

		if clone.InlineEdit {
			prefix := adminContext.JoinStaticURL()
			inlineEditURL := adminContext.URLFor(setting, widgetSettingResource)
			if widgetObj.InlineEditURL != nil {
				inlineEditURL = widgetObj.InlineEditURL(context)
			}

			return template.HTML(fmt.Sprintf(
				"<script data-prefix=\"%v\" src=\"%v/javascripts/widget_check.js?theme=widget\"></script><div class=\"qor-widget qor-widget-%v\" data-widget-inline-edit-url=\"%v\" data-url=\"%v\">\n%v\n</div>",
				prefix,
				prefix,
				utils.ToParamString(widgetObj.Name),
				fmt.Sprintf("%v/%v/inline-edit", prefix, widgets.Resource.ToParam()),
				inlineEditURL,
				widgetObj.Render(widgetSetting, setting.GetTemplate()),
			))
		}

		return widgetObj.Render(widgetSetting, setting.GetTemplate())
	}

	return template.HTML("")
}

func (context *Context) findWidgetSetting(widgetName string, scopes []string, widgetGroupName string) QorWidgetSettingInterface {
	var (
		db                    = context.GetDB()
		widgetSettingResource = context.Widgets.WidgetSettingResource
		setting               QorWidgetSettingInterface
		settings              = widgetSettingResource.NewSlice()
	)

	if context.SourceID != "" {
		db.Order("source_id DESC").Where("name = ? AND scope IN (?) AND ((shared = ? AND source_type = ?) OR (source_type = ? AND source_id = ?))", widgetName, scopes, true, "", context.SourceType, context.SourceID).Find(settings)
	} else {
		db.Where("name = ? AND scope IN (?) AND source_type = ?", widgetName, scopes, "").Find(settings)
	}

	settingsValue := reflect.Indirect(reflect.ValueOf(settings))
	if settingsValue.Len() > 0 {
	OUTTER:
		for _, scope := range scopes {
			for i := 0; i < settingsValue.Len(); i++ {
				s := settingsValue.Index(i).Interface().(QorWidgetSettingInterface)
				if s.GetScope() == scope {
					setting = s
					break OUTTER
				}
			}
		}
	}

	if context.SourceType == "" {
		if setting == nil {
			if widgetGroupName == "" {
				panic(fmt.Errorf("Widget: Can't Create Widget Without Widget Type"))
				return nil
			}
			setting = widgetSettingResource.NewStruct(context.Context.Site).(QorWidgetSettingInterface)
			setting.SetWidgetName(widgetName)
			setting.SetGroupName(widgetGroupName)
			setting.SetSerializableArgumentKind(widgetGroupName)
			db.Create(setting)
		} else if setting.GetGroupName() != widgetGroupName {
			setting.SetGroupName(widgetGroupName)
			db.Save(setting)
		}
	}

	return setting
}
