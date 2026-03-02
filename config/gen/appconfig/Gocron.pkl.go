// Code generated from Pkl module `AppConfig`. DO NOT EDIT.
package appconfig

type Gocron struct {
	UiEnable bool `pkl:"uiEnable"`

	UiPort string `pkl:"uiPort"`

	// List of job names to register. If empty, all jobs are registered.
	EnabledJobs []string `pkl:"enabledJobs"`
}
