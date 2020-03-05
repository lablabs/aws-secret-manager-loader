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

## Writing binary secrets to file

If secret contains binary data it will be written to file. Value after last `/` from secret name will be used as filename.
Path for files can be set using `SM_SECRETS_PATH` env variable. Default is current directory.

## Ignoring secrets

If tag `aws_sm_loader_ignore` with value `true` is set for a secret, it won't be exported into the env.

## File permissions for binary secrets

File permissions for secrets that will be outputted into files can be set using `SM_SECRETS_FILEMODE` env variable.

Values in *octal permissions notation* with leading zero is expected.

Default value is read only: `0440`.

