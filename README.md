# AWS Secrets Manager Loader

Simple go application to load secrets from AWS Secrets Manager and output them to stdout

## Build

```
CGO_ENABLED=0 GOOS=linux go build -o aws_sm_loader main.go
```

## Usage

The application expects the env variable `AWS_REGION` to be set.
To filter the secrets you want to retrieve use AWS tags. Set tags as env variables before running the application with prefix `SM_TAG_`.

To get all secrets tagged with FOO=bar use
```
export SM_TAG_FOO=bar
./aws_sm_loader
```

The secrets matching **all** tags will be printed to stdout in the following format
```
export FOO=bar
export FOO2=bar2
...
```

You can the use eval to export the env variables, for example in dumb-init entrypoint
```
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["bash", "-c", "eval $(./aws_sm_loader) && exec printenv"]
```



