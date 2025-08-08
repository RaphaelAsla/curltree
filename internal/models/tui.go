package models

type AppState int

const (
	StateLoading AppState = iota
	StateProfileView
	StateProfileEdit
	StateProfileCreate
	StateConfirmDelete
)

type TUIModel struct {
	State      AppState
	User       *User
	Width      int
	Height     int
	Message    string
	Error      string
	
	FormModel  FormModel
	Confirmed  bool
}

type FormModel struct {
	Fields      []FormField
	FocusIndex  int
	Editing     bool
}

type FormField struct {
	Label       string
	Value       string
	Placeholder string
	Required    bool
	Type        FieldType
	MaxLength   int
}

type FieldType int

const (
	FieldText FieldType = iota
	FieldTextArea
	FieldURL
)

type KeyBinding struct {
	Key  string
	Desc string
}

var (
	ProfileViewKeys = []KeyBinding{
		{"ctrl+e", "edit profile"},
		{"ctrl+c", "exit"},
		{"ctrl+d", "delete profile"},
	}
	
	ProfileEditKeys = []KeyBinding{
		{"tab", "next field"},
		{"shift+tab", "prev field"},
		{"ctrl+n", "add link"},
		{"ctrl+d", "delete link"},
		{"ctrl+s", "save"},
		{"esc", "cancel"},
	}
	
	ProfileCreateKeys = []KeyBinding{
		{"tab", "next field"},
		{"shift+tab", "prev field"},
		{"ctrl+n", "add link"},
		{"ctrl+d", "delete link"},
		{"ctrl+s", "create"},
		{"esc", "cancel"},
	}
)