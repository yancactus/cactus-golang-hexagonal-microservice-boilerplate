package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"cactus-golang-hexagonal-microservice-boilerplate/config"
	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
)

// Client wraps the DynamoDB client
type Client struct {
	DB          *dynamodb.Client
	TablePrefix string
}

// NewClient creates a new DynamoDB client based on configuration
func NewClient(ctx context.Context) (*Client, error) {
	cfg := config.GlobalConfig.DynamoDB
	if cfg == nil {
		return nil, fmt.Errorf("DynamoDB configuration is missing")
	}

	// Build AWS config options
	var opts []func(*awsconfig.LoadOptions) error

	opts = append(opts, awsconfig.WithRegion(cfg.Region))

	// Use static credentials if provided (for LocalStack)
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	// Load AWS config
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create DynamoDB client options
	var dynamoOpts []func(*dynamodb.Options)

	// Use custom endpoint if provided (for LocalStack)
	if cfg.Endpoint != "" {
		dynamoOpts = append(dynamoOpts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	}

	// Create DynamoDB client
	client := dynamodb.NewFromConfig(awsCfg, dynamoOpts...)

	c := &Client{
		DB:          client,
		TablePrefix: cfg.TablePrefix,
	}

	// Ensure required tables exist
	if err := c.ensureTablesExist(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure tables exist: %w", err)
	}

	return c, nil
}

// GetTableName returns the full table name with prefix
func (c *Client) GetTableName(tableName string) string {
	if c.TablePrefix != "" {
		return c.TablePrefix + tableName
	}
	return tableName
}

// Close closes the DynamoDB client (no-op for AWS SDK)
func (c *Client) Close() error {
	// AWS SDK v2 doesn't require explicit closing
	return nil
}

// ensureTablesExist creates required tables if they don't exist
func (c *Client) ensureTablesExist(ctx context.Context) error {
	return c.createAuditLogsTable(ctx)
}

// createAuditLogsTable creates the audit_logs table if it doesn't exist
func (c *Client) createAuditLogsTable(ctx context.Context) error {
	tableName := c.GetTableName("audit_logs")

	_, err := c.DB.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err == nil {
		return nil
	}

	var notFoundErr *types.ResourceNotFoundException
	if !errors.As(err, &notFoundErr) {
		return fmt.Errorf("failed to describe table: %w", err)
	}

	log.SugaredLogger.Infof("Creating DynamoDB table %s...", tableName)

	_, err = c.DB.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("entity_type"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("timestamp"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("entity_type_timestamp_index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("entity_type"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("timestamp"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	waiter := dynamodb.NewTableExistsWaiter(c.DB)
	err = waiter.Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}, 2*time.Minute)
	if err != nil {
		return fmt.Errorf("failed waiting for table to be active: %w", err)
	}

	log.SugaredLogger.Infof("DynamoDB table %s created successfully", tableName)
	return nil
}
