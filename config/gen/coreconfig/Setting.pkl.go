// Code generated from Pkl module `Core`. DO NOT EDIT.
package coreconfig

import "config/gen/coreconfig/environment"

type Setting struct {
	// Application name identifier
	AppName string `pkl:"appName"`

	// Unique identifier for each microservice (msv)
	AppKey string `pkl:"appKey"`

	// Deployment environment (dev, stag, or prod)
	Environment environment.Environment `pkl:"environment"`

	// Root directory path for the application
	Rootdir string `pkl:"rootdir"`

	// Temporary directory path for temporary files
	Tmpdir string `pkl:"tmpdir"`

	// Internal key for service-to-service communication
	InternalKey string `pkl:"internalKey"`

	// Directory path for uploaded files
	Uploadsdir string `pkl:"uploadsdir"`

	// Secret key for JWT token signing and verification
	Jwtsecret string `pkl:"jwtsecret"`

	// Encryption key for data encryption/decryption operations
	EncryptionKey string `pkl:"encryptionKey"`
}
