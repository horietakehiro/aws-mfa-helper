package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
)

const (
	MIN_DURATION     = 60 * 15
	DEFAULT_DURATION = 60 * 60 * 12
	MAX_DURATION     = 60 * 60 * 36
)

var (
	profile  string
	code     string
	duration int64
)

func init() {

	flag.StringVar(&profile, "p", "default", "AWS cli profile name to use")
	flag.StringVar(&code, "c", "", "MFA one time code")
	flag.Int64Var(&duration, "d", DEFAULT_DURATION, fmt.Sprintf(
		"Duration in second. Min value is %v. Max value is %v", MIN_DURATION, MAX_DURATION),
	)
	flag.Parse()

	// validation
	if code == "" {
		fmt.Println("arg '-c' is required")
		os.Exit(1)
	}
	if duration < MIN_DURATION || duration > MAX_DURATION {
		fmt.Printf("value of arg '-d' is allowed between %v and %v\n", MIN_DURATION, MAX_DURATION)
		os.Exit(1)
	}

}

func main() {
	fmt.Printf("Profile to use is %s\n", profile)
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		SharedConfigState: session.SharedConfigEnable,
	}))

	stsClient := sts.New(sess)
	// get caller identity
	callerIdentity, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	arnElems := strings.Split(*callerIdentity.Arn, "/")
	userName := arnElems[len(arnElems)-1]
	fmt.Printf("Target IAM user is %s\n", userName)

	// get mfa device serial number
	iamClient := iam.New(sess)
	mfaDevices, err := iamClient.ListMFADevices(&iam.ListMFADevicesInput{
		UserName: &userName,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	serialNumber := *mfaDevices.MFADevices[0].SerialNumber
	fmt.Printf("Target MFA device serial number is %s\n", serialNumber)

	// get session token
	sessionTolenOutput, err := stsClient.GetSessionToken(&sts.GetSessionTokenInput{
		DurationSeconds: &duration,
		SerialNumber:    &serialNumber,
		TokenCode:       &code,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	accessKey := sessionTolenOutput.Credentials.AccessKeyId
	secretAccessKey := sessionTolenOutput.Credentials.SecretAccessKey
	sessionToken := sessionTolenOutput.Credentials.SessionToken
	defaultRegion := *sess.Config.Region

	// output results
	fmt.Println("set session token as environment variables like...")
	fmt.Println("==============================================")
	fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", *accessKey)
	fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", *secretAccessKey)
	fmt.Printf("export AWS_SESSION_TOKEN=%s\n", *sessionToken)
	fmt.Printf("export AWS_DEFAULT_REGION=%s\n", defaultRegion)
	fmt.Println("==============================================")

	fmt.Println("")

	newProfileName := fmt.Sprintf("%s-mfa", profile)
	fmt.Println("set session token as CLI config file like...")
	fmt.Println("==============================================")
	fmt.Printf("aws configure set aws_access_key_id %s --profile %s\n", *accessKey, newProfileName)
	fmt.Printf("aws configure set aws_secret_access_key %s --profile %s\n", *secretAccessKey, newProfileName)
	fmt.Printf("aws configure set aws_session_token %s --profile %s\n", *sessionToken, newProfileName)
	fmt.Printf("aws configure set region %s --profile %s\n", defaultRegion, newProfileName)
	fmt.Println("==============================================")
}
