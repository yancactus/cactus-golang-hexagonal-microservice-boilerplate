#!/bin/bash

# Wait for LocalStack to be ready
sleep 5

# Create audit_logs table
awslocal dynamodb create-table \
    --table-name audit_logs \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=entity_type,AttributeType=S \
        AttributeName=timestamp,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        '[
            {
                "IndexName": "entity_type_timestamp_index",
                "KeySchema": [
                    {"AttributeName": "entity_type", "KeyType": "HASH"},
                    {"AttributeName": "timestamp", "KeyType": "RANGE"}
                ],
                "Projection": {"ProjectionType": "ALL"},
                "ProvisionedThroughput": {"ReadCapacityUnits": 5, "WriteCapacityUnits": 5}
            }
        ]' \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --region us-east-1

echo "DynamoDB audit_logs table created successfully"
