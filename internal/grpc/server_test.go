package grpc_test

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/garnizeh/englog/internal/grpc"
	workerpb "github.com/garnizeh/englog/proto/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestServer_NewServer tests the server constructor
func TestServer_NewServer(t *testing.T) {
	t.Run("creates new server successfully", func(t *testing.T) {
		cfg := createTestConfigForManager()
		logger := createTestLoggerForManager()

		server := grpc.NewServer(cfg, logger)

		assert.NotNil(t, server)
	})
}

// TestServer_RegisterWorker tests worker registration functionality
func TestServer_RegisterWorker(t *testing.T) {
	cfg := createTestConfigForManager()
	logger := createTestLoggerForManager()
	server := grpc.NewServer(cfg, logger)

	tests := []struct {
		name     string
		req      *workerpb.RegisterWorkerRequest
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "valid worker registration",
			req: &workerpb.RegisterWorkerRequest{
				WorkerId:   "worker-001",
				WorkerName: "Test Worker",
				Capabilities: []workerpb.WorkerCapability{
					workerpb.WorkerCapability_CAPABILITY_AI_INSIGHTS,
				},
				Version: "1.0.0",
				Metadata: map[string]string{
					"env": "test",
				},
			},
			wantErr:  false,
			wantCode: codes.OK,
		},
		{
			name: "missing worker ID",
			req: &workerpb.RegisterWorkerRequest{
				WorkerId:   "",
				WorkerName: "Test Worker",
				Capabilities: []workerpb.WorkerCapability{
					workerpb.WorkerCapability_CAPABILITY_AI_INSIGHTS,
				},
				Version: "1.0.0",
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
		{
			name: "missing worker name",
			req: &workerpb.RegisterWorkerRequest{
				WorkerId:   "worker-001",
				WorkerName: "",
				Capabilities: []workerpb.WorkerCapability{
					workerpb.WorkerCapability_CAPABILITY_AI_INSIGHTS,
				},
				Version: "1.0.0",
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
		{
			name: "re-registration of existing worker",
			req: &workerpb.RegisterWorkerRequest{
				WorkerId:   "worker-001",
				WorkerName: "Test Worker Updated",
				Capabilities: []workerpb.WorkerCapability{
					workerpb.WorkerCapability_CAPABILITY_WEEKLY_REPORTS,
				},
				Version: "1.1.0",
			},
			wantErr:  false,
			wantCode: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := server.RegisterWorker(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)

				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.SessionToken)
				assert.True(t, resp.RegistrationSuccessful)
				assert.Equal(t, int32(30), resp.HeartbeatIntervalSeconds)
			}
		})
	}
}

// TestServer_WorkerHeartbeat tests worker heartbeat functionality
func TestServer_WorkerHeartbeat(t *testing.T) {
	cfg := createTestConfig()
	logger := createTestLoggerForManager()
	server := grpc.NewServer(cfg, logger)

	// First register a worker
	ctx := context.Background()
	registerReq := &workerpb.RegisterWorkerRequest{
		WorkerId:   "worker-001",
		WorkerName: "Test Worker",
		Capabilities: []workerpb.WorkerCapability{
			workerpb.WorkerCapability_CAPABILITY_AI_INSIGHTS,
		},
		Version: "1.0.0",
	}

	registerResp, err := server.RegisterWorker(ctx, registerReq)
	require.NoError(t, err)
	require.NotNil(t, registerResp)

	tests := []struct {
		name     string
		req      *workerpb.WorkerHeartbeatRequest
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "valid heartbeat",
			req: &workerpb.WorkerHeartbeatRequest{
				WorkerId:     "worker-001",
				SessionToken: registerResp.SessionToken,
				Status:       workerpb.WorkerStatus_WORKER_STATUS_IDLE,
				Stats: &workerpb.WorkerStats{
					ActiveTasks:    0,
					CompletedTasks: 5,
					FailedTasks:    1,
					CpuUsage:       45.2,
					MemoryUsage:    67.8,
					Uptime:         timestamppb.New(time.Now().Add(-time.Hour)),
				},
			},
			wantErr:  false,
			wantCode: codes.OK,
		},
		{
			name: "worker not found",
			req: &workerpb.WorkerHeartbeatRequest{
				WorkerId:     "worker-999",
				SessionToken: "invalid-token",
				Status:       workerpb.WorkerStatus_WORKER_STATUS_IDLE,
			},
			wantErr:  true,
			wantCode: codes.NotFound,
		},
		{
			name: "invalid session token",
			req: &workerpb.WorkerHeartbeatRequest{
				WorkerId:     "worker-001",
				SessionToken: "invalid-token",
				Status:       workerpb.WorkerStatus_WORKER_STATUS_IDLE,
			},
			wantErr:  true,
			wantCode: codes.Unauthenticated,
		},
		{
			name: "status change heartbeat",
			req: &workerpb.WorkerHeartbeatRequest{
				WorkerId:     "worker-001",
				SessionToken: registerResp.SessionToken,
				Status:       workerpb.WorkerStatus_WORKER_STATUS_BUSY,
				Stats: &workerpb.WorkerStats{
					ActiveTasks:    2,
					CompletedTasks: 5,
					FailedTasks:    1,
				},
			},
			wantErr:  false,
			wantCode: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.WorkerHeartbeat(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)

				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.True(t, resp.ConnectionHealthy)
				assert.NotNil(t, resp.ServerTime)
			}
		})
	}
}

// TestServer_QueueTask tests task queuing functionality
func TestServer_QueueTask(t *testing.T) {
	ctx := context.Background()
	cfg := createTestConfig()
	logger := createTestLoggerForManager()
	server := grpc.NewServer(cfg, logger)

	tests := []struct {
		name    string
		task    *workerpb.TaskRequest
		wantErr bool
	}{
		{
			name: "valid task",
			task: &workerpb.TaskRequest{
				TaskId:   "task-001",
				TaskType: workerpb.TaskType_TASK_TYPE_INSIGHT_GENERATION,
				Payload:  `{"user_id": "user-123", "data": "test"}`,
				Priority: 5,
				Deadline: timestamppb.New(time.Now().Add(time.Hour)),
				Metadata: map[string]string{
					"user_id": "user-123",
				},
			},
			wantErr: false,
		},
		{
			name: "missing task ID",
			task: &workerpb.TaskRequest{
				TaskId:   "",
				TaskType: workerpb.TaskType_TASK_TYPE_INSIGHT_GENERATION,
				Payload:  `{"user_id": "user-123"}`,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.QueueTask(ctx, tt.task)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestServer_StreamTasks tests task streaming functionality
func TestServer_StreamTasks(t *testing.T) {
	cfg := createTestConfig()
	logger := createTestLoggerForManager()
	server := grpc.NewServer(cfg, logger)

	// Register a worker first
	ctx := context.Background()
	registerReq := &workerpb.RegisterWorkerRequest{
		WorkerId:   "worker-001",
		WorkerName: "Test Worker",
		Capabilities: []workerpb.WorkerCapability{
			workerpb.WorkerCapability_CAPABILITY_AI_INSIGHTS,
		},
		Version: "1.0.0",
	}

	registerResp, err := server.RegisterWorker(ctx, registerReq)
	require.NoError(t, err)

	t.Run("worker not found", func(t *testing.T) {
		req := &workerpb.StreamTasksRequest{
			WorkerId:     "worker-999",
			SessionToken: "invalid-token",
		}

		mockStream := NewMockStream(ctx)
		err := server.StreamTasks(req, mockStream)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("invalid session token", func(t *testing.T) {
		req := &workerpb.StreamTasksRequest{
			WorkerId:     "worker-001",
			SessionToken: "invalid-token",
		}

		mockStream := NewMockStream(ctx)
		err := server.StreamTasks(req, mockStream)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("successful task streaming with context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)

		req := &workerpb.StreamTasksRequest{
			WorkerId:     "worker-001",
			SessionToken: registerResp.SessionToken,
			Capabilities: []workerpb.WorkerCapability{
				workerpb.WorkerCapability_CAPABILITY_AI_INSIGHTS,
			},
		}

		mockStream := NewMockStream(cancelCtx)

		// Start streaming in a goroutine
		var streamErr error
		done := make(chan bool)
		go func() {
			streamErr = server.StreamTasks(req, mockStream)
			done <- true
		}()

		// Give it a moment to establish the stream
		time.Sleep(10 * time.Millisecond)

		// Cancel the context to simulate disconnect
		cancel()

		// Wait for the stream to finish
		<-done

		assert.Error(t, streamErr)
		assert.Equal(t, context.Canceled, streamErr)
	})
}

// TestServer_ReportTaskResult tests task result reporting
func TestServer_ReportTaskResult(t *testing.T) {
	cfg := createTestConfig()
	logger := createTestLoggerForManager()
	server := grpc.NewServer(cfg, logger)

	now := time.Now()

	tests := []struct {
		name     string
		req      *workerpb.TaskResultRequest
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "successful task completion",
			req: &workerpb.TaskResultRequest{
				TaskId:      "task-001",
				WorkerId:    "worker-001",
				Status:      workerpb.TaskStatus_TASK_STATUS_COMPLETED,
				Result:      `{"insight": "User productivity increased", "confidence": 0.85}`,
				StartedAt:   timestamppb.New(now.Add(-time.Minute)),
				CompletedAt: timestamppb.New(now),
			},
			wantErr:  false,
			wantCode: codes.OK,
		},
		{
			name: "failed task",
			req: &workerpb.TaskResultRequest{
				TaskId:       "task-002",
				WorkerId:     "worker-001",
				Status:       workerpb.TaskStatus_TASK_STATUS_FAILED,
				ErrorMessage: "Failed to process data",
				StartedAt:    timestamppb.New(now.Add(-time.Minute)),
				CompletedAt:  timestamppb.New(now),
			},
			wantErr:  false,
			wantCode: codes.OK,
		},
		{
			name: "missing task ID",
			req: &workerpb.TaskResultRequest{
				TaskId:   "",
				WorkerId: "worker-001",
				Status:   workerpb.TaskStatus_TASK_STATUS_COMPLETED,
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
		{
			name: "missing worker ID",
			req: &workerpb.TaskResultRequest{
				TaskId:   "task-003",
				WorkerId: "",
				Status:   workerpb.TaskStatus_TASK_STATUS_COMPLETED,
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := server.ReportTaskResult(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)

				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.True(t, resp.ResultReceived)

				// Verify task result is stored
				if tt.req.TaskId != "" {
					result, found := server.GetTaskResult(tt.req.TaskId)
					assert.True(t, found)
					assert.Equal(t, tt.req.TaskId, result.TaskID)
					assert.Equal(t, tt.req.WorkerId, result.WorkerID)
					assert.Equal(t, tt.req.Status, result.Status)
				}
			}
		})
	}
}

// TestServer_UpdateTaskProgress tests task progress updates
func TestServer_UpdateTaskProgress(t *testing.T) {
	cfg := createTestConfig()
	logger := createTestLoggerForManager()
	server := grpc.NewServer(cfg, logger)

	tests := []struct {
		name string
		req  *workerpb.TaskProgressRequest
	}{
		{
			name: "valid progress update",
			req: &workerpb.TaskProgressRequest{
				TaskId:          "task-001",
				WorkerId:        "worker-001",
				ProgressPercent: 50,
				StatusMessage:   "Processing data...",
				UpdatedAt:       timestamppb.Now(),
			},
		},
		{
			name: "progress at 0%",
			req: &workerpb.TaskProgressRequest{
				TaskId:          "task-002",
				WorkerId:        "worker-001",
				ProgressPercent: 0,
				StatusMessage:   "Starting task...",
			},
		},
		{
			name: "progress at 100%",
			req: &workerpb.TaskProgressRequest{
				TaskId:          "task-003",
				WorkerId:        "worker-001",
				ProgressPercent: 100,
				StatusMessage:   "Task completed",
			},
		},
		{
			name: "invalid progress percentage (negative)",
			req: &workerpb.TaskProgressRequest{
				TaskId:          "task-004",
				WorkerId:        "worker-001",
				ProgressPercent: -10,
				StatusMessage:   "Invalid progress",
			},
		},
		{
			name: "invalid progress percentage (over 100)",
			req: &workerpb.TaskProgressRequest{
				TaskId:          "task-005",
				WorkerId:        "worker-001",
				ProgressPercent: 150,
				StatusMessage:   "Invalid progress",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := server.UpdateTaskProgress(ctx, tt.req)

			// Progress updates should always succeed (they're informational)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

// TestServer_HealthCheck tests health check functionality
func TestServer_HealthCheck(t *testing.T) {
	cfg := createTestConfig()
	logger := createTestLoggerForManager()
	server := grpc.NewServer(cfg, logger)

	t.Run("health check with no workers", func(t *testing.T) {
		ctx := context.Background()
		resp, err := server.HealthCheck(ctx, &emptypb.Empty{})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "warning", resp.Status) // Status is warning when no workers are available
		assert.NotNil(t, resp.Timestamp)
		assert.Equal(t, int32(0), resp.ActiveWorkers)
		assert.Contains(t, resp.Services, "grpc_server")
		assert.Equal(t, "healthy", resp.Services["grpc_server"])
		assert.Contains(t, resp.Services, "task_queue")
		assert.Equal(t, "healthy", resp.Services["task_queue"])
		assert.Contains(t, resp.Services, "ollama")
		assert.Equal(t, "unknown", resp.Services["ollama"]) // No workers means no Ollama status
		assert.Contains(t, resp.Services, "worker_connections")
		assert.Equal(t, "no_workers", resp.Services["worker_connections"])
	})

	t.Run("health check with registered workers", func(t *testing.T) {
		ctx := context.Background()

		// Register some workers
		for i := 1; i <= 3; i++ {
			registerReq := &workerpb.RegisterWorkerRequest{
				WorkerId:   fmt.Sprintf("worker-%03d", i),
				WorkerName: fmt.Sprintf("Test Worker %d", i),
				Capabilities: []workerpb.WorkerCapability{
					workerpb.WorkerCapability_CAPABILITY_AI_INSIGHTS,
				},
				Version: "1.0.0",
			}

			_, err := server.RegisterWorker(ctx, registerReq)
			require.NoError(t, err)
		}

		resp, err := server.HealthCheck(ctx, &emptypb.Empty{})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "healthy", resp.Status)
		assert.Equal(t, int32(3), resp.ActiveWorkers)
	})
}

// TestServer_HelperMethods tests helper methods
func TestServer_HelperMethods(t *testing.T) {
	ctx := context.Background()
	cfg := createTestConfig()
	logger := createTestLoggerForManager()
	server := grpc.NewServer(cfg, logger)

	t.Run("GetActiveWorkers with no workers", func(t *testing.T) {
		workers := server.GetActiveWorkers(ctx)
		assert.Empty(t, workers)
	})

	t.Run("GetActiveWorkers with registered workers", func(t *testing.T) {
		// Register a worker
		registerReq := &workerpb.RegisterWorkerRequest{
			WorkerId:   "worker-001",
			WorkerName: "Test Worker",
			Capabilities: []workerpb.WorkerCapability{
				workerpb.WorkerCapability_CAPABILITY_AI_INSIGHTS,
			},
			Version: "1.0.0",
		}

		_, err := server.RegisterWorker(ctx, registerReq)
		require.NoError(t, err)

		workers := server.GetActiveWorkers(ctx)
		assert.Len(t, workers, 1)
		assert.Contains(t, workers, "worker-001")
	})

	t.Run("GetTaskResult for non-existent task", func(t *testing.T) {
		result, found := server.GetTaskResult("non-existent-task")
		assert.False(t, found)
		assert.Nil(t, result)
	})
}

// TestServer_Start tests server startup (basic functionality)
func TestServer_Start(t *testing.T) {
	ctx := context.Background()
	cfg := createTestConfig()
	logger := createTestLoggerForManager()
	server := grpc.NewServer(cfg, logger)

	t.Run("start with invalid address", func(t *testing.T) {
		err := server.Start(ctx, "invalid-address")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to listen")
	})

	t.Run("start with available port", func(t *testing.T) {
		// Find an available port
		listener, err := net.Listen("tcp", ":0")
		require.NoError(t, err)

		port := listener.Addr().(*net.TCPAddr).Port
		listener.Close()

		// Start server in goroutine since it blocks
		done := make(chan error, 1)
		go func() {
			done <- server.Start(ctx, fmt.Sprintf(":%d", port))
		}()

		// Give it a moment to start
		time.Sleep(10 * time.Millisecond)

		// The test passes if no immediate error occurs
		select {
		case err := <-done:
			t.Errorf("Server exited unexpectedly: %v", err)
		default:
			// Server is running, which is expected
		}
	})
}

// TestServer_TaskQueueFull tests task queue overflow
func TestServer_TaskQueueFull(t *testing.T) {
	ctx := context.Background()
	cfg := createTestConfig()
	logger := createTestLoggerForManager()
	server := grpc.NewServer(cfg, logger)

	// Fill the task queue (buffer is 100)
	for i := 0; i < 100; i++ {
		task := &workerpb.TaskRequest{
			TaskId:   fmt.Sprintf("task-%03d", i),
			TaskType: workerpb.TaskType_TASK_TYPE_INSIGHT_GENERATION,
			Payload:  fmt.Sprintf(`{"task_num": %d}`, i),
		}
		err := server.QueueTask(ctx, task)
		assert.NoError(t, err)
	}

	// Try to add one more task (should fail)
	overflowTask := &workerpb.TaskRequest{
		TaskId:   "overflow-task",
		TaskType: workerpb.TaskType_TASK_TYPE_INSIGHT_GENERATION,
		Payload:  `{"overflow": true}`,
	}

	err := server.QueueTask(ctx, overflowTask)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue is full")
}

// TestServer_ConcurrentAccess tests concurrent access to server methods
func TestServer_ConcurrentAccess(t *testing.T) {
	cfg := createTestConfig()
	logger := createTestLoggerForManager()
	server := grpc.NewServer(cfg, logger)

	const numWorkers = 10
	const numOperations = 50

	var wg sync.WaitGroup

	// Test concurrent worker registrations
	t.Run("concurrent worker registrations", func(t *testing.T) {
		ctx := context.Background()

		wg.Add(numWorkers)

		for i := 0; i < numWorkers; i++ {
			go func(workerID int) {
				defer wg.Done()

				ctx := context.Background()
				req := &workerpb.RegisterWorkerRequest{
					WorkerId:   fmt.Sprintf("concurrent-worker-%d", workerID),
					WorkerName: fmt.Sprintf("Concurrent Worker %d", workerID),
					Capabilities: []workerpb.WorkerCapability{
						workerpb.WorkerCapability_CAPABILITY_AI_INSIGHTS,
					},
					Version: "1.0.0",
				}

				resp, err := server.RegisterWorker(ctx, req)
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}(i)
		}

		wg.Wait()

		// Verify all workers are registered
		workers := server.GetActiveWorkers(ctx)
		assert.Len(t, workers, numWorkers)
	})

	// Test concurrent task queueing
	t.Run("concurrent task queueing", func(t *testing.T) {
		wg.Add(numOperations)

		for i := 0; i < numOperations; i++ {
			go func(taskID int) {
				defer wg.Done()

				ctx := context.Background()
				task := &workerpb.TaskRequest{
					TaskId:   fmt.Sprintf("concurrent-task-%d", taskID),
					TaskType: workerpb.TaskType_TASK_TYPE_INSIGHT_GENERATION,
					Payload:  fmt.Sprintf(`{"task_id": %d}`, taskID),
				}

				err := server.QueueTask(ctx, task)
				assert.NoError(t, err)
			}(i)
		}

		wg.Wait()
	})
}
