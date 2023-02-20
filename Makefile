build: build-darwin build-linux build-windows
build-darwin: cmd/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/darwin/amd64/aws-mfa-helper cmd/main.go
	zip -j bin/darwin/amd64/aws-mfa-helper.zip bin/darwin/amd64/aws-mfa-helper
	rm -f bin/darwin/amd64/aws-mfa-helper
build-windows: cmd/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/windows/amd64/aws-mfa-helper.exe cmd/main.go
	zip -j bin/windows/amd64/aws-mfa-helper.zip bin/windows/amd64/aws-mfa-helper.exe
	rm -f bin/windows/amd64/aws-mfa-helper.exe
build-linux: cmd/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/linux/amd64/aws-mfa-helper cmd/main.go
	zip -j bin/linux/amd64/aws-mfa-helper.zip bin/linux/amd64/aws-mfa-helper
	rm -f bin/linux/amd64/aws-mfa-helper


publish: build
	aws s3 sync bin/ s3://public-artifact-bucket-382098889955-ap-northeast-1/aws-mfa-helper/latest/ 
