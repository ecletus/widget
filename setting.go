package widget

import (
	"fmt"
	"time"

	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena/template/html/template"
	"github.com/ecletus/admin"
	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/ecletus/serializable_meta"
)

// QorWidgetSettingInterface qor widget setting interface
type QorWidgetSettingInterface interface {
	GetWidgetName() string
	SetWidgetName(string)
	GetGroupName() string
	SetGroupName(string)
	GetScope() string
	SetScope(string)
	GetEnabled() bool
	SetEnabled(bool)
	GetTemplate() string
	SetTemplate(string)
	GetSourceType() string
	SetSourceType(string)
	GetSourceID() string
	SetSourceID(string)
	GetShared() bool
	SetShared(bool)
	serializable_meta.SerializableMetaInterface
}

// QorWidgetSetting default qor widget setting struct
type QorWidgetSetting struct {
	Name        string `gorm:"primary_key"`
	Scope       string `gorm:"primary_key;size:128;default:'default'"`
	SourceType  string `gorm:"primary_key;default:''"`
	SourceID    string `gorm:"primary_key;default:''"`
	Description string
	Shared      bool
	WidgetType  string
	GroupName   string
	Template    string
	serializable_meta.SerializableMeta
	CreatedAt time.Time
	UpdatedAt time.Time
	Enabled   bool
}

// ResourceName get widget setting's resource name
func (widgetSetting *QorWidgetSetting) ResourceName() string {
	return "Widget Content"
}

// GetSerializableArgumentKind get serializable kind
func (widgetSetting *QorWidgetSetting) GetSerializableArgumentKind() string {
	if widgetSetting.WidgetType != "" {
		return widgetSetting.WidgetType
	}
	return widgetSetting.Kind
}

// SetSerializableArgumentKind set serializable kind
func (widgetSetting *QorWidgetSetting) SetSerializableArgumentKind(name string) {
	widgetSetting.WidgetType = name
	widgetSetting.Kind = name
}

// GetWidgetName get widget setting's group name
func (widgetSetting QorWidgetSetting) GetWidgetName() string {
	return widgetSetting.Name
}

// SetWidgetName set widget setting's group name
func (widgetSetting *QorWidgetSetting) SetWidgetName(name string) {
	widgetSetting.Name = name
}

// GetGroupName get widget setting's group name
func (widgetSetting QorWidgetSetting) GetGroupName() string {
	return widgetSetting.GroupName
}

// SetGroupName set widget setting's group name
func (widgetSetting *QorWidgetSetting) SetGroupName(groupName string) {
	widgetSetting.GroupName = groupName
}

// GetScope get widget's scope
func (widgetSetting QorWidgetSetting) GetScope() string {
	return widgetSetting.Scope
}

// SetScope set widget setting's scope
func (widgetSetting *QorWidgetSetting) SetScope(scope string) {
	widgetSetting.Scope = scope
}

// GetSourceType get widget's source type
func (widgetSetting QorWidgetSetting) GetSourceType() string {
	return widgetSetting.SourceType
}

// SetSourceType set widget setting's souce type
func (widgetSetting *QorWidgetSetting) SetSourceType(sourceType string) {
	widgetSetting.SourceType = sourceType
}

// GetSourceID get widget's source ID
func (widgetSetting QorWidgetSetting) GetSourceID() string {
	return widgetSetting.SourceID
}

// SetSourceID set widget setting's source id
func (widgetSetting *QorWidgetSetting) SetSourceID(sourceID string) {
	widgetSetting.SourceID = sourceID
}

// GetShared get widget's source ID
func (widgetSetting QorWidgetSetting) GetShared() bool {
	return widgetSetting.Shared
}

// SetShared set widget setting's source id
func (widgetSetting *QorWidgetSetting) SetShared(shared bool) {
	widgetSetting.Shared = shared
}

// GetShared get widget's source ID
func (widgetSetting QorWidgetSetting) GetEnabled() bool {
	return widgetSetting.Enabled
}

// SetShared set widget setting's source id
func (widgetSetting *QorWidgetSetting) SetEnabled(enabled bool) {
	widgetSetting.Enabled = enabled
}

// GetTemplate get used widget template
func (widgetSetting QorWidgetSetting) GetTemplate() string {
	if widget := GetWidget(widgetSetting.GetSerializableArgumentKind()); widget != nil {
		for _, value := range widget.Templates {
			if value == widgetSetting.Template {
				return value
			}
		}

		// return first value of defined widget templates
		for _, value := range widget.Templates {
			return value
		}
	}
	return ""
}

// SetTemplate set used widget's template
func (widgetSetting *QorWidgetSetting) SetTemplate(template string) {
	widgetSetting.Template = template
}

// GetSerializableArgumentResource get setting's argument's resource
func (widgetSetting *QorWidgetSetting) GetSerializableArgumentResource() *admin.Resource {
	widget := GetWidget(widgetSetting.GetSerializableArgumentKind())
	if widget != nil {
		return widget.Setting
	}
	return nil
}

// ConfigureResource a method used to config Widget for qor admin
func (widgetSetting *QorWidgetSetting) ConfigureResource(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		res.Meta(&admin.Meta{Name: "PreviewIcon", Valuer: func(result interface{}, context *core.Context) interface{} {
			if setting, ok := result.(QorWidgetSettingInterface); ok {
				if widget := GetWidget(setting.GetSerializableArgumentKind()); widget != nil {
					return template.HTML(fmt.Sprintf("<img class='qor-preview-icon' src='%v'/>", widget.PreviewIcon))
				}
			}
			return ""
		}})

		res.Meta(&admin.Meta{Name: "Name", Type: "string"})
		res.Meta(&admin.Meta{Name: "DisplayName", Label: "Name", Type: "readonly", FieldName: "Name"})
		res.Meta(&admin.Meta{Name: "Description", Type: "string"})

		res.Meta(&admin.Meta{
			Name: "Scope",
			Type: "hidden",
			Valuer: func(result interface{}, context *core.Context) interface{} {
				if scope := context.Request.URL.Query().Get("widget_scope"); scope != "" {
					return scope
				}

				if setting, ok := result.(QorWidgetSettingInterface); ok {
					if scope := setting.GetScope(); scope != "" {
						return scope
					}
				}

				return "default"
			},
			Setter: func(result interface{}, metaValue *resource.MetaValue, context *core.Context) error {
				if setting, ok := result.(QorWidgetSettingInterface); ok {
					setting.SetScope(utils.ToString(metaValue.Value))
				}
				return nil
			},
		})

		res.Meta(&admin.Meta{
			Name: "SourceType",
			Type: "hidden",
			Valuer: func(result interface{}, context *core.Context) interface{} {
				if sourceType := context.Request.URL.Query().Get("widget_source_type"); sourceType != "" {
					return sourceType
				}

				if setting, ok := result.(QorWidgetSettingInterface); ok {
					if sourceType := setting.GetSourceType(); sourceType != "" {
						return sourceType
					}
				}
				return ""
			},
			Setter: func(result interface{}, metaValue *resource.MetaValue, context *core.Context) error {
				if setting, ok := result.(QorWidgetSettingInterface); ok {
					setting.SetSourceType(utils.ToString(metaValue.Value))
				}
				return nil
			},
		})

		res.Meta(&admin.Meta{
			Name: "SourceID",
			Type: "hidden",
			Valuer: func(result interface{}, context *core.Context) interface{} {
				if sourceID := context.Request.URL.Query().Get("widget_source_id"); sourceID != "" {
					return sourceID
				}

				if setting, ok := result.(QorWidgetSettingInterface); ok {
					if sourceID := setting.GetSourceID(); sourceID != "" {
						return sourceID
					}
				}
				return ""
			},
			Setter: func(result interface{}, metaValue *resource.MetaValue, context *core.Context) error {
				if setting, ok := result.(QorWidgetSettingInterface); ok {
					setting.SetSourceID(utils.ToString(metaValue.Value))
				}
				return nil
			},
		})

		res.Meta(&admin.Meta{
			Name: "Widgets",
			Type: "select_one",
			Valuer: func(result interface{}, context *core.Context) interface{} {
				if typ := context.Request.URL.Query().Get("widget_type"); typ != "" {
					return typ
				}

				if setting, ok := result.(QorWidgetSettingInterface); ok {
					widget := GetWidget(setting.GetSerializableArgumentKind())
					if widget == nil {
						return ""
					}
					return widget.Name
				}

				return ""
			},
			Collection: func(result interface{}, context *core.Context) (results [][]string) {
				if setting, ok := result.(QorWidgetSettingInterface); ok {
					if setting.GetWidgetName() == "" {
						for _, widget := range registeredWidgets {
							results = append(results, []string{widget.Name, widget.Name})
						}
					} else {
						groupName := setting.GetGroupName()
						for _, group := range registeredWidgetsGroup {
							if group.Name == groupName {
								for _, widget := range group.Widgets {
									results = append(results, []string{widget, widget})
								}
							}
						}
					}

					if len(results) == 0 {
						results = append(results, []string{setting.GetSerializableArgumentKind(), setting.GetSerializableArgumentKind()})
					}
				}
				return
			},
			Setter: func(result interface{}, metaValue *resource.MetaValue, context *core.Context) error {
				if setting, ok := result.(QorWidgetSettingInterface); ok {
					setting.SetSerializableArgumentKind(utils.ToString(metaValue.Value))
				}
				return nil
			},
		})

		res.Meta(&admin.Meta{
			Name: "Template",
			Type: "select_one",
			Valuer: func(result interface{}, context *core.Context) interface{} {
				if setting, ok := result.(QorWidgetSettingInterface); ok {
					return setting.GetTemplate()
				}
				return ""
			},
			Collection: func(result interface{}, context *core.Context) (results [][]string) {
				if setting, ok := result.(QorWidgetSettingInterface); ok {
					if widget := GetWidget(setting.GetSerializableArgumentKind()); widget != nil {
						for _, value := range widget.Templates {
							results = append(results, []string{value, value})
						}
					}
				}
				return
			},
			Setter: func(result interface{}, metaValue *resource.MetaValue, context *core.Context) error {
				if setting, ok := result.(QorWidgetSettingInterface); ok {
					setting.SetTemplate(utils.ToString(metaValue.Value))
				}
				return nil
			},
		})

		res.Meta(&admin.Meta{
			Name:  "Shared",
			Label: "Add to Container Library (can be reused on other pages)",
		})

		res.Action(&admin.Action{
			Name: "Preview",
			URL: func(record interface{}, context *admin.Context, args ...interface{}) string {
				return fmt.Sprintf("%v/%v/!preview", res.GetLink(record, context.Context, args...), record.(QorWidgetSettingInterface).GetWidgetName())
			},
			Modes: []string{"edit", "menu_item"},
		})

		res.AddProcessor(func(value interface{}, metaValues *resource.MetaValues, context *core.Context) error {
			if widgetSetting, ok := value.(QorWidgetSettingInterface); ok {
				if widgetSetting.GetShared() {
					widgetSetting.SetSourceType("")
					widgetSetting.SetSourceID("")
				}
			}
			return nil
		})

		res.UseTheme("widget")

		res.IndexAttrs("PreviewIcon", "Name", "Description", "CreatedAt", "UpdatedAt", "Enabled")
		res.ShowAttrs("PreviewIcon", "Name", "Scope", "WidgetType", "Template", "Description", "Value", "CreatedAt", "UpdatedAt", false)
		res.EditAttrs(
			"DisplayName", "Description", "Scope", "Widgets", "Template",
			&admin.Section{
				Title: "Settings",
				Rows:  [][]string{{"Kind"}, {"SerializableMeta"}},
			},
			"Shared", "SourceType", "SourceID",
			"Enabled",
		)
		res.NewAttrs("Name", "Description", "Scope", "Widgets", "Template",
			&admin.Section{
				Title: "Settings",
				Rows:  [][]string{{"Kind"}, {"SerializableMeta"}},
			},
			"Shared", "SourceType", "SourceID",
			"Enabled",
		)

		searchHandler := res.SearchHandler
		res.SearchHandler = func(searcher *admin.Searcher, db *aorm.DB, keyword string) (_ *aorm.DB, err error) {
			// don't include widgets have source_type in index page
			if searcher.ResourceID == nil {
				db = db.Where("source_type = ?", "")
			}
			return searchHandler(searcher, db, keyword)
		}
	}
}
