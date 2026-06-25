// Code generated from Pkl module `Logger`. DO NOT EDIT.
package loggerconfig

type Setting struct {
	EnableStdout bool `pkl:"enableStdout"`

	EnableOtel bool `pkl:"enableOtel"`

	EnableFile bool `pkl:"enableFile"`

	LocalConfig LocalConfig `pkl:"localConfig"`
}
