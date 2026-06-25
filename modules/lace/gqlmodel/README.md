# Gqlmodel  [![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/knesklab/util/blob/master/LICENSE)

Gqlmodel custom models for [gqlgen.com](http://gqlgen.com) grqphql module.

## Installation
```sh
go get github.com/ubgo/gqlmodel
```
Note: If you are getting error `modelgen: unable to find type github.com/ubgo/gqlmodel.UidScalar`
then create a new go file or use existing go file and add below code
```go
import _ github.com/ubgo/gqlmodel
```

## Models Support
* uuid.UUID

## How to Use
1. Edit the `gqlgen.yml` file and add the code like this under `models`
```yml
Uid:
  model: github.com/ubgo/gqlmodel.UidScalar
```

2. Under your `schema.graphqls` file add this `scalar Uid`
3. Now you can use the UUID like this:
```gql
type Category {
  id: Uid!
  name: String!
}
```

You will see the model file generated as following:
```go
type Category struct {
  ID   uuid.UUID `json:"id"`
  Name string    `json:"name"`
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
