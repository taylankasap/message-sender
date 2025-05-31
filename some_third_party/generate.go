package somethirdparty

//go:generate go tool oapi-codegen -config oapi-codegen.yaml openapi.yaml
//go:generate go tool mockgen --package=somethirdparty --destination=gen_mock.go . ClientWithResponsesInterface
