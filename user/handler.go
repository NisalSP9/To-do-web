package user

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"github.com/NisalSP9/To-Do-Web/connection"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"

	"github.com/google/uuid"
)

type Handler struct{}

const TABLE_NAME string = "Users"

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	type parameters struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	params := parameters{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Println("User ", "Create User", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(err.Error()); err != nil {
			log.Println("User ", "Create User ", err.Error())
		}
		return
	}

	if len(params.Username) <= 5 || len(params.Password) <= 5 {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode("Too short username or password, use at least 6 characters !!!"); err != nil {
			log.Println("User ", "Create User ", err.Error())
		}
		return
	}
	// Create DynamoDB client
	svc := connection.GetSVC()
	// Check if username already exists
	usernameKeyCondition := expression.Key("username").Equal(expression.Value(params.Username))
	expr, err := expression.NewBuilder().WithKeyCondition(usernameKeyCondition).Build()
	if err != nil {
		log.Fatalf("Got error building expression: %s", err)
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(TABLE_NAME),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Limit:                     aws.Int64(1), // We only need to know if at least one item exists
	}

	queryResult, err := svc.Query(queryInput)
	if err != nil {
		log.Fatalf("Got error calling Query: %s", err)
	}

	if *queryResult.Count > 0 {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode("Username already exist !!!"); err != nil {
			log.Println("User ", "Create User ", err.Error())
		}
		return
	}

	key := make([]byte, 32) // 32 bytes = 256 bits

	_, err = rand.Read(key)
	if err != nil {
		log.Fatalf("failed to generate API key: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(err.Error()); err != nil {
			log.Println("User ", "Create User ", err.Error())
		}
		return
	}

	user := User{UserID: uuid.New(), Username: params.Username, Password: params.Password, APIKey: base64.URLEncoding.EncodeToString(key)}

	userav, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		log.Fatalf("Got error marshalling new user: %s", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      userav,
		TableName: aws.String(TABLE_NAME),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		log.Fatalf("Got error calling PutItem: %s", err)
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode("user created!"); err != nil {
		log.Println("User ", "Create User ", err.Error())
	}

}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	type parameters struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	params := parameters{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Println("User ", "Login ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(err.Error()); err != nil {
			log.Println("User ", "Login ", err.Error())
		}
		return
	}

	if len(params.Username) <= 1 || len(params.Password) <= 1 {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode("Too short username or password !!!"); err != nil {
			log.Println("User ", "Create User ", err.Error())
		}
		return
	}

	// Create DynamoDB client
	svc := connection.GetSVC()

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"username": {
				S: aws.String(params.Username),
			},
		},
	})
	if err != nil {
		log.Println("User ", "Login ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(err.Error()); err != nil {
			log.Println("User ", "Login ", err.Error())
		}
		return
	}

	user := User{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		log.Println("User ", "Login ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(err.Error()); err != nil {
			log.Println("User ", "Login ", err.Error())
		}
		return
	}

	if user.Password != params.Password {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode("Wrong password or username !!!"); err != nil {
			log.Println("User ", "Login ", err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user.APIKey); err != nil {
		log.Println("User ", "Login ", err.Error())
	}

}

func (h *Handler) Auth(token string) (uuid.UUID, error) {

	// Create DynamoDB client
	svc := connection.GetSVC()

	indexName := "apiKey-index" // Replace with your GSI name

	// Query using the GSI
	input := &dynamodb.QueryInput{
		TableName:              aws.String(TABLE_NAME),
		IndexName:              aws.String(indexName),
		KeyConditionExpression: aws.String("apiKey = :token"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":token": {
				S: aws.String(token),
			},
		},
	}

	result, err := svc.Query(input)
	if err != nil {
		log.Println("User ", "Auth ", err.Error())
		return uuid.Nil, err
	}

	if len(result.Items) == 0 {
		log.Println("User ", "Auth ", "No user found with the given API key")
		return uuid.Nil, nil
	}

	user := User{}
	err = dynamodbattribute.UnmarshalMap(result.Items[0], &user)
	if err != nil {
		log.Println("User ", "Auth ", err.Error())
		return uuid.Nil, err
	}

	return user.UserID, nil
}
