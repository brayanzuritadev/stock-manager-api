package dto

type contextKey string

const (
	KeyPath           contextKey = "path"
	KeyMethod         contextKey = "method"
	KeyBody           contextKey = "body"
	KeyDatabase       contextKey = "database"
	KeyJWTSign        contextKey = "jwtSign"
	KeyEmail          contextKey = "email"
	KeyEmailPwd       contextKey = "emailPwd"
	KeyCloudName      contextKey = "cloudName"
	KeyCloudAPIKey    contextKey = "cloudAPIKey"
	KeyCloudAPISecret contextKey = "cloudAPISecret"
)
