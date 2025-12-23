package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/repo"
)

const auditLogsTable = "audit_logs"

// AuditRepository implements the IAuditRepo interface using DynamoDB
type AuditRepository struct {
	client *Client
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(client *Client) repo.IAuditRepo {
	return &AuditRepository{client: client}
}

// auditLogItem represents the DynamoDB item structure
type auditLogItem struct {
	ID         string                 `dynamodbav:"id"`
	EntityType string                 `dynamodbav:"entity_type"`
	EntityID   string                 `dynamodbav:"entity_id"`
	Action     string                 `dynamodbav:"action"`
	Payload    map[string]interface{} `dynamodbav:"payload"`
	Timestamp  string                 `dynamodbav:"timestamp"`
	UserID     *int                   `dynamodbav:"user_id,omitempty"`
}

// Create saves a new audit log entry
func (r *AuditRepository) Create(ctx context.Context, audit *model.AuditLog) error {
	item := auditLogItem{
		ID:         audit.ID,
		EntityType: audit.EntityType,
		EntityID:   audit.EntityID,
		Action:     audit.Action,
		Payload:    audit.Payload,
		Timestamp:  audit.Timestamp.Format(time.RFC3339),
		UserID:     audit.UserID,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}

	tableName := r.client.GetTableName(auditLogsTable)
	_, err = r.client.DB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put audit log item: %w", err)
	}

	return nil
}

// GetByID retrieves an audit log by its ID
func (r *AuditRepository) GetByID(ctx context.Context, id string) (*model.AuditLog, error) {
	tableName := r.client.GetTableName(auditLogsTable)

	result, err := r.client.DB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var item auditLogItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal audit log: %w", err)
	}

	timestamp, _ := time.Parse(time.RFC3339, item.Timestamp)

	return &model.AuditLog{
		ID:         item.ID,
		EntityType: item.EntityType,
		EntityID:   item.EntityID,
		Action:     item.Action,
		Payload:    item.Payload,
		Timestamp:  timestamp,
		UserID:     item.UserID,
	}, nil
}

// FindByEntityType retrieves audit logs by entity type with pagination
func (r *AuditRepository) FindByEntityType(ctx context.Context, entityType string, limit int, lastKey string) ([]*model.AuditLog, string, error) {
	tableName := r.client.GetTableName(auditLogsTable)

	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		IndexName:              aws.String("entity_type_timestamp_index"),
		KeyConditionExpression: aws.String("entity_type = :et"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":et": &types.AttributeValueMemberS{Value: entityType},
		},
		ScanIndexForward: aws.Bool(false), // Descending order by timestamp
		Limit:            aws.Int32(int32(limit)),
	}

	// Handle pagination
	if lastKey != "" {
		input.ExclusiveStartKey = map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: lastKey},
		}
	}

	result, err := r.client.DB.Query(ctx, input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query audit logs: %w", err)
	}

	var audits []*model.AuditLog
	for _, item := range result.Items {
		var logItem auditLogItem
		if err := attributevalue.UnmarshalMap(item, &logItem); err != nil {
			continue
		}

		timestamp, _ := time.Parse(time.RFC3339, logItem.Timestamp)
		audits = append(audits, &model.AuditLog{
			ID:         logItem.ID,
			EntityType: logItem.EntityType,
			EntityID:   logItem.EntityID,
			Action:     logItem.Action,
			Payload:    logItem.Payload,
			Timestamp:  timestamp,
			UserID:     logItem.UserID,
		})
	}

	var nextKey string
	if result.LastEvaluatedKey != nil {
		if idAttr, ok := result.LastEvaluatedKey["id"].(*types.AttributeValueMemberS); ok {
			nextKey = idAttr.Value
		}
	}

	return audits, nextKey, nil
}

// FindByEntityID retrieves audit logs for a specific entity
func (r *AuditRepository) FindByEntityID(ctx context.Context, entityType, entityID string) ([]*model.AuditLog, error) {
	tableName := r.client.GetTableName(auditLogsTable)

	input := &dynamodb.ScanInput{
		TableName:        aws.String(tableName),
		FilterExpression: aws.String("entity_type = :et AND entity_id = :eid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":et":  &types.AttributeValueMemberS{Value: entityType},
			":eid": &types.AttributeValueMemberS{Value: entityID},
		},
	}

	result, err := r.client.DB.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan audit logs: %w", err)
	}

	var audits []*model.AuditLog
	for _, item := range result.Items {
		var logItem auditLogItem
		if err := attributevalue.UnmarshalMap(item, &logItem); err != nil {
			continue
		}

		timestamp, _ := time.Parse(time.RFC3339, logItem.Timestamp)
		audits = append(audits, &model.AuditLog{
			ID:         logItem.ID,
			EntityType: logItem.EntityType,
			EntityID:   logItem.EntityID,
			Action:     logItem.Action,
			Payload:    logItem.Payload,
			Timestamp:  timestamp,
			UserID:     logItem.UserID,
		})
	}

	return audits, nil
}
