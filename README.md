### Description
The project uses only standard libraries.
In tests are used the following libraries: 
- [httpexpect](github.com/gavv/httpexpect)
- [testify](github.com/stretchr/testify)
##### Run tests

```bash
make test
```

##### Build service
```bash
make build
```

##### Run service
```bash
./.build/url-reader
```
For more information about parameters execute the following command:
```bash
./.build/url-reader -help
```

