// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package ecsutil // import "github.com/amazon-contributing/opentelemetry-collector-contrib/processor/awsapplicationsignalsprocessor/internal/ecsutil"

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/amazon-contributing/opentelemetry-collector-contrib/processor/awsapplicationsignalsprocessor/internal/httpclient"
)

const (
	v2MetadataEndpoint    = "http://169.254.170.2/v2/metadata"
	v3MetadataEndpointEnv = "ECS_CONTAINER_METADATA_URI"
	v4MetadataEndpointEnv = "ECS_CONTAINER_METADATA_URI_V4"
)

// The following values are borrowed from:
// - https://github.com/aws/amazon-cloudwatch-agent/blob/bde3bd9775ae1d4e4f8a2fdb92d7b6fdd5186fba/cfg/envconfig/envconfig.go
const (
	RunInContainer = "RUN_IN_CONTAINER"
	TrueValue      = "True"
)

type ecsMetadataResponse struct {
	Cluster string
	TaskARN string
}

type EcsUtil struct {
	Cluster    string
	Region     string
	TaskARN    string
	httpClient *httpclient.HTTPClient
}

var ecsUtilInstance *EcsUtil

var ecsUtilOnce sync.Once

func GetECSUtilSingleton() *EcsUtil {
	ecsUtilOnce.Do(func() {
		ecsUtilInstance = initECSUtilSingleton()
	})
	return ecsUtilInstance
}

func initECSUtilSingleton() (newInstance *EcsUtil) {
	newInstance = &EcsUtil{httpClient: httpclient.New()}
	if os.Getenv(RunInContainer) != TrueValue {
		return
	}
	log.Println("I! attempt to access ECS task metadata to determine whether I'm running in ECS.")
	ecsMetadataResponse, err := newInstance.getECSMetadata()

	if err != nil {
		log.Printf("I! access ECS task metadata fail with response %v, assuming I'm not running in ECS.\n", err)
		return
	}

	newInstance.parseRegion(ecsMetadataResponse)
	newInstance.parseClusterName(ecsMetadataResponse)
	newInstance.TaskARN = ecsMetadataResponse.TaskARN
	return

}

func (e *EcsUtil) IsECS() bool {
	return e.Region != ""
}

func (e *EcsUtil) getECSMetadata() (em *ecsMetadataResponse, err error) {
	// Based on endpoint to get ECS metadata, for more information on the respond, https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint.html
	if v4MetadataEndpoint, ok := os.LookupEnv(v4MetadataEndpointEnv); ok {
		em, err = e.getMetadataResponse(v4MetadataEndpoint + "/task")
	} else if v3MetadataEndpoint, ok := os.LookupEnv(v3MetadataEndpointEnv); ok {
		em, err = e.getMetadataResponse(v3MetadataEndpoint + "/task")
	} else {
		em, err = e.getMetadataResponse(v2MetadataEndpoint)
	}
	return
}

func (e *EcsUtil) getMetadataResponse(endpoint string) (em *ecsMetadataResponse, err error) {
	em = &ecsMetadataResponse{}
	resp, err := e.httpClient.Request(endpoint)

	if err != nil {
		return
	}

	err = json.Unmarshal(resp, em)
	if err != nil {
		log.Printf("E! Unable to parse response from ecsmetadata endpoint, error: %v", err)
		log.Printf("D! Content is %s", string(resp))
	}
	return
}

// There are two formats of Task ARN (https://docs.aws.amazon.com/AmazonECS/latest/userguide/ecs-account-settings.html#ecs-resource-ids)
// arn:aws:ecs:region:aws_account_id:task/task-id
// arn:aws:ecs:region:aws_account_id:task/cluster-name/task-id
// This function will return region extracted from Task ARN
func (e *EcsUtil) parseRegion(em *ecsMetadataResponse) {
	splitedContent := strings.Split(em.TaskARN, ":")
	// When splitting the ARN with ":", the 4th segment is the region
	if len(splitedContent) < 4 {
		log.Printf("E! Invalid ecs task arn: %s", em.TaskARN)
	}
	e.Region = splitedContent[3]
}

// There is only one format for ClusterArn (https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_Cluster.html)
// arn:aws:ecs:region:aws_account_id:cluster/cluster-name
func (e *EcsUtil) parseClusterName(em *ecsMetadataResponse) {
	splitedContent := strings.Split(em.Cluster, "/")
	// When splitting the ClusterName with /, the last is always the cluster name
	if len(splitedContent) == 0 {
		log.Printf("E! Invalid cluster arn: %s", em.Cluster)
	}
	e.Cluster = splitedContent[len(splitedContent)-1]
}
