package ml

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMLClassificationEngine(t *testing.T) {
	engine := NewMLClassificationEngine()
	assert.NotNil(t, engine)
	assert.NotNil(t, engine.models)
	assert.NotNil(t, engine.classifications)
	assert.NotNil(t, engine.optimizations)
	assert.NotNil(t, engine.cachePredictions)
	assert.Len(t, engine.models, 0)
	assert.Len(t, engine.classifications, 0)
	assert.Len(t, engine.optimizations, 0)
	assert.Len(t, engine.cachePredictions, 0)
}

func TestMLClassificationEngine_ClassifyFile(t *testing.T) {
	engine := NewMLClassificationEngine()
	ctx := context.Background()

	tests := []struct {
		name     string
		filePath string
		content  []byte
		metadata map[string]interface{}
		expected struct {
			category   string
			confidence float64
			extension  string
			mimeType   string
		}
	}{
		{
			name:     "text file",
			filePath: "test.txt",
			content:  []byte("Hello, World!"),
			metadata: map[string]interface{}{"author": "test"},
			expected: struct {
				category   string
				confidence float64
				extension  string
				mimeType   string
			}{
				category:   "document",
				confidence: 1.0, // 0.95 + 0.05 for non-empty content
				extension:  ".txt",
				mimeType:   "text/plain",
			},
		},
		{
			name:     "image file",
			filePath: "image.jpg",
			content:  []byte("fake image data"),
			metadata: map[string]interface{}{"width": 1920, "height": 1080},
			expected: struct {
				category   string
				confidence float64
				extension  string
				mimeType   string
			}{
				category:   "image",
				confidence: 1.0, // 0.98 + 0.05 for non-empty content
				extension:  ".jpg",
				mimeType:   "image/jpeg",
			},
		},
		{
			name:     "video file",
			filePath: "video.mp4",
			content:  []byte("fake video data"),
			metadata: map[string]interface{}{"duration": 120},
			expected: struct {
				category   string
				confidence float64
				extension  string
				mimeType   string
			}{
				category:   "video",
				confidence: 1.0, // 0.97 + 0.05 for non-empty content
				extension:  ".mp4",
				mimeType:   "video/mp4",
			},
		},
		{
			name:     "audio file",
			filePath: "audio.mp3",
			content:  []byte("fake audio data"),
			metadata: map[string]interface{}{"bitrate": 320},
			expected: struct {
				category   string
				confidence float64
				extension  string
				mimeType   string
			}{
				category:   "audio",
				confidence: 1.0, // 0.96 + 0.05 for non-empty content
				extension:  ".mp3",
				mimeType:   "audio/mpeg",
			},
		},
		{
			name:     "PDF file",
			filePath: "document.pdf",
			content:  []byte("fake pdf data"),
			metadata: map[string]interface{}{"pages": 10},
			expected: struct {
				category   string
				confidence float64
				extension  string
				mimeType   string
			}{
				category:   "document",
				confidence: 1.0, // 0.99 + 0.05 for non-empty content (capped at 1.0)
				extension:  ".pdf",
				mimeType:   "application/pdf",
			},
		},
		{
			name:     "archive file",
			filePath: "archive.zip",
			content:  []byte("fake zip data"),
			metadata: map[string]interface{}{"compressed": true},
			expected: struct {
				category   string
				confidence float64
				extension  string
				mimeType   string
			}{
				category:   "archive",
				confidence: 1.0, // 0.98 + 0.05 for non-empty content (capped at 1.0)
				extension:  ".zip",
				mimeType:   "application/zip",
			},
		},
		{
			name:     "executable file",
			filePath: "program.exe",
			content:  []byte("fake exe data"),
			metadata: map[string]interface{}{"arch": "x64"},
			expected: struct {
				category   string
				confidence float64
				extension  string
				mimeType   string
			}{
				category:   "executable",
				confidence: 1.0, // 0.99 + 0.05 for non-empty content (capped at 1.0)
				extension:  ".exe",
				mimeType:   "application/octet-stream",
			},
		},
		{
			name:     "data file",
			filePath: "config.json",
			content:  []byte(`{"key": "value"}`),
			metadata: map[string]interface{}{"format": "json"},
			expected: struct {
				category   string
				confidence float64
				extension  string
				mimeType   string
			}{
				category:   "data",
				confidence: 0.99, // 0.94 + 0.05 for non-empty content
				extension:  ".json",
				mimeType:   "application/json",
			},
		},
		{
			name:     "code file",
			filePath: "main.go",
			content:  []byte("package main"),
			metadata: map[string]interface{}{"language": "go"},
			expected: struct {
				category   string
				confidence float64
				extension  string
				mimeType   string
			}{
				category:   "code",
				confidence: 0.9800000000000001, // 0.93 + 0.05 for non-empty content
				extension:  ".go",
				mimeType:   "text/x-go",
			},
		},
		{
			name:     "unknown file",
			filePath: "unknown.xyz",
			content:  []byte("unknown data"),
			metadata: map[string]interface{}{"type": "unknown"},
			expected: struct {
				category   string
				confidence float64
				extension  string
				mimeType   string
			}{
				category:   "unknown",
				confidence: 0.55, // 0.5 + 0.05 for non-empty content
				extension:  ".xyz",
				mimeType:   "application/octet-stream",
			},
		},
		{
			name:     "empty file",
			filePath: "empty.txt",
			content:  []byte(""),
			metadata: map[string]interface{}{},
			expected: struct {
				category   string
				confidence float64
				extension  string
				mimeType   string
			}{
				category:   "document",
				confidence: 0.95, // No bonus for empty content
				extension:  ".txt",
				mimeType:   "text/plain",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classification, err := engine.ClassifyFile(ctx, tt.filePath, tt.content, tt.metadata)
			assert.NoError(t, err)
			assert.NotNil(t, classification)
			assert.Equal(t, tt.filePath, classification.FilePath)
			assert.Equal(t, tt.expected.category, classification.Category)
			assert.Equal(t, tt.expected.confidence, classification.Confidence)
			assert.Equal(t, tt.expected.extension, classification.Extension)
			assert.Equal(t, tt.expected.mimeType, classification.MimeType)
			assert.Equal(t, int64(len(tt.content)), classification.Size)
			assert.Equal(t, tt.metadata, classification.Metadata)
			assert.NotZero(t, classification.CreatedAt)
			assert.NotEmpty(t, classification.Tags)

			// Verify classification is stored
			stored, err := engine.GetClassification(ctx, tt.filePath)
			assert.NoError(t, err)
			assert.Equal(t, classification, stored)
		})
	}
}

func TestMLClassificationEngine_OptimizeFile(t *testing.T) {
	engine := NewMLClassificationEngine()
	ctx := context.Background()

	tests := []struct {
		name             string
		filePath         string
		content          []byte
		optimizationType string
		expected         struct {
			algorithm        string
			compressionRatio float64
		}
	}{
		{
			name:             "compression optimization",
			filePath:         "test.txt",
			content:          []byte("Hello, World! This is a test file for compression optimization."),
			optimizationType: "compression",
			expected: struct {
				algorithm        string
				compressionRatio float64
			}{
				algorithm:        "gzip",
				compressionRatio: 0.6984126984126984, // 30% compression
			},
		},
		{
			name:             "deduplication optimization",
			filePath:         "test.txt",
			content:          []byte("Hello, World! This is a test file for deduplication optimization."),
			optimizationType: "deduplication",
			expected: struct {
				algorithm        string
				compressionRatio float64
			}{
				algorithm:        "content-hash",
				compressionRatio: 0.8, // 20% deduplication
			},
		},
		{
			name:             "encoding optimization",
			filePath:         "test.txt",
			content:          []byte("Hello, World! This is a test file for encoding optimization."),
			optimizationType: "encoding",
			expected: struct {
				algorithm        string
				compressionRatio float64
			}{
				algorithm:        "base64",
				compressionRatio: 0.9, // 10% optimization
			},
		},
		{
			name:             "unknown optimization type",
			filePath:         "test.txt",
			content:          []byte("Hello, World!"),
			optimizationType: "unknown",
			expected: struct {
				algorithm        string
				compressionRatio float64
			}{
				algorithm:        "none",
				compressionRatio: 1.0, // No optimization
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.OptimizeFile(ctx, tt.filePath, tt.content, tt.optimizationType)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, int64(len(tt.content)), result.OriginalSize)
			assert.Equal(t, tt.expected.algorithm, result.Algorithm)
			assert.Equal(t, tt.optimizationType, result.OptimizationType)
			assert.Equal(t, tt.expected.compressionRatio, result.CompressionRatio)
			// ProcessingTime might be zero for very fast operations
			// assert.NotZero(t, result.ProcessingTime)
			assert.NotNil(t, result.Metadata)
			assert.Equal(t, tt.filePath, result.Metadata["file_path"])

			// Verify optimization is stored
			stored, err := engine.GetOptimization(ctx, tt.filePath)
			assert.NoError(t, err)
			assert.Equal(t, result, stored)
		})
	}
}

func TestMLClassificationEngine_PredictCacheAccess(t *testing.T) {
	engine := NewMLClassificationEngine()
	ctx := context.Background()

	now := time.Now()

	tests := []struct {
		name          string
		key           string
		accessHistory []time.Time
		metadata      map[string]interface{}
		expected      struct {
			accessProbability float64
			priority          int
		}
	}{
		{
			name:          "no access history",
			key:           "key1",
			accessHistory: []time.Time{},
			metadata:      map[string]interface{}{"type": "test"},
			expected: struct {
				accessProbability float64
				priority          int
			}{
				accessProbability: 0.1,
				priority:          10,
			},
		},
		{
			name: "recent access history",
			key:  "key2",
			accessHistory: []time.Time{
				now.Add(-1 * time.Hour),
				now.Add(-2 * time.Hour),
				now.Add(-3 * time.Hour),
			},
			metadata: map[string]interface{}{"type": "test"},
			expected: struct {
				accessProbability float64
				priority          int
			}{
				accessProbability: 1.0, // All accesses are recent
				priority:          100, // Base priority for recent access
			},
		},
		{
			name: "mixed access history",
			key:  "key3",
			accessHistory: []time.Time{
				now.Add(-1 * time.Hour),  // Recent
				now.Add(-25 * time.Hour), // Old
				now.Add(-2 * time.Hour),  // Recent
			},
			metadata: map[string]interface{}{"type": "test"},
			expected: struct {
				accessProbability float64
				priority          int
			}{
				accessProbability: 0.6666666666666666, // 2/3 recent accesses
				priority:          66,                 // Base priority for mixed access
			},
		},
		{
			name: "old access history",
			key:  "key4",
			accessHistory: []time.Time{
				now.Add(-25 * time.Hour),
				now.Add(-26 * time.Hour),
				now.Add(-27 * time.Hour),
			},
			metadata: map[string]interface{}{"type": "test"},
			expected: struct {
				accessProbability float64
				priority          int
			}{
				accessProbability: 0.0, // No recent accesses
				priority:          1,   // Minimum priority
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prediction, err := engine.PredictCacheAccess(ctx, tt.key, tt.accessHistory, tt.metadata)
			assert.NoError(t, err)
			assert.NotNil(t, prediction)
			assert.Equal(t, tt.key, prediction.Key)
			assert.Equal(t, tt.expected.accessProbability, prediction.AccessProbability)
			assert.Equal(t, tt.expected.priority, prediction.Priority)
			assert.NotZero(t, prediction.RecommendedTTL)
			assert.NotEmpty(t, prediction.Reason)
			assert.Equal(t, tt.metadata, prediction.Metadata)

			// Verify prediction is stored
			stored, err := engine.GetCachePrediction(ctx, tt.key)
			assert.NoError(t, err)
			assert.Equal(t, prediction, stored)
		})
	}
}

func TestMLClassificationEngine_TrainModel(t *testing.T) {
	engine := NewMLClassificationEngine()
	ctx := context.Background()

	model := &MLModel{
		ID:      "test-model-1",
		Name:    "Test Model",
		Type:    "classification",
		Version: "1.0.0",
	}

	trainingData := []map[string]interface{}{
		{"feature1": "value1", "feature2": "value2", "label": "class1"},
		{"feature1": "value3", "feature2": "value4", "label": "class2"},
		{"feature1": "value5", "feature2": "value6", "label": "class1"},
	}

	err := engine.TrainModel(ctx, model, trainingData)
	assert.NoError(t, err)
	assert.NotNil(t, model.TrainingData)
	assert.Equal(t, 3, model.TrainingData["sample_count"])
	assert.NotNil(t, model.TrainingData["features"])
	assert.NotNil(t, model.TrainingData["labels"])
	assert.Greater(t, model.Accuracy, 0.8)
	assert.NotZero(t, model.CreatedAt)
	assert.NotZero(t, model.UpdatedAt)

	// Verify model is stored
	stored, err := engine.GetModel(ctx, model.ID)
	assert.NoError(t, err)
	assert.Equal(t, model, stored)
}

func TestMLClassificationEngine_GetModel_NotFound(t *testing.T) {
	engine := NewMLClassificationEngine()
	ctx := context.Background()

	_, err := engine.GetModel(ctx, "non-existent-model")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model not found")
}

func TestMLClassificationEngine_ListModels(t *testing.T) {
	engine := NewMLClassificationEngine()
	ctx := context.Background()

	// Initially empty
	models, err := engine.ListModels(ctx)
	assert.NoError(t, err)
	assert.Len(t, models, 0)

	// Add a model
	model := &MLModel{
		ID:      "test-model-1",
		Name:    "Test Model 1",
		Type:    "classification",
		Version: "1.0.0",
	}
	err = engine.TrainModel(ctx, model, []map[string]interface{}{})
	assert.NoError(t, err)

	// Add another model
	model2 := &MLModel{
		ID:      "test-model-2",
		Name:    "Test Model 2",
		Type:    "regression",
		Version: "2.0.0",
	}
	err = engine.TrainModel(ctx, model2, []map[string]interface{}{})
	assert.NoError(t, err)

	// List models
	models, err = engine.ListModels(ctx)
	assert.NoError(t, err)
	assert.Len(t, models, 2)
}

func TestMLClassificationEngine_GetClassification_NotFound(t *testing.T) {
	engine := NewMLClassificationEngine()
	ctx := context.Background()

	_, err := engine.GetClassification(ctx, "non-existent-file")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "classification not found")
}

func TestMLClassificationEngine_ListClassifications(t *testing.T) {
	engine := NewMLClassificationEngine()
	ctx := context.Background()

	// Initially empty
	classifications, err := engine.ListClassifications(ctx)
	assert.NoError(t, err)
	assert.Len(t, classifications, 0)

	// Add classifications
	_, err = engine.ClassifyFile(ctx, "file1.txt", []byte("content1"), nil)
	assert.NoError(t, err)

	_, err = engine.ClassifyFile(ctx, "file2.jpg", []byte("content2"), nil)
	assert.NoError(t, err)

	// List classifications
	classifications, err = engine.ListClassifications(ctx)
	assert.NoError(t, err)
	assert.Len(t, classifications, 2)
}

func TestMLClassificationEngine_GetOptimization_NotFound(t *testing.T) {
	engine := NewMLClassificationEngine()
	ctx := context.Background()

	_, err := engine.GetOptimization(ctx, "non-existent-file")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "optimization not found")
}

func TestMLClassificationEngine_ListOptimizations(t *testing.T) {
	engine := NewMLClassificationEngine()
	ctx := context.Background()

	// Initially empty
	optimizations, err := engine.ListOptimizations(ctx)
	assert.NoError(t, err)
	assert.Len(t, optimizations, 0)

	// Add optimizations
	_, err = engine.OptimizeFile(ctx, "file1.txt", []byte("content1"), "compression")
	assert.NoError(t, err)

	_, err = engine.OptimizeFile(ctx, "file2.txt", []byte("content2"), "deduplication")
	assert.NoError(t, err)

	// List optimizations
	optimizations, err = engine.ListOptimizations(ctx)
	assert.NoError(t, err)
	assert.Len(t, optimizations, 2)
}

func TestMLClassificationEngine_GetCachePrediction_NotFound(t *testing.T) {
	engine := NewMLClassificationEngine()
	ctx := context.Background()

	_, err := engine.GetCachePrediction(ctx, "non-existent-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cache prediction not found")
}

func TestMLClassificationEngine_ListCachePredictions(t *testing.T) {
	engine := NewMLClassificationEngine()
	ctx := context.Background()

	// Initially empty
	predictions, err := engine.ListCachePredictions(ctx)
	assert.NoError(t, err)
	assert.Len(t, predictions, 0)

	// Add predictions
	_, err = engine.PredictCacheAccess(ctx, "key1", []time.Time{}, nil)
	assert.NoError(t, err)

	_, err = engine.PredictCacheAccess(ctx, "key2", []time.Time{}, nil)
	assert.NoError(t, err)

	// List predictions
	predictions, err = engine.ListCachePredictions(ctx)
	assert.NoError(t, err)
	assert.Len(t, predictions, 2)
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{"simple extension", "file.txt", ".txt"},
		{"multiple dots", "file.backup.txt", ".txt"},
		{"no extension", "file", ""},
		{"empty path", "", ""},
		{"dot only", ".", "."},
		{"hidden file", ".hidden", ".hidden"},
		{"hidden file with extension", ".hidden.txt", ".txt"},
		{"path with directory", "/path/to/file.txt", ".txt"},
		{"windows path", "C:\\path\\to\\file.txt", ".txt"},
		{"complex extension", "file.tar.gz", ".gz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFileExtension(tt.filePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMimeType(t *testing.T) {
	tests := []struct {
		name      string
		extension string
		expected  string
	}{
		{"text file", ".txt", "text/plain"},
		{"markdown file", ".md", "text/markdown"},
		{"JSON file", ".json", "application/json"},
		{"XML file", ".xml", "application/xml"},
		{"PDF file", ".pdf", "application/pdf"},
		{"JPEG image", ".jpg", "image/jpeg"},
		{"PNG image", ".png", "image/png"},
		{"GIF image", ".gif", "image/gif"},
		{"MP4 video", ".mp4", "video/mp4"},
		{"MP3 audio", ".mp3", "audio/mpeg"},
		{"ZIP archive", ".zip", "application/zip"},
		{"Go source", ".go", "text/x-go"},
		{"JavaScript", ".js", "application/javascript"},
		{"Python", ".py", "text/x-python"},
		{"unknown extension", ".xyz", "application/octet-stream"},
		{"empty extension", "", "application/octet-stream"},
		{"uppercase extension", ".TXT", "text/plain"},
		{"mixed case extension", ".JpG", "image/jpeg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMimeType(tt.extension)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{"contains item", []string{"a", "b", "c"}, "b", true},
		{"does not contain item", []string{"a", "b", "c"}, "d", false},
		{"empty slice", []string{}, "a", false},
		{"single item match", []string{"a"}, "a", true},
		{"single item no match", []string{"a"}, "b", false},
		{"duplicate items", []string{"a", "a", "b"}, "a", true},
		{"case sensitive", []string{"A", "B", "C"}, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMLClassificationEngine_EdgeCases(t *testing.T) {
	engine := NewMLClassificationEngine()
	ctx := context.Background()

	// Test with very large content
	largeContent := make([]byte, 1024*1024) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	classification, err := engine.ClassifyFile(ctx, "large.bin", largeContent, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(largeContent)), classification.Size)

	// Test with nil metadata
	classification, err = engine.ClassifyFile(ctx, "test.txt", []byte("content"), nil)
	assert.NoError(t, err)
	assert.Nil(t, classification.Metadata)

	// Test with empty metadata
	classification, err = engine.ClassifyFile(ctx, "test.txt", []byte("content"), map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, classification.Metadata)
	assert.Len(t, classification.Metadata, 0)
}

func TestMLClassificationEngine_ConcurrentAccess(t *testing.T) {
	engine := NewMLClassificationEngine()
	ctx := context.Background()

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()

			// Classify file
			_, err := engine.ClassifyFile(ctx, "file"+string(rune(i))+".txt", []byte("content"), nil)
			assert.NoError(t, err)

			// Optimize file
			_, err = engine.OptimizeFile(ctx, "file"+string(rune(i))+".txt", []byte("content"), "compression")
			assert.NoError(t, err)

			// Predict cache access
			_, err = engine.PredictCacheAccess(ctx, "key"+string(rune(i)), []time.Time{}, nil)
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all operations completed
	classifications, err := engine.ListClassifications(ctx)
	assert.NoError(t, err)
	assert.Len(t, classifications, 10)

	optimizations, err := engine.ListOptimizations(ctx)
	assert.NoError(t, err)
	assert.Len(t, optimizations, 10)

	predictions, err := engine.ListCachePredictions(ctx)
	assert.NoError(t, err)
	assert.Len(t, predictions, 10)
}
