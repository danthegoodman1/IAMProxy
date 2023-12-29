import {CreateTableCommand, DynamoDBClient, GetItemCommand, PutItemCommand, DynamoDB } from '@aws-sdk/client-dynamodb'
import { DynamoDBDocument, DynamoDBDocumentClient } from "@aws-sdk/lib-dynamodb";

const TABLE_NAME = 'test_table'
const TABLE_NAME_2 = 'test_table_two'

async function main() {
  console.log("Running client test...")
  const client = new DynamoDBClient({
    endpoint: "http://localhost:8080",
    credentials: {
      accessKeyId: 'testuser',
      secretAccessKey: 'testpassword'
    }
  })
  const documentClient = DynamoDBDocument.from(new DynamoDB({
    endpoint: "http://localhost:8080",
    credentials: {
      accessKeyId: 'testuser',
      secretAccessKey: 'testpassword'
    }
  }))

  try {
    console.log('Trying to create table with strings...')
    const cmd = new CreateTableCommand({
      TableName: TABLE_NAME,
      KeySchema: [
        {
          AttributeName: "pk",
          KeyType: "HASH"
        },
        {
          AttributeName: "sk",
          KeyType: "RANGE"
        },
      ],
      AttributeDefinitions: [
        {
          AttributeName: "pk",
          AttributeType: "S"
        },
        {
          AttributeName: "sk",
          AttributeType: "S"
        },
      ]
    })
    const res = await client.send(cmd)
    console.log('Create table response:')
    console.log(JSON.stringify(res, null, 2))
  } catch (error) {
    console.error("Error trying to create table:")
    throw error
  }

  try {
    console.log('Trying to create table with indexes...')
    const cmd = new CreateTableCommand({
      TableName: TABLE_NAME_2,
      KeySchema: [
        {
          AttributeName: "pk",
          KeyType: "HASH"
        },
        {
          AttributeName: "sk",
          KeyType: "RANGE"
        },
      ],
      AttributeDefinitions: [
        {
          AttributeName: "pk",
          AttributeType: "S"
        },
        {
          AttributeName: "sk",
          AttributeType: "S"
        },
        {
          AttributeName: "GSI1PK",
          AttributeType: "S"
        },
        {
          AttributeName: "GSI2PK",
          AttributeType: "N"
        },
      ],
      GlobalSecondaryIndexes: [
        {
          IndexName: "test_index_1",
          KeySchema: [
            {
              AttributeName: "GSI1PK",
              KeyType: "HASH"
            },
            {
              AttributeName: "sk",
              KeyType: "RANGE"
            }
          ],
          Projection: {} // not used by api currently
        },
      ],
      LocalSecondaryIndexes: [
        {
          IndexName: "test_index_2",
          KeySchema: [
            {
              AttributeName: "GSI2PK",
              KeyType: "HASH"
            },
          ],
          Projection: {} // not used by api currently
        },
      ]
    })
    const res = await client.send(cmd)
    console.log('Create table response:')
    console.log(JSON.stringify(res, null, 2))
  } catch (error) {
    console.error("Error trying to create table:")
    throw error
  }

  try {
    console.log("Attempting to put record")
    const res = documentClient.put({
      TableName: TABLE_NAME,
      Item: {
        pk: 'prim key',
        sk: 'a sort key',
        bool: true,
        numberArray: [1, 2],
        stringArray: ['hey', 'ho'],
        nestedJSON: {
          hey: 'ho',
          lets: [1, 2, 3]
        }
      }
    })
    console.log('Put record response:')
    console.log(JSON.stringify(res, null, 2))
  } catch (error) {
    console.error("Error trying to put record:")
    throw error
  }

  try {
    console.log("Attempting to get record")
    const res = await documentClient.get({
      TableName: TABLE_NAME,
      Key: {
        pk: 'prim key',
        sk: 'a sort key'
      }
    })
    console.log('Get record response:')
    console.log(JSON.stringify(res, null, 2))
  } catch (error) {
    console.error("Error trying to put record:")
    throw error
  }

  try {
    console.log("Attempting to query record")
    const res = await documentClient.query({
      TableName: TABLE_NAME,
      KeyConditionExpression: "#p = :i AND #s >= :o",
      ExpressionAttributeNames: {
        "#p": "pk",
        "#s": "sk"
      },
      ExpressionAttributeValues: {
        ":i": "prim key",
        ":o": "a sort"
      }
    })
    console.log('Get record response:')
    console.log(JSON.stringify(res, null, 2))
  } catch (error) {
    console.error("Error trying to put record:")
    throw error
  }

  try {
    console.log("Attempting to query record again")
    const res = await documentClient.query({
      TableName: TABLE_NAME,
      KeyConditionExpression: "#p = :i",
      ExpressionAttributeNames: {
        "#p": "pk",
      },
      ExpressionAttributeValues: {
        ":i": "prim key",
      }
    })
    console.log('Get record response:')
    console.log(JSON.stringify(res, null, 2))
  } catch (error) {
    console.error("Error trying to put record:")
    throw error
  }

  try {
    console.log("Attempting to put record")
    const res = documentClient.put({
      TableName: TABLE_NAME_2,
      Item: {
        pk: 'prim key',
        sk: 'a sort key',
        bool: true,
        numberArray: [1, 2],
        stringArray: ['hey', 'ho'],
        nestedJSON: {
          hey: 'ho',
          lets: [1, 2, 3]
        },
        GSI1PK: "yeye"
      }
    })
    console.log('Put record response:')
    console.log(JSON.stringify(res, null, 2))
  } catch (error) {
    console.error("Error trying to put record:")
    throw error
  }

  try {
    console.log("Attempting to query record")
    const res = await documentClient.query({
      TableName: TABLE_NAME_2,
      KeyConditionExpression: "#p = :i AND #s >= :o",
      ExpressionAttributeNames: {
        "#p": "GSI1PK",
        "#s": "sk"
      },
      ExpressionAttributeValues: {
        ":i": "yeye",
        ":o": "a sort"
      },
      IndexName: "test_index_1"
    })
    console.log('Get record response:')
    console.log(JSON.stringify(res, null, 2))
  } catch (error) {
    console.error("Error trying to put record:")
    throw error
  }

  // try {
  //   console.log('Trying to create table with numbers and secondary indexes...')
  //   const cmd = new CreateTableCommand({
  //     TableName: TABLE_NAME,
  //     KeySchema: [
  //       {
  //         AttributeName: "pk'",
  //         KeyType: "HASH"
  //       },
  //       {
  //         AttributeName: "sk",
  //         KeyType: "RANGE"
  //       },
  //     ],
  //     AttributeDefinitions: [
  //       {
  //         AttributeName: "pk'",
  //         AttributeType: "S"
  //       },
  //       {
  //         AttributeName: "sk",
  //         AttributeType: "N"
  //       },
  //     ],
  //     GlobalSecondaryIndexes: [
  //       {
  //         IndexName: 'TestIndex', // This gets ignored in queries
  //         KeySchema: [
  //           {
  //             AttributeName: "pk'",
  //             KeyType: "RANGE"
  //           },
  //           {
  //             AttributeName: "sk",
  //             KeyType: "HASH"
  //           }
  //         ],
  //         Projection: {
  //           ProjectionType: 'ALL' // This gets ignored
  //         }
  //       }
  //     ]
  //   })
  //   const res = client.send(cmd)
  //   console.log('Create table response:')
  //   console.log(JSON.stringify(res, null, 2))
  // } catch (error) {
  //   console.error("Error trying to create table:")
  //   throw error
  // }
}

main()
