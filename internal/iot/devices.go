package iot

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// IoTDevice represents an IoT device
type IoTDevice struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Manufacturer string                 `json:"manufacturer"`
	Model        string                 `json:"model"`
	Version      string                 `json:"version"`
	Status       string                 `json:"status"`
	Location     *DeviceLocation        `json:"location"`
	Capabilities *DeviceCapabilities    `json:"capabilities"`
	Protocols    []string               `json:"protocols"`
	LastSeen     time.Time              `json:"last_seen"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// DeviceLocation represents the location of an IoT device
type DeviceLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
	Room      string  `json:"room"`
	Building  string  `json:"building"`
	Floor     int     `json:"floor"`
}

// DeviceCapabilities represents the capabilities of an IoT device
type DeviceCapabilities struct {
	Sensors      []*Sensor       `json:"sensors"`
	Actuators    []*Actuator     `json:"actuators"`
	Connectivity *Connectivity   `json:"connectivity"`
	Power        *PowerSpec      `json:"power"`
	Storage      *StorageSpec    `json:"storage"`
	Processing   *ProcessingSpec `json:"processing"`
}

// Sensor represents a sensor on an IoT device
type Sensor struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Unit         string                 `json:"unit"`
	Range        *SensorRange           `json:"range"`
	Accuracy     float64                `json:"accuracy"`
	Resolution   float64                `json:"resolution"`
	SamplingRate float64                `json:"sampling_rate"`
	Status       string                 `json:"status"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// SensorRange represents the range of a sensor
type SensorRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// Actuator represents an actuator on an IoT device
type Actuator struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	ControlType string                 `json:"control_type"`
	Range       *ActuatorRange         `json:"range"`
	Precision   float64                `json:"precision"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ActuatorRange represents the range of an actuator
type ActuatorRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// Connectivity represents the connectivity capabilities of an IoT device
type Connectivity struct {
	WiFi      *WiFiSpec      `json:"wifi,omitempty"`
	Bluetooth *BluetoothSpec `json:"bluetooth,omitempty"`
	Zigbee    *ZigbeeSpec    `json:"zigbee,omitempty"`
	ZWave     *ZWaveSpec     `json:"zwave,omitempty"`
	LoRaWAN   *LoRaWANSpec   `json:"lorawan,omitempty"`
	NB_IoT    *NBIoTSpec     `json:"nb_iot,omitempty"`
	Ethernet  *EthernetSpec  `json:"ethernet,omitempty"`
	Serial    *SerialSpec    `json:"serial,omitempty"`
}

// WiFiSpec represents WiFi connectivity specifications
type WiFiSpec struct {
	SSID       string   `json:"ssid"`
	BSSID      string   `json:"bssid"`
	Frequency  float64  `json:"frequency"` // GHz
	Channel    int      `json:"channel"`
	Signal     float64  `json:"signal"` // dBm
	Encryption string   `json:"encryption"`
	Protocols  []string `json:"protocols"`
}

// BluetoothSpec represents Bluetooth connectivity specifications
type BluetoothSpec struct {
	Version  string   `json:"version"`
	Class    string   `json:"class"`
	Address  string   `json:"address"`
	Signal   float64  `json:"signal"` // dBm
	Services []string `json:"services"`
	Profiles []string `json:"profiles"`
}

// ZigbeeSpec represents Zigbee connectivity specifications
type ZigbeeSpec struct {
	PANID      string   `json:"panid"`
	Channel    int      `json:"channel"`
	NetworkKey string   `json:"network_key"`
	DeviceType string   `json:"device_type"`
	Endpoints  []string `json:"endpoints"`
}

// ZWaveSpec represents Z-Wave connectivity specifications
type ZWaveSpec struct {
	HomeID         string   `json:"home_id"`
	NodeID         int      `json:"node_id"`
	DeviceType     string   `json:"device_type"`
	CommandClasses []string `json:"command_classes"`
}

// LoRaWANSpec represents LoRaWAN connectivity specifications
type LoRaWANSpec struct {
	DevEUI          string  `json:"dev_eui"`
	AppEUI          string  `json:"app_eui"`
	AppKey          string  `json:"app_key"`
	Frequency       float64 `json:"frequency"` // MHz
	SpreadingFactor int     `json:"spreading_factor"`
	Bandwidth       int     `json:"bandwidth"` // kHz
}

// NBIoTSpec represents NB-IoT connectivity specifications
type NBIoTSpec struct {
	IMSI     string  `json:"imsi"`
	IMEI     string  `json:"imei"`
	Operator string  `json:"operator"`
	Signal   float64 `json:"signal"` // dBm
	CellID   string  `json:"cell_id"`
}

// EthernetSpec represents Ethernet connectivity specifications
type EthernetSpec struct {
	MACAddress string `json:"mac_address"`
	Speed      int    `json:"speed"` // Mbps
	Duplex     string `json:"duplex"`
	IPAddress  string `json:"ip_address"`
	SubnetMask string `json:"subnet_mask"`
	Gateway    string `json:"gateway"`
}

// SerialSpec represents serial connectivity specifications
type SerialSpec struct {
	Port        string `json:"port"`
	BaudRate    int    `json:"baud_rate"`
	DataBits    int    `json:"data_bits"`
	StopBits    int    `json:"stop_bits"`
	Parity      string `json:"parity"`
	FlowControl string `json:"flow_control"`
}

// PowerSpec represents power specifications
type PowerSpec struct {
	Type     string  `json:"type"`     // battery, mains, solar, etc.
	Voltage  float64 `json:"voltage"`  // Volts
	Current  float64 `json:"current"`  // Amperes
	Capacity float64 `json:"capacity"` // mAh or Wh
	Level    float64 `json:"level"`    // Percentage
	Status   string  `json:"status"`
}

// StorageSpec represents storage specifications
type StorageSpec struct {
	Type      string  `json:"type"`       // flash, eeprom, sd, etc.
	Size      int64   `json:"size"`       // Bytes
	Used      int64   `json:"used"`       // Bytes
	Available int64   `json:"available"`  // Bytes
	WearLevel float64 `json:"wear_level"` // Percentage
}

// ProcessingSpec represents processing specifications
type ProcessingSpec struct {
	CPU          string  `json:"cpu"`
	Frequency    float64 `json:"frequency"` // MHz
	Cores        int     `json:"cores"`
	RAM          int64   `json:"ram"`   // Bytes
	Flash        int64   `json:"flash"` // Bytes
	Architecture string  `json:"architecture"`
}

// SensorData represents data from a sensor
type SensorData struct {
	DeviceID  string                 `json:"device_id"`
	SensorID  string                 `json:"sensor_id"`
	Value     float64                `json:"value"`
	Unit      string                 `json:"unit"`
	Timestamp time.Time              `json:"timestamp"`
	Quality   string                 `json:"quality"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ActuatorCommand represents a command to an actuator
type ActuatorCommand struct {
	DeviceID   string                 `json:"device_id"`
	ActuatorID string                 `json:"actuator_id"`
	Command    string                 `json:"command"`
	Value      float64                `json:"value"`
	Unit       string                 `json:"unit"`
	Timestamp  time.Time              `json:"timestamp"`
	Priority   int                    `json:"priority"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// IoTManager provides IoT device management functionality
type IoTManager struct {
	devices    map[string]*IoTDevice
	sensorData map[string][]*SensorData
	commands   map[string][]*ActuatorCommand
	mu         sync.RWMutex
	metrics    *IoTMetrics
}

// IoTMetrics represents IoT system metrics
type IoTMetrics struct {
	TotalDevices    int     `json:"total_devices"`
	ActiveDevices   int     `json:"active_devices"`
	TotalSensors    int     `json:"total_sensors"`
	ActiveSensors   int     `json:"active_sensors"`
	TotalActuators  int     `json:"total_actuators"`
	ActiveActuators int     `json:"active_actuators"`
	DataPoints      int     `json:"data_points"`
	CommandsSent    int     `json:"commands_sent"`
	AverageLatency  float64 `json:"average_latency"`
	DataThroughput  float64 `json:"data_throughput"`
}

// NewIoTManager creates a new IoT manager
func NewIoTManager() *IoTManager {
	return &IoTManager{
		devices:    make(map[string]*IoTDevice),
		sensorData: make(map[string][]*SensorData),
		commands:   make(map[string][]*ActuatorCommand),
		metrics:    &IoTMetrics{},
	}
}

// RegisterDevice registers an IoT device
func (im *IoTManager) RegisterDevice(ctx context.Context, device *IoTDevice) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	device.Status = "active"
	device.LastSeen = time.Now()
	im.devices[device.ID] = device

	im.updateMetrics()

	return nil
}

// UnregisterDevice unregisters an IoT device
func (im *IoTManager) UnregisterDevice(ctx context.Context, deviceID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if device, exists := im.devices[deviceID]; exists {
		device.Status = "inactive"
		im.updateMetrics()
	}

	return nil
}

// GetDevice retrieves an IoT device by ID
func (im *IoTManager) GetDevice(ctx context.Context, deviceID string) (*IoTDevice, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	device, exists := im.devices[deviceID]
	if !exists {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	return device, nil
}

// ListDevices lists all IoT devices
func (im *IoTManager) ListDevices(ctx context.Context) ([]*IoTDevice, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	devices := make([]*IoTDevice, 0, len(im.devices))
	for _, device := range im.devices {
		devices = append(devices, device)
	}

	return devices, nil
}

// ListDevicesByType lists devices by type
func (im *IoTManager) ListDevicesByType(ctx context.Context, deviceType string) ([]*IoTDevice, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	var devices []*IoTDevice
	for _, device := range im.devices {
		if device.Type == deviceType {
			devices = append(devices, device)
		}
	}

	return devices, nil
}

// ListDevicesByLocation lists devices by location
func (im *IoTManager) ListDevicesByLocation(ctx context.Context, location *DeviceLocation) ([]*IoTDevice, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	var devices []*IoTDevice
	for _, device := range im.devices {
		if device.Location != nil && im.isLocationMatch(device.Location, location) {
			devices = append(devices, device)
		}
	}

	return devices, nil
}

// SendSensorData sends sensor data from a device
func (im *IoTManager) SendSensorData(ctx context.Context, data *SensorData) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Validate device exists
	device, exists := im.devices[data.DeviceID]
	if !exists {
		return fmt.Errorf("device not found: %s", data.DeviceID)
	}

	// Validate sensor exists
	sensorExists := false
	for _, sensor := range device.Capabilities.Sensors {
		if sensor.ID == data.SensorID {
			sensorExists = true
			break
		}
	}

	if !sensorExists {
		return fmt.Errorf("sensor not found: %s", data.SensorID)
	}

	// Store sensor data
	im.sensorData[data.DeviceID] = append(im.sensorData[data.DeviceID], data)

	// Update device last seen
	device.LastSeen = time.Now()

	im.updateMetrics()

	return nil
}

// GetSensorData retrieves sensor data for a device
func (im *IoTManager) GetSensorData(ctx context.Context, deviceID string, sensorID string, limit int) ([]*SensorData, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	deviceData, exists := im.sensorData[deviceID]
	if !exists {
		return nil, fmt.Errorf("no data found for device: %s", deviceID)
	}

	var filteredData []*SensorData
	for _, data := range deviceData {
		if sensorID == "" || data.SensorID == sensorID {
			filteredData = append(filteredData, data)
		}
	}

	// Limit results
	if limit > 0 && len(filteredData) > limit {
		filteredData = filteredData[len(filteredData)-limit:]
	}

	return filteredData, nil
}

// SendActuatorCommand sends a command to an actuator
func (im *IoTManager) SendActuatorCommand(ctx context.Context, command *ActuatorCommand) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Validate device exists
	device, exists := im.devices[command.DeviceID]
	if !exists {
		return fmt.Errorf("device not found: %s", command.DeviceID)
	}

	// Validate actuator exists
	actuatorExists := false
	for _, actuator := range device.Capabilities.Actuators {
		if actuator.ID == command.ActuatorID {
			actuatorExists = true
			break
		}
	}

	if !actuatorExists {
		return fmt.Errorf("actuator not found: %s", command.ActuatorID)
	}

	// Store command
	im.commands[command.DeviceID] = append(im.commands[command.DeviceID], command)

	// Update device last seen
	device.LastSeen = time.Now()

	im.updateMetrics()

	return nil
}

// GetActuatorCommands retrieves commands for a device
func (im *IoTManager) GetActuatorCommands(ctx context.Context, deviceID string, actuatorID string, limit int) ([]*ActuatorCommand, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	deviceCommands, exists := im.commands[deviceID]
	if !exists {
		return nil, fmt.Errorf("no commands found for device: %s", deviceID)
	}

	var filteredCommands []*ActuatorCommand
	for _, command := range deviceCommands {
		if actuatorID == "" || command.ActuatorID == actuatorID {
			filteredCommands = append(filteredCommands, command)
		}
	}

	// Limit results
	if limit > 0 && len(filteredCommands) > limit {
		filteredCommands = filteredCommands[len(filteredCommands)-limit:]
	}

	return filteredCommands, nil
}

// GetMetrics returns IoT system metrics
func (im *IoTManager) GetMetrics(ctx context.Context) (*IoTMetrics, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return im.metrics, nil
}

// UpdateDeviceStatus updates the status of a device
func (im *IoTManager) UpdateDeviceStatus(ctx context.Context, deviceID string, status string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	device, exists := im.devices[deviceID]
	if !exists {
		return fmt.Errorf("device not found: %s", deviceID)
	}

	device.Status = status
	device.LastSeen = time.Now()

	im.updateMetrics()

	return nil
}

// UpdateDeviceLocation updates the location of a device
func (im *IoTManager) UpdateDeviceLocation(ctx context.Context, deviceID string, location *DeviceLocation) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	device, exists := im.devices[deviceID]
	if !exists {
		return fmt.Errorf("device not found: %s", deviceID)
	}

	device.Location = location
	device.LastSeen = time.Now()

	return nil
}

// isLocationMatch checks if two locations match
func (im *IoTManager) isLocationMatch(loc1, loc2 *DeviceLocation) bool {
	if loc1 == nil || loc2 == nil {
		return false
	}

	// Check if locations are in the same building and room
	if loc1.Building != loc2.Building || loc1.Room != loc2.Room {
		return false
	}

	// Check if floors are the same (if specified)
	if loc1.Floor != 0 && loc2.Floor != 0 && loc1.Floor != loc2.Floor {
		return false
	}

	return true
}

// updateMetrics updates the IoT system metrics
func (im *IoTManager) updateMetrics() {
	im.metrics.TotalDevices = len(im.devices)
	im.metrics.ActiveDevices = 0
	im.metrics.TotalSensors = 0
	im.metrics.ActiveSensors = 0
	im.metrics.TotalActuators = 0
	im.metrics.ActiveActuators = 0
	im.metrics.DataPoints = 0
	im.metrics.CommandsSent = 0

	for _, device := range im.devices {
		if device.Status == "active" {
			im.metrics.ActiveDevices++
		}

		// Count sensors
		for _, sensor := range device.Capabilities.Sensors {
			im.metrics.TotalSensors++
			if sensor.Status == "active" {
				im.metrics.ActiveSensors++
			}
		}

		// Count actuators
		for _, actuator := range device.Capabilities.Actuators {
			im.metrics.TotalActuators++
			if actuator.Status == "active" {
				im.metrics.ActiveActuators++
			}
		}
	}

	// Count data points and commands
	for _, data := range im.sensorData {
		im.metrics.DataPoints += len(data)
	}

	for _, commands := range im.commands {
		im.metrics.CommandsSent += len(commands)
	}
}
