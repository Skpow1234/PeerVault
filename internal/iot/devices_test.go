package iot

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewIoTManager(t *testing.T) {
	manager := NewIoTManager()
	
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.devices)
	assert.NotNil(t, manager.sensorData)
	assert.NotNil(t, manager.commands)
	assert.NotNil(t, manager.metrics)
	assert.Equal(t, 0, len(manager.devices))
	assert.Equal(t, 0, len(manager.sensorData))
	assert.Equal(t, 0, len(manager.commands))
}

func TestIoTManager_RegisterDevice(t *testing.T) {
	manager := NewIoTManager()
	ctx := context.Background()
	
	device := &IoTDevice{
		ID:           "device1",
		Name:         "Test Device",
		Type:         "sensor",
		Manufacturer: "TestCorp",
		Model:        "TC-100",
		Version:      "1.0",
		Status:       "inactive",
		Location: &DeviceLocation{
			Latitude:  40.7128,
			Longitude: -74.0060,
			Room:      "Living Room",
			Building:  "Home",
			Floor:     1,
		},
		Capabilities: &DeviceCapabilities{
			Sensors: []*Sensor{
				{
					ID:     "sensor1",
					Name:   "Temperature",
					Type:   "temperature",
					Unit:   "°C",
					Status: "active",
				},
			},
			Actuators: []*Actuator{
				{
					ID:     "actuator1",
					Name:   "LED",
					Type:   "light",
					Status: "active",
				},
			},
		},
		Protocols: []string{"MQTT", "HTTP"},
	}
	
	err := manager.RegisterDevice(ctx, device)
	assert.NoError(t, err)
	
	// Verify device was registered
	registeredDevice, err := manager.GetDevice(ctx, "device1")
	assert.NoError(t, err)
	assert.Equal(t, "device1", registeredDevice.ID)
	assert.Equal(t, "active", registeredDevice.Status)
	assert.NotZero(t, registeredDevice.LastSeen)
	
	// Verify metrics were updated
	metrics, err := manager.GetMetrics(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, metrics.TotalDevices)
	assert.Equal(t, 1, metrics.ActiveDevices)
	assert.Equal(t, 1, metrics.TotalSensors)
	assert.Equal(t, 1, metrics.ActiveSensors)
	assert.Equal(t, 1, metrics.TotalActuators)
	assert.Equal(t, 1, metrics.ActiveActuators)
}

func TestIoTManager_UnregisterDevice(t *testing.T) {
	manager := NewIoTManager()
	ctx := context.Background()
	
	device := &IoTDevice{
		ID:     "device1",
		Name:   "Test Device",
		Type:   "sensor",
		Status: "active",
		Capabilities: &DeviceCapabilities{
			Sensors: []*Sensor{
				{ID: "sensor1", Status: "active"},
			},
		},
	}
	
	// Register device first
	err := manager.RegisterDevice(ctx, device)
	assert.NoError(t, err)
	
	// Unregister device
	err = manager.UnregisterDevice(ctx, "device1")
	assert.NoError(t, err)
	
	// Verify device status was updated
	registeredDevice, err := manager.GetDevice(ctx, "device1")
	assert.NoError(t, err)
	assert.Equal(t, "inactive", registeredDevice.Status)
	
	// Verify metrics were updated
	metrics, err := manager.GetMetrics(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, metrics.TotalDevices)
	assert.Equal(t, 0, metrics.ActiveDevices)
}

func TestIoTManager_GetDevice(t *testing.T) {
	manager := NewIoTManager()
	ctx := context.Background()
	
	device := &IoTDevice{
		ID:   "device1",
		Name: "Test Device",
		Type: "sensor",
	}
	
	// Register device
	err := manager.RegisterDevice(ctx, device)
	assert.NoError(t, err)
	
	// Test getting existing device
	retrievedDevice, err := manager.GetDevice(ctx, "device1")
	assert.NoError(t, err)
	assert.Equal(t, "device1", retrievedDevice.ID)
	assert.Equal(t, "Test Device", retrievedDevice.Name)
	assert.Equal(t, "sensor", retrievedDevice.Type)
	
	// Test getting non-existent device
	_, err = manager.GetDevice(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device not found")
}

func TestIoTManager_ListDevices(t *testing.T) {
	manager := NewIoTManager()
	ctx := context.Background()
	
	// Test empty list
	devices, err := manager.ListDevices(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(devices))
	
	// Register multiple devices
	device1 := &IoTDevice{ID: "device1", Name: "Device 1", Type: "sensor"}
	device2 := &IoTDevice{ID: "device2", Name: "Device 2", Type: "actuator"}
	
	err = manager.RegisterDevice(ctx, device1)
	assert.NoError(t, err)
	err = manager.RegisterDevice(ctx, device2)
	assert.NoError(t, err)
	
	// Test listing all devices
	devices, err = manager.ListDevices(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(devices))
	
	// Verify both devices are in the list
	deviceIDs := make(map[string]bool)
	for _, device := range devices {
		deviceIDs[device.ID] = true
	}
	assert.True(t, deviceIDs["device1"])
	assert.True(t, deviceIDs["device2"])
}

func TestIoTManager_ListDevicesByType(t *testing.T) {
	manager := NewIoTManager()
	ctx := context.Background()
	
	// Register devices of different types
	device1 := &IoTDevice{ID: "device1", Name: "Sensor 1", Type: "sensor"}
	device2 := &IoTDevice{ID: "device2", Name: "Actuator 1", Type: "actuator"}
	device3 := &IoTDevice{ID: "device3", Name: "Sensor 2", Type: "sensor"}
	
	err := manager.RegisterDevice(ctx, device1)
	assert.NoError(t, err)
	err = manager.RegisterDevice(ctx, device2)
	assert.NoError(t, err)
	err = manager.RegisterDevice(ctx, device3)
	assert.NoError(t, err)
	
	// Test listing sensors
	sensors, err := manager.ListDevicesByType(ctx, "sensor")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(sensors))
	
	// Test listing actuators
	actuators, err := manager.ListDevicesByType(ctx, "actuator")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actuators))
	
	// Test listing non-existent type
	nonexistent, err := manager.ListDevicesByType(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(nonexistent))
}

func TestIoTManager_ListDevicesByLocation(t *testing.T) {
	manager := NewIoTManager()
	ctx := context.Background()
	
	location1 := &DeviceLocation{
		Building: "Building A",
		Room:     "Room 1",
		Floor:    1,
	}
	
	location2 := &DeviceLocation{
		Building: "Building A",
		Room:     "Room 2",
		Floor:    1,
	}
	
	location3 := &DeviceLocation{
		Building: "Building B",
		Room:     "Room 1",
		Floor:    1,
	}
	
	// Register devices with different locations
	device1 := &IoTDevice{ID: "device1", Name: "Device 1", Location: location1}
	device2 := &IoTDevice{ID: "device2", Name: "Device 2", Location: location2}
	device3 := &IoTDevice{ID: "device3", Name: "Device 3", Location: location3}
	device4 := &IoTDevice{ID: "device4", Name: "Device 4", Location: nil}
	
	err := manager.RegisterDevice(ctx, device1)
	assert.NoError(t, err)
	err = manager.RegisterDevice(ctx, device2)
	assert.NoError(t, err)
	err = manager.RegisterDevice(ctx, device3)
	assert.NoError(t, err)
	err = manager.RegisterDevice(ctx, device4)
	assert.NoError(t, err)
	
	// Test listing devices in Building A, Room 1
	searchLocation := &DeviceLocation{
		Building: "Building A",
		Room:     "Room 1",
		Floor:    1,
	}
	
	devices, err := manager.ListDevicesByLocation(ctx, searchLocation)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(devices))
	assert.Equal(t, "device1", devices[0].ID)
	
	// Test listing devices in Building A, Room 1 (any floor)
	searchLocation2 := &DeviceLocation{
		Building: "Building A",
		Room:     "Room 1",
		Floor:    0, // Don't specify floor
	}
	
	devices, err = manager.ListDevicesByLocation(ctx, searchLocation2)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(devices)) // Only device1 (Building A, Room 1)
	assert.Equal(t, "device1", devices[0].ID)
}

func TestIoTManager_SendSensorData(t *testing.T) {
	manager := NewIoTManager()
	ctx := context.Background()
	
	// Register device with sensor
	device := &IoTDevice{
		ID:   "device1",
		Name: "Test Device",
		Capabilities: &DeviceCapabilities{
			Sensors: []*Sensor{
				{ID: "sensor1", Name: "Temperature", Type: "temperature"},
			},
		},
	}
	
	err := manager.RegisterDevice(ctx, device)
	assert.NoError(t, err)
	
	// Send sensor data
	sensorData := &SensorData{
		DeviceID:  "device1",
		SensorID:  "sensor1",
		Value:     25.5,
		Unit:      "°C",
		Timestamp: time.Now(),
		Quality:   "good",
	}
	
	err = manager.SendSensorData(ctx, sensorData)
	assert.NoError(t, err)
	
	// Verify data was stored
	data, err := manager.GetSensorData(ctx, "device1", "sensor1", 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(data))
	assert.Equal(t, 25.5, data[0].Value)
	assert.Equal(t, "°C", data[0].Unit)
	
	// Test sending data for non-existent device
	invalidData := &SensorData{
		DeviceID: "nonexistent",
		SensorID: "sensor1",
		Value:    25.5,
	}
	
	err = manager.SendSensorData(ctx, invalidData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device not found")
	
	// Test sending data for non-existent sensor
	invalidSensorData := &SensorData{
		DeviceID: "device1",
		SensorID: "nonexistent",
		Value:    25.5,
	}
	
	err = manager.SendSensorData(ctx, invalidSensorData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sensor not found")
}

func TestIoTManager_GetSensorData(t *testing.T) {
	manager := NewIoTManager()
	ctx := context.Background()
	
	// Register device with multiple sensors
	device := &IoTDevice{
		ID:   "device1",
		Name: "Test Device",
		Capabilities: &DeviceCapabilities{
			Sensors: []*Sensor{
				{ID: "sensor1", Name: "Temperature"},
				{ID: "sensor2", Name: "Humidity"},
			},
		},
	}
	
	err := manager.RegisterDevice(ctx, device)
	assert.NoError(t, err)
	
	// Send multiple sensor data points
	data1 := &SensorData{DeviceID: "device1", SensorID: "sensor1", Value: 25.0, Timestamp: time.Now()}
	data2 := &SensorData{DeviceID: "device1", SensorID: "sensor1", Value: 26.0, Timestamp: time.Now()}
	data3 := &SensorData{DeviceID: "device1", SensorID: "sensor2", Value: 60.0, Timestamp: time.Now()}
	
	err = manager.SendSensorData(ctx, data1)
	assert.NoError(t, err)
	err = manager.SendSensorData(ctx, data2)
	assert.NoError(t, err)
	err = manager.SendSensorData(ctx, data3)
	assert.NoError(t, err)
	
	// Test getting all data for device
	allData, err := manager.GetSensorData(ctx, "device1", "", 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(allData))
	
	// Test getting data for specific sensor
	sensor1Data, err := manager.GetSensorData(ctx, "device1", "sensor1", 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(sensor1Data))
	
	// Test limiting results
	limitedData, err := manager.GetSensorData(ctx, "device1", "", 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(limitedData))
	
	// Test getting data for non-existent device
	_, err = manager.GetSensorData(ctx, "nonexistent", "", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no data found for device")
}

func TestIoTManager_SendActuatorCommand(t *testing.T) {
	manager := NewIoTManager()
	ctx := context.Background()
	
	// Register device with actuator
	device := &IoTDevice{
		ID:   "device1",
		Name: "Test Device",
		Capabilities: &DeviceCapabilities{
			Actuators: []*Actuator{
				{ID: "actuator1", Name: "LED", Type: "light"},
			},
		},
	}
	
	err := manager.RegisterDevice(ctx, device)
	assert.NoError(t, err)
	
	// Send actuator command
	command := &ActuatorCommand{
		DeviceID:   "device1",
		ActuatorID: "actuator1",
		Command:    "turn_on",
		Value:      100.0,
		Unit:       "%",
		Timestamp:  time.Now(),
		Priority:   1,
	}
	
	err = manager.SendActuatorCommand(ctx, command)
	assert.NoError(t, err)
	
	// Verify command was stored
	commands, err := manager.GetActuatorCommands(ctx, "device1", "actuator1", 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(commands))
	assert.Equal(t, "turn_on", commands[0].Command)
	assert.Equal(t, 100.0, commands[0].Value)
	
	// Test sending command for non-existent device
	invalidCommand := &ActuatorCommand{
		DeviceID:   "nonexistent",
		ActuatorID: "actuator1",
		Command:    "turn_on",
	}
	
	err = manager.SendActuatorCommand(ctx, invalidCommand)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device not found")
	
	// Test sending command for non-existent actuator
	invalidActuatorCommand := &ActuatorCommand{
		DeviceID:   "device1",
		ActuatorID: "nonexistent",
		Command:    "turn_on",
	}
	
	err = manager.SendActuatorCommand(ctx, invalidActuatorCommand)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "actuator not found")
}

func TestIoTManager_GetActuatorCommands(t *testing.T) {
	manager := NewIoTManager()
	ctx := context.Background()
	
	// Register device with multiple actuators
	device := &IoTDevice{
		ID:   "device1",
		Name: "Test Device",
		Capabilities: &DeviceCapabilities{
			Actuators: []*Actuator{
				{ID: "actuator1", Name: "LED"},
				{ID: "actuator2", Name: "Motor"},
			},
		},
	}
	
	err := manager.RegisterDevice(ctx, device)
	assert.NoError(t, err)
	
	// Send multiple commands
	cmd1 := &ActuatorCommand{DeviceID: "device1", ActuatorID: "actuator1", Command: "turn_on", Timestamp: time.Now()}
	cmd2 := &ActuatorCommand{DeviceID: "device1", ActuatorID: "actuator1", Command: "turn_off", Timestamp: time.Now()}
	cmd3 := &ActuatorCommand{DeviceID: "device1", ActuatorID: "actuator2", Command: "start", Timestamp: time.Now()}
	
	err = manager.SendActuatorCommand(ctx, cmd1)
	assert.NoError(t, err)
	err = manager.SendActuatorCommand(ctx, cmd2)
	assert.NoError(t, err)
	err = manager.SendActuatorCommand(ctx, cmd3)
	assert.NoError(t, err)
	
	// Test getting all commands for device
	allCommands, err := manager.GetActuatorCommands(ctx, "device1", "", 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(allCommands))
	
	// Test getting commands for specific actuator
	actuator1Commands, err := manager.GetActuatorCommands(ctx, "device1", "actuator1", 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(actuator1Commands))
	
	// Test limiting results
	limitedCommands, err := manager.GetActuatorCommands(ctx, "device1", "", 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(limitedCommands))
	
	// Test getting commands for non-existent device
	_, err = manager.GetActuatorCommands(ctx, "nonexistent", "", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no commands found for device")
}

func TestIoTManager_UpdateDeviceStatus(t *testing.T) {
	manager := NewIoTManager()
	ctx := context.Background()
	
	device := &IoTDevice{
		ID:     "device1",
		Name:   "Test Device",
		Status: "active",
		Capabilities: &DeviceCapabilities{
			Sensors: []*Sensor{
				{ID: "sensor1", Status: "active"},
			},
		},
	}
	
	err := manager.RegisterDevice(ctx, device)
	assert.NoError(t, err)
	
	// Update status
	err = manager.UpdateDeviceStatus(ctx, "device1", "maintenance")
	assert.NoError(t, err)
	
	// Verify status was updated
	updatedDevice, err := manager.GetDevice(ctx, "device1")
	assert.NoError(t, err)
	assert.Equal(t, "maintenance", updatedDevice.Status)
	
	// Test updating non-existent device
	err = manager.UpdateDeviceStatus(ctx, "nonexistent", "active")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device not found")
}

func TestIoTManager_UpdateDeviceLocation(t *testing.T) {
	manager := NewIoTManager()
	ctx := context.Background()
	
	device := &IoTDevice{
		ID:       "device1",
		Name:     "Test Device",
		Location: &DeviceLocation{Building: "Building A", Room: "Room 1"},
	}
	
	err := manager.RegisterDevice(ctx, device)
	assert.NoError(t, err)
	
	// Update location
	newLocation := &DeviceLocation{
		Building: "Building B",
		Room:     "Room 2",
		Floor:    2,
	}
	
	err = manager.UpdateDeviceLocation(ctx, "device1", newLocation)
	assert.NoError(t, err)
	
	// Verify location was updated
	updatedDevice, err := manager.GetDevice(ctx, "device1")
	assert.NoError(t, err)
	assert.Equal(t, "Building B", updatedDevice.Location.Building)
	assert.Equal(t, "Room 2", updatedDevice.Location.Room)
	assert.Equal(t, 2, updatedDevice.Location.Floor)
	
	// Test updating non-existent device
	err = manager.UpdateDeviceLocation(ctx, "nonexistent", newLocation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device not found")
}

func TestIoTManager_GetMetrics(t *testing.T) {
	manager := NewIoTManager()
	ctx := context.Background()
	
	// Test initial metrics
	metrics, err := manager.GetMetrics(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, metrics.TotalDevices)
	assert.Equal(t, 0, metrics.ActiveDevices)
	assert.Equal(t, 0, metrics.TotalSensors)
	assert.Equal(t, 0, metrics.ActiveSensors)
	assert.Equal(t, 0, metrics.TotalActuators)
	assert.Equal(t, 0, metrics.ActiveActuators)
	assert.Equal(t, 0, metrics.DataPoints)
	assert.Equal(t, 0, metrics.CommandsSent)
	
	// Register device with sensors and actuators
	device := &IoTDevice{
		ID:     "device1",
		Name:   "Test Device",
		Status: "active",
		Capabilities: &DeviceCapabilities{
			Sensors: []*Sensor{
				{ID: "sensor1", Status: "active"},
				{ID: "sensor2", Status: "inactive"},
			},
			Actuators: []*Actuator{
				{ID: "actuator1", Status: "active"},
			},
		},
	}
	
	err = manager.RegisterDevice(ctx, device)
	assert.NoError(t, err)
	
	// Send some sensor data and commands
	sensorData := &SensorData{DeviceID: "device1", SensorID: "sensor1", Value: 25.0}
	command := &ActuatorCommand{DeviceID: "device1", ActuatorID: "actuator1", Command: "turn_on"}
	
	err = manager.SendSensorData(ctx, sensorData)
	assert.NoError(t, err)
	err = manager.SendActuatorCommand(ctx, command)
	assert.NoError(t, err)
	
	// Test updated metrics
	metrics, err = manager.GetMetrics(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, metrics.TotalDevices)
	assert.Equal(t, 1, metrics.ActiveDevices)
	assert.Equal(t, 2, metrics.TotalSensors)
	assert.Equal(t, 1, metrics.ActiveSensors)
	assert.Equal(t, 1, metrics.TotalActuators)
	assert.Equal(t, 1, metrics.ActiveActuators)
	assert.Equal(t, 1, metrics.DataPoints)
	assert.Equal(t, 1, metrics.CommandsSent)
}

func TestIoTManager_isLocationMatch(t *testing.T) {
	manager := NewIoTManager()
	
	// Test nil locations
	assert.False(t, manager.isLocationMatch(nil, nil))
	assert.False(t, manager.isLocationMatch(&DeviceLocation{}, nil))
	assert.False(t, manager.isLocationMatch(nil, &DeviceLocation{}))
	
	// Test matching locations
	loc1 := &DeviceLocation{Building: "Building A", Room: "Room 1", Floor: 1}
	loc2 := &DeviceLocation{Building: "Building A", Room: "Room 1", Floor: 1}
	assert.True(t, manager.isLocationMatch(loc1, loc2))
	
	// Test different buildings
	loc3 := &DeviceLocation{Building: "Building B", Room: "Room 1", Floor: 1}
	assert.False(t, manager.isLocationMatch(loc1, loc3))
	
	// Test different rooms
	loc4 := &DeviceLocation{Building: "Building A", Room: "Room 2", Floor: 1}
	assert.False(t, manager.isLocationMatch(loc1, loc4))
	
	// Test different floors (when both specified)
	loc5 := &DeviceLocation{Building: "Building A", Room: "Room 1", Floor: 2}
	assert.False(t, manager.isLocationMatch(loc1, loc5))
	
	// Test floor not specified (should match)
	loc6 := &DeviceLocation{Building: "Building A", Room: "Room 1", Floor: 0}
	assert.True(t, manager.isLocationMatch(loc1, loc6))
	assert.True(t, manager.isLocationMatch(loc6, loc1))
}
