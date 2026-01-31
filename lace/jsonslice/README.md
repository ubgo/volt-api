## How to use

- Add the line in entgo schema
```go
field.JSON("body", jsonslice.JsonSlice{}).Optional(),
```
- Add the line in schema.graphql file
```gql
scalar JsonSlice
```

- In gqlgen.yml file add the following code
```yml
autobind:
  - github.com/99designs/gqlgen/graphql/introspection
  - lace/jsonslice
```

- That's it now can send the `body` as `[]interface{}` slice as below

```gql

mutation updatePostTypeForm($id: ID!, $input: UpdatePostTypeFormInput!) {
  updatePostTypeForm(id: $id, input: $input) {
    id
    name
  }
}
```

```json
{
  "id": "ute82dg64orkmh7fic",
  "input": {
    "name": "Test",
    "status": true,
    "postTypeID": "01J37T7P4HA9KB1H0T61T0ZPA0",
    "body": [
      {
        "label": "Name",
        "name": "name",
        "type": "text"
      }
    ]
  }
}
```