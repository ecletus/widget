package widget

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/aghape/core"
	"github.com/aghape/core/utils"
	"github.com/moisespsena/go-assetfs"
	"github.com/moisespsena/go-assetfs/assetfsapi"
	"github.com/moisespsena/template/html/template"
)

// Render find widget by name, render it based on current context
func (widgets *Widgets) Render(context *core.Context, widgetName string, widgetGroupName string) template.HTML {
	return widgets.NewContext(context, nil).Render(widgetName, widgetGroupName)
}

// NewContext create new context for widgets
func (widgets *Widgets) NewContext(qorContext *core.Context, context *Context) *Context {
	if context == nil {
		context = &Context{}
	}

	if context.Context == nil {
		context.Context = qorContext
	}

	if context.Options == nil {
		context.Options = map[string]interface{}{}
	}

	if context.FuncMaps == nil {
		context.FuncMaps = template.FuncMap{}
	}

	for key, fc := range widgets.funcMaps {
		if _, ok := context.FuncMaps[key]; !ok {
			context.FuncMaps[key] = fc
		}
	}

	context.Widgets = widgets
	return context
}

// Funcs return view functions map
func (context *Context) Funcs(funcMaps template.FuncMap) *Context {
	if context.FuncMaps == nil {
		context.FuncMaps = template.FuncMap{}
	}

	for key, fc := range funcMaps {
		context.FuncMaps[key] = fc
	}

	return context
}

// FuncMap return funcmap
func (context *Context) FuncMap() template.FuncMap {
	funcMap := template.FuncMap{}

	funcMap["render_widget"] = func(widgetName string, widgetGroupName ...string) template.HTML {
		var groupName string
		if len(widgetGroupName) == 0 {
			groupName = ""
		} else {
			groupName = widgetGroupName[0]
		}
		return context.RenderWidget(widgetName, groupName, false)
	}

	funcMap["render_enabled_widget"] = func(widgetName string, widgetGroupName ...string) template.HTML {
		var groupName string
		if len(widgetGroupName) == 0 {
			groupName = ""
		} else {
			groupName = widgetGroupName[0]
		}
		return context.Render(widgetName, groupName)
	}

	return funcMap
}

// Render register widget itself content
func (w *Widget) Render(context *Context, file string) template.HTML {
	var (
		err   error
		asset assetfs.AssetInterface
		tmpl  *template.Template
	)

	if file == "" {
		file = w.Templates[0]
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Get error when render file %v: %v", file, r)
			utils.ExitWithMsg(err)
		}
	}()

	if asset, err = context.Widgets.AssetFS.Asset(file + ".tmpl"); err == nil {
		if tmpl, err = template.New(filepath.Base(file)).SetPath(asset.GetName()).Parse(asset.GetString()); err == nil {
			var result = bytes.NewBufferString("")
			if err = tmpl.Execute(result, context.Options, context.FuncMaps); err == nil {
				return template.HTML(result.String())
			}
		}
	}

	return template.HTML(err.Error())
}

// RegisterViewPath register views directory
func (widgets *Widgets) RegisterViewPath(p string) {
	widgets.AssetFS.(assetfsapi.PathRegistrator).RegisterPath(p)
}

// LoadPreviewAssets will return assets tag used for Preview
func (widgets *Widgets) LoadPreviewAssets() template.HTML {
	tags := ""
	for _, asset := range widgets.Config.PreviewAssets {
		extension := filepath.Ext(asset)
		if extension == ".css" {
			tags += fmt.Sprintf("<link rel=\"stylesheet\" type=\"text/css\" href=\"%v\">\n", asset)
		} else if extension == ".js" {
			tags += fmt.Sprintf("<script src=\"%v\"></script>\n", asset)
		} else {
			tags += fmt.Sprintf("%v\n", asset)
		}
	}
	return template.HTML(tags)
}
