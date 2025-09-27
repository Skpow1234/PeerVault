package coap

import (
	"log/slog"
	"net"
	"sync"
	"time"
)

// Resource represents a CoAP resource
type Resource struct {
	// Resource metadata
	Name          string
	Description   string
	Content       []byte
	ContentFormat *CoAPContentFormat

	// Request handlers
	GetHandler    RequestHandler
	PostHandler   RequestHandler
	PutHandler    RequestHandler
	DeleteHandler RequestHandler

	// Resource state
	Observable   bool
	MaxAge       int
	ETag         []byte
	LastModified time.Time

	// Mutex for thread-safe access
	mu sync.RWMutex
}

// RequestHandler is a function that handles CoAP requests
type RequestHandler func(*Message, *Client) (*Message, error)

// Observer represents a CoAP observer
type Observer struct {
	Client    *Client
	Token     []byte
	Resource  *Resource
	CreatedAt time.Time
}

// Client represents a CoAP client
type Client struct {
	ID        string
	Address   *net.UDPAddr
	CreatedAt time.Time
	LastSeen  time.Time
	MessageID uint16
	mu        sync.Mutex
}

// getNextMessageID gets the next message ID for the client
func (c *Client) getNextMessageID() uint16 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.MessageID++
	return c.MessageID
}

// Close closes the client connection
func (c *Client) Close() {
	// Client cleanup if needed
}

// GetContent gets the current content of the resource
func (r *Resource) GetContent() []byte {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Content
}

// SetContent sets the content of the resource
func (r *Resource) SetContent(content []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Content = content
	r.LastModified = time.Now()
	// Update ETag
	r.ETag = generateETag(content)
}

// IsObservable returns true if the resource is observable
func (r *Resource) IsObservable() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Observable
}

// SetObservable sets whether the resource is observable
func (r *Resource) SetObservable(observable bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Observable = observable
}

// GetMaxAge returns the max age of the resource
func (r *Resource) GetMaxAge() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.MaxAge
}

// SetMaxAge sets the max age of the resource
func (r *Resource) SetMaxAge(maxAge int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.MaxAge = maxAge
}

// GetETag returns the ETag of the resource
func (r *Resource) GetETag() []byte {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ETag
}

// GetLastModified returns the last modified time of the resource
func (r *Resource) GetLastModified() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.LastModified
}

// generateETag generates an ETag for the content
func generateETag(content []byte) []byte {
	// Simple ETag generation based on content hash
	// In a real implementation, you might use a proper hash function
	hash := uint32(0)
	for _, b := range content {
		hash = hash*31 + uint32(b)
	}

	// Convert to bytes
	etag := make([]byte, 4)
	etag[0] = byte(hash >> 24)
	etag[1] = byte(hash >> 16)
	etag[2] = byte(hash >> 8)
	etag[3] = byte(hash)

	return etag
}

// NewResource creates a new CoAP resource
func NewResource(name, description string, content []byte) *Resource {
	return &Resource{
		Name:         name,
		Description:  description,
		Content:      content,
		MaxAge:       60, // Default max age
		Observable:   false,
		LastModified: time.Now(),
		ETag:         generateETag(content),
	}
}

// NewObservableResource creates a new observable CoAP resource
func NewObservableResource(name, description string, content []byte) *Resource {
	resource := NewResource(name, description, content)
	resource.Observable = true
	return resource
}

// SetGetHandler sets the GET request handler
func (r *Resource) SetGetHandler(handler RequestHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.GetHandler = handler
}

// SetPostHandler sets the POST request handler
func (r *Resource) SetPostHandler(handler RequestHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.PostHandler = handler
}

// SetPutHandler sets the PUT request handler
func (r *Resource) SetPutHandler(handler RequestHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.PutHandler = handler
}

// SetDeleteHandler sets the DELETE request handler
func (r *Resource) SetDeleteHandler(handler RequestHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.DeleteHandler = handler
}

// SetContentFormat sets the content format of the resource
func (r *Resource) SetContentFormat(format CoAPContentFormat) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ContentFormat = &format
}

// UpdateContent updates the content and notifies observers
func (r *Resource) UpdateContent(content []byte, notifyObservers func([]byte)) {
	r.SetContent(content)
	if r.IsObservable() && notifyObservers != nil {
		notifyObservers(content)
	}
}

// ResourceManager manages CoAP resources
type ResourceManager struct {
	resources map[string]*Resource
	mu        sync.RWMutex
	logger    *slog.Logger
}

// NewResourceManager creates a new resource manager
func NewResourceManager(logger *slog.Logger) *ResourceManager {
	return &ResourceManager{
		resources: make(map[string]*Resource),
		logger:    logger,
	}
}

// RegisterResource registers a resource
func (rm *ResourceManager) RegisterResource(path string, resource *Resource) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.resources[path] = resource
	rm.logger.Debug("Resource registered", "path", path, "name", resource.Name)
}

// GetResource gets a resource by path
func (rm *ResourceManager) GetResource(path string) (*Resource, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	resource, exists := rm.resources[path]
	return resource, exists
}

// UnregisterResource unregisters a resource
func (rm *ResourceManager) UnregisterResource(path string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.resources, path)
	rm.logger.Debug("Resource unregistered", "path", path)
}

// ListResources returns all registered resources
func (rm *ResourceManager) ListResources() map[string]*Resource {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Return a copy to avoid race conditions
	resources := make(map[string]*Resource)
	for path, resource := range rm.resources {
		resources[path] = resource
	}
	return resources
}

// GetResourceCount returns the number of registered resources
func (rm *ResourceManager) GetResourceCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return len(rm.resources)
}

// ObserverManager manages CoAP observers
type ObserverManager struct {
	observers map[string][]*Observer
	mu        sync.RWMutex
	logger    *slog.Logger
}

// NewObserverManager creates a new observer manager
func NewObserverManager(logger *slog.Logger) *ObserverManager {
	return &ObserverManager{
		observers: make(map[string][]*Observer),
		logger:    logger,
	}
}

// AddObserver adds an observer to a resource
func (om *ObserverManager) AddObserver(path string, observer *Observer) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.observers[path] = append(om.observers[path], observer)
	om.logger.Debug("Observer added", "path", path, "client", observer.Client.ID)
}

// RemoveObserver removes an observer from a resource
func (om *ObserverManager) RemoveObserver(path string, clientID string) {
	om.mu.Lock()
	defer om.mu.Unlock()

	observers := om.observers[path]
	for i, observer := range observers {
		if observer.Client.ID == clientID {
			om.observers[path] = append(observers[:i], observers[i+1:]...)
			om.logger.Debug("Observer removed", "path", path, "client", clientID)
			break
		}
	}
}

// GetObservers returns all observers for a resource
func (om *ObserverManager) GetObservers(path string) []*Observer {
	om.mu.RLock()
	defer om.mu.RUnlock()

	observers := om.observers[path]
	// Return a copy to avoid race conditions
	result := make([]*Observer, len(observers))
	copy(result, observers)
	return result
}

// GetObserverCount returns the number of observers for a resource
func (om *ObserverManager) GetObserverCount(path string) int {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return len(om.observers[path])
}

// GetTotalObserverCount returns the total number of observers
func (om *ObserverManager) GetTotalObserverCount() int {
	om.mu.RLock()
	defer om.mu.RUnlock()

	total := 0
	for _, observers := range om.observers {
		total += len(observers)
	}
	return total
}

// CleanupObservers removes expired observers
func (om *ObserverManager) CleanupObservers(timeout time.Duration) {
	om.mu.Lock()
	defer om.mu.Unlock()

	for path, observers := range om.observers {
		var activeObservers []*Observer
		for _, observer := range observers {
			if time.Since(observer.CreatedAt) < timeout {
				activeObservers = append(activeObservers, observer)
			}
		}
		om.observers[path] = activeObservers
	}
}
