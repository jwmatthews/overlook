package cmd

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/jwmatthews/overlook/pkg/overlook"
	"github.com/spf13/cobra"
)

// EmailCommand cobra command to email a Report
var EmailCommand = &cobra.Command{
	Use:   "email",
	Short: "Email a usage report",
	Long:  `Email a usage report`,
	Run: func(cmd *cobra.Command, args []string) {
		EmailReport()
	},
}

const (
	// This address must be verified with Amazon SES.
	Sender = "jmatthews+migration@redhat.com"

	// If the SES account is still in the sandbox, this address must be verified.
	Recipient = "jmatthews+migration@redhat.com"

	// The character encoding for the email.
	CharSet = "UTF-8"
)

func EmailReport() {

	usageFileNames := overlook.GetBillingDataSortedFileNames()
	fmt.Println(usageFileNames)

	reports := make([]overlook.ReportDaily, 0)
	for _, f := range usageFileNames {
		fmt.Println("Processing: ", f)
		dailyEntry := overlook.ReadSnapshotInfo(f)
		r := overlook.GetReport(dailyEntry)
		reports = append(reports, r)
	}
	SendEmail(reports)
}

func SendEmail(reports []overlook.ReportDaily) {
	var body string
	for _, r := range reports {
		body = body + r.FormatByCost() + "\n"
	}
	now := time.Now()
	ymd := now.Format("01-02-2006")

	subject := "Migration Eng AWS Usage for " + ymd

	htmlBody := "<h1>AWS EC2 Usage Report for Migration Engineering</h1>" +
		"<p>This report was produced by <a href='https://github.com/jwmatthews/overlook'>https://github.com/jwmatthews/overlook</a></p>" +
		"<h3>Report Output Below</h3>" +
		"<p><pre>" + body + "</pre></p>"

	textBody := "This report was produced by 'https://github.com/jwmatthews/overlook'\n" +
		"Report Output Below" + body

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	svc := ses.New(sess)

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(htmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(textBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(Sender),
	}

	// Attempt to send the email.
	result, err := svc.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

		return
	}

	fmt.Println("Email Sent to address: " + Recipient)
	fmt.Println(result)
}
