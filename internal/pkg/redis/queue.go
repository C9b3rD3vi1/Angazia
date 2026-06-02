package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

type Job struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Attempts  int                    `json:"attempts"`
	MaxRetries int                   `json:"max_retries"`
	CreatedAt time.Time              `json:"created_at"`
	ScheduledAt *time.Time           `json:"scheduled_at,omitempty"`
}

type JobHandler func(ctx context.Context, job Job) error

type JobQueue struct {
	client    *RedisClient
	handlers  map[string]JobHandler
	mu        sync.RWMutex
	prefix    string
	consumers int
	done      chan struct{}
}

func NewJobQueue(client *RedisClient, consumers int) *JobQueue {
	if consumers <= 0 {
		consumers = 1
	}
	return &JobQueue{
		client:    client,
		handlers:  make(map[string]JobHandler),
		prefix:    "queue:",
		consumers: consumers,
		done:      make(chan struct{}),
	}
}

func (q *JobQueue) RegisterHandler(jobType string, handler JobHandler) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[jobType] = handler
}

func (q *JobQueue) Enqueue(ctx context.Context, jobType string, payload map[string]interface{}, opts ...JobOption) error {
	job := Job{
		ID:         generateJobID(),
		Type:       jobType,
		Payload:    payload,
		Attempts:   0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}

	for _, opt := range opts {
		opt(&job)
	}

	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	key := q.prefix + jobType

	if job.ScheduledAt != nil && job.ScheduledAt.After(time.Now()) {
		return q.client.ZAdd(ctx, q.prefix+"scheduled", float64(job.ScheduledAt.Unix()), string(data))
	}

	return q.client.LPush(ctx, key, string(data))
}

func (q *JobQueue) Dequeue(ctx context.Context, jobType string, timeout time.Duration) (*Job, error) {
	key := q.prefix + jobType
	result, err := q.client.BRPop(ctx, timeout, key)
	if err != nil {
		return nil, err
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid queue response")
	}

	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, err
	}

	return &job, nil
}

func (q *JobQueue) Start(ctx context.Context) {
	q.moveScheduledJobs(ctx)

	for i := 0; i < q.consumers; i++ {
		go q.consumer(ctx, i)
	}

	go q.scheduler(ctx)

	log.Printf("Job queue started with %d consumers", q.consumers)
}

func (q *JobQueue) Stop() {
	close(q.done)
}

func (q *JobQueue) consumer(ctx context.Context, id int) {
	for {
		select {
		case <-q.done:
			return
		default:
			q.processNext(ctx, id)
		}
	}
}

func (q *JobQueue) processNext(ctx context.Context, id int) {
	q.mu.RLock()
	jobTypes := make([]string, 0, len(q.handlers))
	for jobType := range q.handlers {
		jobTypes = append(jobTypes, q.prefix+jobType)
	}
	q.mu.RUnlock()

	if len(jobTypes) == 0 {
		time.Sleep(1 * time.Second)
		return
	}

	result, err := q.client.BRPop(ctx, time.Second, jobTypes...)
	if err != nil {
		return
	}

	if len(result) < 2 {
		return
	}

	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		log.Printf("Consumer %d: failed to unmarshal job: %v", id, err)
		return
	}

	q.mu.RLock()
	handler, ok := q.handlers[job.Type]
	q.mu.RUnlock()

	if !ok {
		log.Printf("Consumer %d: no handler for job type %s", id, job.Type)
		return
	}

	if err := handler(ctx, job); err != nil {
		job.Attempts++
		log.Printf("Consumer %d: job %s failed (attempt %d/%d): %v", id, job.ID, job.Attempts, job.MaxRetries, err)

		if job.Attempts < job.MaxRetries {
			backoff := time.Duration(1<<uint(job.Attempts-1)) * time.Second
			if backoff > 60*time.Second {
				backoff = 60 * time.Second
			}
			scheduledAt := time.Now().Add(backoff)
			job.ScheduledAt = &scheduledAt
			data, _ := json.Marshal(job)
			q.client.ZAdd(ctx, q.prefix+"retry", float64(scheduledAt.Unix()), string(data))
		} else {
			data, _ := json.Marshal(job)
			q.client.LPush(ctx, q.prefix+"dead", string(data))
			log.Printf("Consumer %d: job %s moved to dead letter queue", id, job.ID)
		}
	}
}

func (q *JobQueue) scheduler(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-q.done:
			return
		case <-ticker.C:
			q.moveScheduledJobs(ctx)
		}
	}
}

func (q *JobQueue) moveScheduledJobs(ctx context.Context) {
	now := time.Now().Unix()
	results, err := q.client.ZRangeByScore(ctx, q.prefix+"scheduled", 0, float64(now))
	if err != nil {
		return
	}

	for _, result := range results {
		var job Job
		if err := json.Unmarshal([]byte(result), &job); err != nil {
			continue
		}
		q.client.ZRem(ctx, q.prefix+"scheduled", result)
		data, _ := json.Marshal(job)
		q.client.LPush(ctx, q.prefix+job.Type, string(data))
	}

	retryResults, err := q.client.ZRangeByScore(ctx, q.prefix+"retry", 0, float64(now))
	if err != nil {
		return
	}

	for _, result := range retryResults {
		var job Job
		if err := json.Unmarshal([]byte(result), &job); err != nil {
			continue
		}
		q.client.ZRem(ctx, q.prefix+"retry", result)
		data, _ := json.Marshal(job)
		q.client.LPush(ctx, q.prefix+job.Type, string(data))
	}
}

type JobOption func(*Job)

func WithMaxRetries(retries int) JobOption {
	return func(j *Job) {
		j.MaxRetries = retries
	}
}

func WithScheduledAt(t time.Time) JobOption {
	return func(j *Job) {
		j.ScheduledAt = &t
	}
}

var jobIDCounter int64
var jobIDMu sync.Mutex

func generateJobID() string {
	jobIDMu.Lock()
	defer jobIDMu.Unlock()
	jobIDCounter++
	now := time.Now().UnixNano()
	return fmt.Sprintf("%d-%d", now, jobIDCounter)
}
