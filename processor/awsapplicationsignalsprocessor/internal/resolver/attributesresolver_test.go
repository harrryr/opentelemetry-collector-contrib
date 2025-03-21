// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package resolver

import (
	"context"
	"errors"
	"testing"

	"github.com/amazon-contributing/opentelemetry-collector-contrib/processor/awsapplicationsignalsprocessor/common"
	"github.com/amazon-contributing/opentelemetry-collector-contrib/processor/awsapplicationsignalsprocessor/config"
	attr "github.com/amazon-contributing/opentelemetry-collector-contrib/processor/awsapplicationsignalsprocessor/internal/attributes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/collector/semconv/v1.22.0"
	"go.uber.org/zap"
)

type MockSubResolver struct {
	mock.Mock
}

func (m *MockSubResolver) Process(attributes, resourceAttributes pcommon.Map) error {
	args := m.Called(attributes, resourceAttributes)
	return args.Error(0)
}

func (m *MockSubResolver) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestResourceAttributesResolverWithNoConfiguredName(t *testing.T) {
	tests := []struct {
		name         string
		platformCode string
		platformType string
		resolver     config.Resolver
	}{
		{
			"testOnGeneric",
			config.PlatformGeneric,
			AttributePlatformGeneric,
			config.NewGenericResolver(""),
		},
		{
			"testOnEC2",
			config.PlatformEC2,
			AttributePlatformEC2,
			config.NewEC2Resolver(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := zap.NewDevelopment()
			attributesResolver := NewAttributesResolver([]config.Resolver{tt.resolver}, logger)
			resolver := attributesResolver.subResolvers[0]

			attributes := pcommon.NewMap()
			resourceAttributes := pcommon.NewMap()

			_ = resolver.Process(attributes, resourceAttributes)

			attribute, ok := attributes.Get(common.AttributePlatformType)
			assert.True(t, ok)
			assert.Equal(t, tt.platformType, attribute.Str())

			attribute, ok = attributes.Get(attr.AWSLocalEnvironment)
			assert.True(t, ok)
			assert.Equal(t, tt.platformCode+":default", attribute.Str())
		})
	}
}

func TestResourceAttributesResolverWithOnEC2WithASG(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	attributesResolver := NewAttributesResolver([]config.Resolver{config.NewEC2Resolver("")}, logger)
	resolver := attributesResolver.subResolvers[0]

	attributes := pcommon.NewMap()
	resourceAttributes := pcommon.NewMap()
	resourceAttributes.PutStr(attr.ResourceDetectionASG, "my-asg")

	_ = resolver.Process(attributes, resourceAttributes)
	platformAttr, ok := attributes.Get(common.AttributePlatformType)
	assert.True(t, ok)
	assert.Equal(t, "AWS::EC2", platformAttr.Str())
	envAttr, ok := attributes.Get(attr.AWSLocalEnvironment)
	assert.True(t, ok)
	assert.Equal(t, "ec2:my-asg", envAttr.Str())
}

func TestResourceAttributesResolverWithHostname(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	attributesResolver := NewAttributesResolver([]config.Resolver{config.NewGenericResolver("")}, logger)
	resolver := attributesResolver.subResolvers[0]

	attributes := pcommon.NewMap()
	resourceAttributes := pcommon.NewMap()
	resourceAttributes.PutStr(attr.ResourceDetectionHostName, "hostname")

	_ = resolver.Process(attributes, resourceAttributes)
	envAttr, ok := attributes.Get(common.AttributeHost)
	assert.True(t, ok)
	assert.Equal(t, "hostname", envAttr.AsString())
}

func TestResourceAttributesResolverWithCustomEnvironment(t *testing.T) {
	tests := []struct {
		name         string
		platformCode string
		resolver     config.Resolver
	}{
		{
			"testOnGeneric",
			config.PlatformGeneric,
			config.NewGenericResolver(""),
		},
		{
			"testOnEC2",
			config.PlatformEC2,
			config.NewEC2Resolver(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := zap.NewDevelopment()
			attributesResolver := NewAttributesResolver([]config.Resolver{tt.resolver}, logger)
			resolver := attributesResolver.subResolvers[0]

			attributes := pcommon.NewMap()
			resourceAttributes := pcommon.NewMap()
			// insert default env
			resourceAttributes.PutStr(attr.ResourceDetectionASG, "my-asg")
			resourceAttributes.PutStr(semconv.AttributeAWSECSTaskARN, "arn:aws:ecs:us-west-1:123456789123:task/my-cluster/10838bed-421f-43ef-870a-f43feacbbb5b")

			// insert custom env
			resourceAttributes.PutStr(attr.AWSHostedInEnvironment, "env1")
			_ = resolver.Process(attributes, resourceAttributes)
			envAttr, ok := attributes.Get(attr.AWSLocalEnvironment)
			assert.True(t, ok)
			assert.Equal(t, "env1", envAttr.Str())

			attributes = pcommon.NewMap()
			resourceAttributes = pcommon.NewMap()

			resourceAttributes.PutStr(attr.AWSHostedInEnvironment, "error")
			resourceAttributes.PutStr(semconv.AttributeDeploymentEnvironment, "env2")
			_ = resolver.Process(attributes, resourceAttributes)
			envAttr, ok = attributes.Get(attr.AWSLocalEnvironment)
			assert.True(t, ok)
			assert.Equal(t, "env2", envAttr.Str())

			attributes = pcommon.NewMap()
			resourceAttributes = pcommon.NewMap()

			resourceAttributes.PutStr(semconv.AttributeDeploymentEnvironment, "env3")
			_ = resolver.Process(attributes, resourceAttributes)
			envAttr, ok = attributes.Get(attr.AWSLocalEnvironment)
			assert.True(t, ok)
			assert.Equal(t, "env3", envAttr.Str())
		})
	}
}

func TestAttributesResolver_Process(t *testing.T) {
	attributes := pcommon.NewMap()
	resourceAttributes := pcommon.NewMap()

	mockSubResolver1 := new(MockSubResolver)
	mockSubResolver1.On("Process", attributes, resourceAttributes).Return(nil)

	mockSubResolver2 := new(MockSubResolver)
	mockSubResolver2.On("Process", attributes, resourceAttributes).Return(errors.New("error"))

	r := &AttributesResolver{
		subResolvers: []subResolver{mockSubResolver1, mockSubResolver2},
	}

	err := r.Process(attributes, resourceAttributes, true)
	assert.Error(t, err)
	mockSubResolver1.AssertExpectations(t)
	mockSubResolver2.AssertExpectations(t)
}

func TestAttributesResolver_Stop(t *testing.T) {
	ctx := context.Background()

	mockSubResolver1 := new(MockSubResolver)
	mockSubResolver1.On("Stop", ctx).Return(nil)

	mockSubResolver2 := new(MockSubResolver)
	mockSubResolver2.On("Stop", ctx).Return(errors.New("error"))

	r := &AttributesResolver{
		subResolvers: []subResolver{mockSubResolver1, mockSubResolver2},
	}

	err := r.Stop(ctx)
	assert.Error(t, err)
	mockSubResolver1.AssertExpectations(t)
	mockSubResolver2.AssertExpectations(t)
}
