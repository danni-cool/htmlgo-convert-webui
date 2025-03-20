package parser_utils

// TagMapping contains information about HTML tag mappings for Go to HTML conversion
type TagMapping struct {
	TagName    string            // HTML tag name
	GoFuncName string            // Go function name in htmlgo
	Attributes map[string]string // Mapping of HTML attribute names to Go method names
	SelfClose  bool              // Whether the tag is self-closing
	GoPackage  string            // Go package name (default: htmlgo, could be "h" etc.)
}

// HTML Standard Tags - based on html2go and htmlgo packages
var StandardTagMappings = map[string]TagMapping{
	"div": {
		TagName:    "div",
		GoFuncName: "Div",
		Attributes: map[string]string{
			"class":     "Class",
			"id":        "ID",
			"style":     "Style",
			"data-*":    "Attr", // Special case for data attributes
			"aria-*":    "Attr", // Special case for aria attributes
			"innerHTML": "Children",
			"innerText": "Text",
		},
		SelfClose: false,
		GoPackage: "htmlgo",
	},
	"span": {
		TagName:    "span",
		GoFuncName: "Span",
		Attributes: map[string]string{
			"class":     "Class",
			"id":        "ID",
			"style":     "Style",
			"innerHTML": "Children",
			"innerText": "Text",
		},
		SelfClose: false,
		GoPackage: "htmlgo",
	},
	"p": {
		TagName:    "p",
		GoFuncName: "P",
		Attributes: map[string]string{
			"class":     "Class",
			"id":        "ID",
			"style":     "Style",
			"innerHTML": "Children",
			"innerText": "Text",
		},
		SelfClose: false,
		GoPackage: "htmlgo",
	},
	"h1": {
		TagName:    "h1",
		GoFuncName: "H1",
		Attributes: map[string]string{
			"class":     "Class",
			"id":        "ID",
			"style":     "Style",
			"innerHTML": "Children",
			"innerText": "Text",
		},
		SelfClose: false,
		GoPackage: "htmlgo",
	},
	"input": {
		TagName:    "input",
		GoFuncName: "Input",
		Attributes: map[string]string{
			"class":       "Class",
			"id":          "ID",
			"style":       "Style",
			"type":        "Type",
			"value":       "Value",
			"placeholder": "Placeholder",
			"required":    "Required",
		},
		SelfClose: true,
		GoPackage: "htmlgo",
	},
	"button": {
		TagName:    "button",
		GoFuncName: "Button",
		Attributes: map[string]string{
			"class":     "Class",
			"id":        "ID",
			"style":     "Style",
			"type":      "Type",
			"disabled":  "Disabled",
			"innerHTML": "Children",
			"innerText": "Text",
		},
		SelfClose: false,
		GoPackage: "htmlgo",
	},
	"form": {
		TagName:    "form",
		GoFuncName: "Form",
		Attributes: map[string]string{
			"class":     "Class",
			"id":        "ID",
			"style":     "Style",
			"action":    "Action",
			"method":    "Method",
			"innerHTML": "Children",
		},
		SelfClose: false,
		GoPackage: "htmlgo",
	},
}

// VuetifyTagMappings - Based on Vuetify components
var VuetifyTagMappings = map[string]TagMapping{
	"v-btn": {
		TagName:    "v-btn",
		GoFuncName: "VBtn",
		Attributes: map[string]string{
			"color":        "Color",
			"variant":      "Variant",
			"size":         "Size",
			"disabled":     "Disabled",
			"icon":         "Icon",
			"loading":      "Loading",
			"outlined":     "Outlined",
			"rounded":      "Rounded",
			"text":         "Text",
			"innerText":    "Text",
			"prepend-icon": "PrependIcon",
			"append-icon":  "AppendIcon",
		},
		SelfClose: false,
		GoPackage: "vuetify",
	},
	"v-card": {
		TagName:    "v-card",
		GoFuncName: "VCard",
		Attributes: map[string]string{
			"color":     "Color",
			"variant":   "Variant",
			"elevation": "Elevation",
			"flat":      "Flat",
			"class":     "Class",
			"innerHTML": "Children",
		},
		SelfClose: false,
		GoPackage: "vuetify",
	},
	"vx-dialog": {
		TagName:    "vx-dialog",
		GoFuncName: "VXDialog",
		Attributes: map[string]string{
			"title":       "Title",
			"text":        "Text",
			"icon":        "Icon",
			"icon-color":  "IconColor",
			"ok-text":     "OkText",
			"ok-color":    "OkColor",
			"cancel-text": "CancelText",
			"innerHTML":   "Children",
		},
		SelfClose: false,
		GoPackage: "vuetifyx",
	},
	"vx-date-picker": {
		TagName:    "vx-date-picker",
		GoFuncName: "VXDatePicker",
		Attributes: map[string]string{
			"label":     "Label",
			"clearable": "Clearable",
			"tips":      "Tips",
		},
		SelfClose: true,
		GoPackage: "vuetifyx",
	},
	"vx-tiptap-editor": {
		TagName:    "vx-tiptap-editor",
		GoFuncName: "VXTiptapEditor",
		Attributes: map[string]string{
			"label":      "Label",
			"min-height": "MinHeight",
		},
		SelfClose: true,
		GoPackage: "vuetifyx",
	},
}

// Get a combined map of all tag mappings
func GetAllTagMappings() map[string]TagMapping {
	allMappings := make(map[string]TagMapping)

	// Add standard HTML tags
	for k, v := range StandardTagMappings {
		allMappings[k] = v
	}

	// Add Vuetify components
	for k, v := range VuetifyTagMappings {
		allMappings[k] = v
	}

	return allMappings
}
