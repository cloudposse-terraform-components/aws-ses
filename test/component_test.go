package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/cloudposse/test-helpers/pkg/atmos"
	helper "github.com/cloudposse/test-helpers/pkg/atmos/aws-component-helper"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent(t *testing.T) {
	// Define the AWS region to use for the tests
	awsRegion := "us-east-2"

	// Initialize the test fixture
	fixture := helper.NewFixture(t, "../", awsRegion, "test/fixtures")

	// Ensure teardown is executed after the test
	defer fixture.TearDown()
	fixture.SetUp(&atmos.Options{})

	// Define the test suite
	fixture.Suite("default", func(t *testing.T, suite *helper.Suite) {
		// Setup phase: Create DNS zones for testing
		suite.Setup(t, func(t *testing.T, atm *helper.Atmos) {
			basicDomain := "components.cptest.test-automation.app"

			// Deploy the delegated DNS zone
			inputs := map[string]interface{}{
				"zone_config": []map[string]interface{}{
					{
						"subdomain": suite.GetRandomIdentifier(),
						"zone_name": basicDomain,
					},
				},
			}
			atm.GetAndDeploy("dns-delegated", "default-test", inputs)
		})

		// Teardown phase: Destroy the DNS zones created during setup
		suite.TearDown(t, func(t *testing.T, atm *helper.Atmos) {
			// Deploy the delegated DNS zone
			inputs := map[string]interface{}{
				"zone_config": []map[string]interface{}{
					{
						"subdomain": suite.GetRandomIdentifier(),
						"zone_name": "components.cptest.test-automation.app",
					},
				},
			}
			atm.GetAndDestroy("dns-delegated", "default-test", inputs)
		})

		// Test phase: Validate the functionality of the ALB component
		suite.Test(t, "basic", func(t *testing.T, atm *helper.Atmos) {
			dnsDelegatedComponent := helper.NewAtmosComponent("dns-delegated", "default-test", map[string]interface{}{})
			domain := atm.Output(dnsDelegatedComponent, "default_domain_name")

			hostnamePrefix := strings.ToLower(random.UniqueId())
			inputs := map[string]interface{}{
				"domain_template": hostnamePrefix + "-%[3]v.%[2]v.%[1]v." + domain,
				"ssm_prefix":      fmt.Sprintf("/ses/%s", hostnamePrefix),
			}
			defer atm.GetAndDestroy("ses/basic", "default-test", inputs)
			component := atm.GetAndDeploy("ses/basic", "default-test", inputs)
			assert.NotNil(t, component)

			smtpPassword := atm.Output(component, "smtp_password")
			assert.NotEmpty(t, smtpPassword)

			smtpUser := atm.Output(component, "smtp_user")
			assert.NotEmpty(t, smtpUser)

			userUniqueId := atm.Output(component, "user_unique_id")
			assert.NotEmpty(t, userUniqueId)

			userArn := atm.Output(component, "user_arn")
			assert.NotEmpty(t, userArn)

			randomContentMessage := strings.ToLower(random.UniqueId())

			indentityDomain := fmt.Sprintf("%s-test.ue2.default.%s", hostnamePrefix, domain)
			senderEmail := fmt.Sprintf("test@%s", indentityDomain)
			client := NewSESV2Client(t, awsRegion)

			var identityState string
			var attempt int = 0
			for strings.ToLower(identityState) != "success" && attempt <= 60 {
				indenties, err := client.GetEmailIdentity(context.Background(), &sesv2.GetEmailIdentityInput{
					EmailIdentity: &indentityDomain,
				})
				assert.NoError(t, err)

				identityState = string(indenties.VerificationStatus)
				time.Sleep(2 * time.Second)
				attempt++
			}
			assert.Equal(t, "SUCCESS", identityState)

			_, err := client.SendEmail(context.Background(), &sesv2.SendEmailInput{
				Content: &types.EmailContent{
					Raw: &types.RawMessage{
						Data: []byte(fmt.Sprintf("Test email %s", randomContentMessage)),
					},
				},
				Destination: &types.Destination{
					// https://docs.aws.amazon.com/ses/latest/dg/send-an-email-from-console.html
					ToAddresses: []string{"success@simulator.amazonses.com"},
				},
				FromEmailAddress: &senderEmail,
			})
			assert.NoError(t, err)
		})
	})
}

// NewElbV2Client creates en ELB client.
func NewSESV2Client(t *testing.T, region string) *sesv2.Client {
	client, err := NewSESV2ClientE(t, region)
	require.NoError(t, err)

	return client
}

// NewElbV2ClientE creates an ELB client.
func NewSESV2ClientE(t *testing.T, region string) (*sesv2.Client, error) {
	sess, err := aws.NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}
	return sesv2.NewFromConfig(*sess), nil
}
