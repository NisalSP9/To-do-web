package task

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/NisalSP9/To-Do-Web/connection"
	"github.com/NisalSP9/To-Do-Web/middleware"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/google/uuid"
)

type Handler struct{}

const TABLE_NAME string = "Tasks"

// In-memory cache
var taskCache = struct {
	sync.RWMutex
	tasks map[uuid.UUID]Task
}{tasks: make(map[uuid.UUID]Task)}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	type parameters struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	params := parameters{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Println("Task ", "Create Task", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(err.Error()); err != nil {
			log.Println("Task ", "Create Task ", err.Error())
		}
		return
	}

	if params.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode("Task's title can't be empty !!!"); err != nil {
			log.Println("Task ", "Create Task ", err.Error())
		}
		return
	}

	// Create DynamoDB client
	svc := connection.GetSVC()
	userID, err := middleware.GetAuthUserID(r)
	if err != nil {
		log.Fatalf("Got error getting user id : %s", err)
	}

	taskID := uuid.New()

	task := Task{TaskID: taskID, UserID: userID, Title: params.Title, Description: params.Description, Status: "To Do"}

	taskav, err := dynamodbattribute.MarshalMap(task)
	if err != nil {
		log.Fatalf("Got error marshalling new task: %s", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      taskav,
		TableName: aws.String(TABLE_NAME),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		log.Fatalf("Got error calling PutItem: %s", err)
	}

	// Update cache
	taskCache.Lock()
	taskCache.tasks[taskID] = task
	taskCache.Unlock()

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode("task created!"); err != nil {
		log.Println("Task ", "Create Task ", err.Error())
	}
}

func (h *Handler) GetTasksByUserID(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	userID, err := middleware.GetAuthUserID(r)
	if err != nil {
		log.Fatalf("Got error getting user id : %s", err)
	}

	tasks := []Task{}
	ch := make(chan Task)
	errCh := make(chan error)

	// Check the cache first
	taskCache.RLock()
	for _, task := range taskCache.tasks {
		if task.UserID == userID {
			tasks = append(tasks, task)
		}
	}
	taskCache.RUnlock()

	// If tasks found in cache, return them
	if len(tasks) > 0 {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(tasks); err != nil {
			log.Println("Task ", "GetTasksByUserID ", err.Error())
		}
		return
	}

	// Fetch from DynamoDB if not found in cache
	go func() {
		log.Println("Fetch from DynamoDB if not found in cache")
		svc := connection.GetSVC()
		filt := expression.Name("userID").Equal(expression.Value(userID))

		proj := expression.NamesList(expression.Name("taskID"), expression.Name("userID"), expression.Name("title"), expression.Name("description"), expression.Name("status"))

		expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
		if err != nil {
			errCh <- err
			return
		}

		params := &dynamodb.ScanInput{
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			FilterExpression:          expr.Filter(),
			ProjectionExpression:      expr.Projection(),
			TableName:                 aws.String(TABLE_NAME),
		}

		result, err := svc.Scan(params)
		if err != nil {
			errCh <- err
			return
		}

		for _, i := range result.Items {
			task := Task{}
			err = dynamodbattribute.UnmarshalMap(i, &task)
			if err != nil {
				errCh <- err
				return
			}

			// Update cache
			taskCache.Lock()
			taskCache.tasks[task.TaskID] = task
			taskCache.Unlock()

			ch <- task
		}

		close(ch)
		close(errCh)
	}()

	// Collect tasks from the channel
	for task := range ch {
		tasks = append(tasks, task)
	}

	// Check for errors
	select {
	case err := <-errCh:
		if err != nil {
			http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
			return
		}
	default:
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Println("Task ", "GetTasksByUserID ", err.Error())
	}

}

func (h *Handler) EditTaskDetails(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	task := Task{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&task)
	if err != nil {
		log.Println("Task ", "Create Task", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(err.Error()); err != nil {
			log.Println("Task ", "Create Task ", err.Error())
		}
		return
	}

	svc := connection.GetSVC()

	updateExpression := "SET #title = :title, #description = :description, #status = :status"
	expressionAttributeNames := map[string]*string{
		"#title":       aws.String("title"),
		"#description": aws.String("description"),
		"#status":      aws.String("status"),
	}
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":title":       {S: aws.String(task.Title)},
		":description": {S: aws.String(task.Description)},
		":status":      {S: aws.String(task.Status)},
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"taskID": {B: task.TaskID[:]},
			"userID": {B: task.UserID[:]},
		},
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		ReturnValues:              aws.String(dynamodb.ReturnValueUpdatedNew),
	}

	_, err = svc.UpdateItem(input)
	if err != nil {
		log.Println("failed to update task: %w", err)
	}

	// Update cache
	taskCache.Lock()
	taskCache.tasks[task.TaskID] = task
	taskCache.Unlock()

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode("Updated !!!"); err != nil {
		log.Println("Task ", "GetTasksByUserID ", err.Error())
	}

}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	taskIDStr := r.PathValue("taskID")

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		http.Error(w, "invalid taskID", http.StatusBadRequest)
		return
	}

	userID, err := middleware.GetAuthUserID(r)
	if err != nil {
		log.Fatalf("Got error getting user id : %s", err)
	}

	svc := connection.GetSVC()

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"taskID": {B: taskID[:]},
			"userID": {B: userID[:]},
		},
		TableName: aws.String(TABLE_NAME),
	}

	_, err = svc.DeleteItem(input)
	if err != nil {
		log.Printf("Got error calling Delete task: %s", err)
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	// Remove from cache
	taskCache.Lock()
	delete(taskCache.tasks, taskID)
	taskCache.Unlock()

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode("Deleted !!!"); err != nil {
		log.Println("Task ", "DeleteTask ", err.Error())
	}
}
