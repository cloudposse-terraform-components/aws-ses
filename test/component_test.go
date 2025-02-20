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
	helper "github.com/cloudposse/test-helpers/pkg/atmos/component-helper"
	awshelper "github.com/cloudposse/test-helpers/pkg/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
)

type ComponentSuite struct {
	helper.TestSuite
}

func (s *ComponentSuite) TestBasic() {
	const component = "ses/basic"
	const stack = "default-test"
	const awsRegion = "us-east-2"

	dnsDelegatedOptions := s.GetAtmosOptions("dns-delegated", stack, nil)
	domain := atmos.Output(s.T(), dnsDelegatedOptions, "default_domain_name")

	hostnamePrefix := strings.ToLower(random.UniqueId())
	inputs := map[string]interface{}{
		"domain_template": hostnamePrefix + "-%[3]v.%[2]v.%[1]v." + domain,
		"ssm_prefix":      fmt.Sprintf("/ses/%s", hostnamePrefix),
	}
	defer s.DestroyAtmosComponent(s.T(), component, stack, &inputs)
	options, _ := s.DeployAtmosComponent(s.T(), component, stack, &inputs)
	assert.NotNil(s.T(), options)

	smtpPassword := atmos.Output(s.T(), options, "smtp_password")
	assert.NotEmpty(s.T(), smtpPassword)

	smtpUser := atmos.Output(s.T(), options, "smtp_user")
	assert.NotEmpty(s.T(), smtpUser)

	userUniqueId := atmos.Output(s.T(), options, "user_unique_id")
	assert.NotEmpty(s.T(), userUniqueId)

	userArn := atmos.Output(s.T(), options, "user_arn")
	assert.NotEmpty(s.T(), userArn)

	randomContentMessage := strings.ToLower(random.UniqueId())
	identityDomain := fmt.Sprintf("%s-test.ue2.default.%s", hostnamePrefix, domain)
	senderEmail := fmt.Sprintf("test@%s", identityDomain)
	client := awshelper.NewSESV2Client(s.T(), awsRegion)

	var identityState string
	var attempt int
	for strings.ToLower(identityState) != "success" && attempt <= 60 {
		identities, err := client.GetEmailIdentity(context.Background(), &sesv2.GetEmailIdentityInput{
			EmailIdentity: &identityDomain,
		})
		assert.NoError(s.T(), err)

		identityState = string(identities.VerificationStatus)
		time.Sleep(2 * time.Second)
		attempt++
	}
	assert.Equal(s.T(), "SUCCESS", identityState)

	_, err := client.SendEmail(context.Background(), &sesv2.SendEmailInput{
		Content: &types.EmailContent{
			Raw: &types.RawMessage{
				Data: []byte(fmt.Sprintf("Test email %s", randomContentMessage)),
			},
		},
		Destination: &types.Destination{
			ToAddresses: []string{"success@simulator.amazonses.com"},
		},
		FromEmailAddress: &senderEmail,
	})
	assert.NoError(s.T(), err)

	s.DriftTest(component, stack, &inputs)
}

func (s *ComponentSuite) TestEnabledFlag() {
	const component = "ses/disabled"
	const stack = "default-test"
	const awsRegion = "us-east-2"

	dnsDelegatedOptions := s.GetAtmosOptions("dns-delegated", stack, nil)
	domain := atmos.Output(s.T(), dnsDelegatedOptions, "default_domain_name")

	hostnamePrefix := strings.ToLower(random.UniqueId())
	inputs := map[string]interface{}{
		"domain_template": hostnamePrefix + "-%[3]v.%[2]v.%[1]v." + domain,
		"ssm_prefix":      fmt.Sprintf("/ses/%s", hostnamePrefix),
	}

	s.VerifyEnabledFlag(component, stack, &inputs)
}


func TestRunSuite(t *testing.T) {
	suite := new(ComponentSuite)

	subdomain := strings.ToLower(random.UniqueId())
	inputs := map[string]interface{}{
		"zone_config": []map[string]interface{}{
			{
				"subdomain": subdomain,
				"zone_name": "components.cptest.test-automation.app",
			},
		},
	}
	suite.AddDependency(t, "dns-delegated", "default-test", &inputs)
	helper.Run(t, suite)
}
