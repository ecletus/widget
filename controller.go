package widget

import (
	"net/http"

	"github.com/ecletus/admin"
	"github.com/ecletus/responder"
	"github.com/ecletus/serializable_meta"
	"github.com/moisespsena/template/html/template"
)

type widgetController struct {
	Widgets *Widgets
}

func (wc widgetController) Index(context *admin.Context) {
	context = context.NewResourceContext(wc.Widgets.WidgetSettingResource)
	context.DefaulLayout("index")

	result, _, err := wc.getWidget(context)
	context.AddError(err)

	if context.HasError() {
		http.NotFound(context.Writer, context.Request)
	} else {
		responder.With("html", func() {
			context.Execute("index", result)
		}).With("json", func() {
			context.JSON(result)
		}).Respond(context.Request)
	}
}

func (wc widgetController) New(context *admin.Context) {
	widgetInter := wc.Widgets.WidgetSettingResource.NewStruct(context.Site).(QorWidgetSettingInterface)
	context.Execute("new", widgetInter)
}

func (wc widgetController) Setting(context *admin.Context) {
	widgetInter := wc.Widgets.WidgetSettingResource.NewStruct(context.Site).(QorWidgetSettingInterface)
	widgetType := context.Request.URL.Query().Get("widget_type")
	if widgetType != "" {
		if serializableMeta, ok := widgetInter.(serializable_meta.SerializableMetaInterface); ok && serializableMeta.GetSerializableArgumentKind() != widgetType {
			serializableMeta.SetSerializableArgumentKind(widgetType)
			serializableMeta.SetSerializableArgumentValue(nil)
		}
	}
	section := []*admin.Section{{
		Resource: wc.Widgets.WidgetSettingResource,
		Title:    "Settings",
		Rows:     [][]string{{"Kind"}, {"SerializableMeta"}},
	}}
	content := context.RenderHtml("setting", struct {
		Widget  interface{}
		Section []*admin.Section
	}{
		Widget:  widgetInter,
		Section: section,
	})
	context.Writer.Write([]byte(content))
}

func (wc widgetController) Edit(context *admin.Context) {
	context = context.NewResourceContext(wc.Widgets.WidgetSettingResource)
	widgetSetting, scopes, err := wc.getWidget(context)
	context.AddError(err)

	responder.With("html", func() {
		context.Funcs(template.FuncMap{
			"get_widget_scopes": func() []string { return scopes },
		}).Execute("edit", widgetSetting)
	}).With("json", func() {
		context.Layout = "show"
		context.JSON(widgetSetting)
	}).Respond(context.Request)
}

func (wc widgetController) Preview(context *admin.Context) {
	widgetContext := wc.Widgets.NewContext(context.Context, &Context{
		Options: map[string]interface{}{"Request": context.Request, "AdminContext": context},
	})

	content := context.Funcs(template.FuncMap{
		"load_preview_assets": wc.Widgets.LoadPreviewAssets,
	}).Funcs(widgetContext.FuncMap()).RenderHtml("preview", struct {
		WidgetName string
	}{
		WidgetName: context.ResourceID.String(),
	})
	context.Writer.Write([]byte(content))
}

func (wc widgetController) Update(context *admin.Context) {
	context = context.NewResourceContext(wc.Widgets.WidgetSettingResource)
	widgetSetting, scopes, err := wc.getWidget(context)
	context.AddError(err)

	if context.AddError(context.Resource.Decode(context.Context, widgetSetting)); !context.HasError() {
		context.AddError(context.Resource.Crud(context.Context).Update(widgetSetting))
	}

	if context.HasError() {
		responder.With("html", func() {
			context.Writer.WriteHeader(admin.HTTPUnprocessableEntity)
			context.Funcs(template.FuncMap{
				"get_widget_scopes": func() []string { return scopes },
			}).Execute("edit", widgetSetting)
		}).With([]string{"json", "xml"}, func() {
			context.Writer.WriteHeader(admin.HTTPUnprocessableEntity)
			context.Encode(map[string]interface{}{"errors": context.GetErrors()}, "index")
		}).Respond(context.Request)
	} else {
		responder.With("html", func() {
			http.Redirect(context.Writer, context.Request, context.Request.URL.Path, http.StatusFound)
		}).With("json", func() {
			context.JSON(widgetSetting, "index")
		}).Respond(context.Request)
	}
}

func (wc widgetController) InlineEdit(context *admin.Context) {
	context.Include(context.Writer, "widget/inline_edit")
}

func (wc widgetController) getWidget(context *admin.Context) (interface{}, []string, error) {
	var DB = context.DB()

	// index page
	if context.ResourceID == nil && context.Request.Method == "GET" {
		scope := context.Request.URL.Query().Get("widget_scope")
		if scope == "" {
			scope = "default"
		}

		context.SetDB(DB.Where("scope = ?", scope))
		defer context.SetDB(DB)
		results, err := context.FindMany()
		return results, []string{}, err
	}

	// show page
	var (
		scopes     []string
		result     = wc.Widgets.WidgetSettingResource.NewStruct(context.Site)
		scope      = context.Request.URL.Query().Get("widget_scope")
		widgetType = context.Request.URL.Query().Get("widget_type")
	)

	if scope == "" {
		scope = context.Request.Form.Get("QorResource.Scope")
		if scope == "" {
			scope = "default"
		}
	}

	if widgetType == "" {
		widgetType = context.Request.Form.Get("QorResource.Kind")
	}

	err := DB.FirstOrInit(result, QorWidgetSetting{Name: context.ResourceID.String(), Scope: scope}).Error

	if widgetType != "" {
		if serializableMeta, ok := result.(serializable_meta.SerializableMetaInterface); ok && serializableMeta.GetSerializableArgumentKind() != widgetType {
			serializableMeta.SetSerializableArgumentKind(widgetType)
			serializableMeta.SetSerializableArgumentValue(nil)
		}
	}
	return result, scopes, err
}
