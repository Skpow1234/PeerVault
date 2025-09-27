package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewObjectPool(t *testing.T) {
	// Test creating a new object pool
	pool := NewObjectPool(func() []byte {
		return make([]byte, 1024)
	})

	assert.NotNil(t, pool)
	assert.NotNil(t, pool.pool)
}

func TestObjectPoolGetPut(t *testing.T) {
	// Test getting and putting objects in the pool
	pool := NewObjectPool(func() []byte {
		return make([]byte, 1024)
	})

	// Get an object from the pool
	obj1 := pool.Get()
	assert.NotNil(t, obj1)
	assert.Len(t, obj1, 1024)

	// Put the object back
	pool.Put(obj1)

	// Get another object (should be the same one we put back)
	obj2 := pool.Get()
	assert.NotNil(t, obj2)
	assert.Len(t, obj2, 1024)
}

func TestObjectPoolConcurrentAccess(t *testing.T) {
	// Test concurrent access to the pool
	pool := NewObjectPool(func() []byte {
		return make([]byte, 512)
	})

	// Test concurrent get/put operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Get object
			obj := pool.Get()
			assert.NotNil(t, obj)
			assert.Len(t, obj, 512)

			// Put object back
			pool.Put(obj)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestNewResettableObjectPool(t *testing.T) {
	// Test creating a new resettable object pool
	resetFunc := func(buf []byte) {
		for i := range buf {
			buf[i] = 0
		}
	}

	pool := NewResettableObjectPool(func() []byte {
		return make([]byte, 1024)
	}, resetFunc)

	assert.NotNil(t, pool)
	assert.NotNil(t, pool.pool)
	assert.NotNil(t, pool.reset)
}

func TestResettableObjectPoolGetPut(t *testing.T) {
	// Test getting and putting objects in the resettable pool
	resetFunc := func(buf []byte) {
		for i := range buf {
			buf[i] = 0
		}
	}

	pool := NewResettableObjectPool(func() []byte {
		return make([]byte, 1024)
	}, resetFunc)

	// Get an object from the pool
	obj1 := pool.Get()
	assert.NotNil(t, obj1)
	assert.Len(t, obj1, 1024)

	// Fill the object with data
	for i := range obj1 {
		obj1[i] = byte(i % 256)
	}

	// Put the object back (should be reset)
	pool.Put(obj1)

	// Get another object (should be reset)
	obj2 := pool.Get()
	assert.NotNil(t, obj2)
	assert.Len(t, obj2, 1024)

	// Check that the object was reset
	for i := range obj2 {
		assert.Equal(t, byte(0), obj2[i])
	}
}

func TestResettableObjectPoolConcurrentAccess(t *testing.T) {
	// Test concurrent access to the resettable pool
	resetFunc := func(buf []byte) {
		for i := range buf {
			buf[i] = 0
		}
	}

	pool := NewResettableObjectPool(func() []byte {
		return make([]byte, 512)
	}, resetFunc)

	// Test concurrent get/put operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Get object
			obj := pool.Get()
			assert.NotNil(t, obj)
			assert.Len(t, obj, 512)

			// Fill with data
			for j := range obj {
				obj[j] = byte(j % 256)
			}

			// Put object back (should be reset)
			pool.Put(obj)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestObjectPoolWithStruct(t *testing.T) {
	// Test object pool with a struct type
	type TestStruct struct {
		ID   int
		Name string
		Data []byte
	}

	pool := NewObjectPool(func() TestStruct {
		return TestStruct{
			ID:   0,
			Name: "",
			Data: make([]byte, 100),
		}
	})

	// Get an object from the pool
	obj1 := pool.Get()
	assert.Equal(t, 0, obj1.ID)
	assert.Equal(t, "", obj1.Name)
	assert.Len(t, obj1.Data, 100)

	// Modify the object
	obj1.ID = 123
	obj1.Name = "test"
	obj1.Data[0] = 42

	// Put the object back
	pool.Put(obj1)

	// Get another object
	obj2 := pool.Get()
	assert.Equal(t, 123, obj2.ID)
	assert.Equal(t, "test", obj2.Name)
	assert.Equal(t, byte(42), obj2.Data[0])
}

func TestResettableObjectPoolWithStruct(t *testing.T) {
	// Test resettable object pool with a struct type
	type TestStruct struct {
		ID   int
		Name string
		Data []byte
	}

	// For structs, we need to use pointers to make reset work properly
	resetFunc := func(obj *TestStruct) {
		obj.ID = 0
		obj.Name = ""
		for i := range obj.Data {
			obj.Data[i] = 0
		}
	}

	pool := NewResettableObjectPool(func() *TestStruct {
		return &TestStruct{
			ID:   0,
			Name: "",
			Data: make([]byte, 100),
		}
	}, resetFunc)

	// Get an object from the pool
	obj1 := pool.Get()
	assert.Equal(t, 0, obj1.ID)
	assert.Equal(t, "", obj1.Name)
	assert.Len(t, obj1.Data, 100)

	// Modify the object
	obj1.ID = 123
	obj1.Name = "test"
	obj1.Data[0] = 42

	// Put the object back (should be reset)
	pool.Put(obj1)

	// Get another object
	obj2 := pool.Get()
	assert.Equal(t, 0, obj2.ID)
	assert.Equal(t, "", obj2.Name)
	assert.Equal(t, byte(0), obj2.Data[0])
}
