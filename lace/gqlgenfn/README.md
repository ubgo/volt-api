# gqlgenfn  [![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/knesklab/util/blob/master/LICENSE)

A small sets of useful function for [gqlgen.com](http://gqlgen.com) grqphql module with [Gin Gonic](https://github.com/gin-gonic/gin).

## Installation
```
go get github.com/ubgo/gqlgenfn
```

### GinContextToContextMiddleware() gin.HandlerFunc
Make the GIN Context available within Gqlgen Context
```go
router.Use(gql_fn.GinContextToContextMiddleware())
```

### GinContextFromContext(ctx context.Context) (*gin.Context, error)
return the GIN Context from the Gqlgen Context
```go
gc, err := gql_fn.GinContextFromContext(ctx)
if err != nil {
  return false
}
```


## Contribute

If you would like to contribute to the project, please fork it and send us a pull request.  Please add tests
for any new features or bug fixes.

## Stay in touch

* Author - [Aman Khanakia](https://twitter.com/mrkhanakia)
* Website - [https://khanakia.com](https://khanakia.com/)

## License

goutil is [MIT licensed](LICENSE).
