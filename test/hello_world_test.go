package test

import (
    "testing"

    "github.com/gruntwork-io/terratest/modules/terraform"
    "github.com/stretchr/testify/assert"
)

func TestHelloWorld(t *testing.T) {
    t.Parallel()

    // Define Terraform options
    terraformOptions := &terraform.Options{
        TerraformDir: "../", // Path to the Terraform module directory
    }

    // Clean up resources at the end
    defer terraform.Destroy(t, terraformOptions)

    // Initialize and apply Terraform
    terraform.InitAndApply(t, terraformOptions)

    // Retrieve the "hello_world" output
    helloWorldOutput := terraform.Output(t, terraformOptions, "hello_world")

    // Validate the output
    assert.Equal(t, "Hello, Terratest!", helloWorldOutput, "Output does not match expected value")
} 