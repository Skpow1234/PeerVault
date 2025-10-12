package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
	"github.com/Skpow1234/Peervault/internal/cli/iot"
)

// IoTCommand handles IoT device management commands
type IoTCommand struct {
	BaseCommand
	deviceManager *iot.DeviceManager
}

// NewIoTCommand creates a new IoT command
func NewIoTCommand(client *client.Client, formatter *formatter.Formatter, deviceManager *iot.DeviceManager) *IoTCommand {
	return &IoTCommand{
		BaseCommand: BaseCommand{
			name:        "iot",
			description: "IoT device operations",
			usage:       "iot [command] [options]",
			client:      client,
			formatter:   formatter,
		},
		deviceManager: deviceManager,
	}
}

// Execute executes the IoT command
func (c *IoTCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showHelp()
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "add-device":
		return c.addDevice(ctx, subArgs)
	case "remove-device":
		return c.removeDevice(ctx, subArgs)
	case "list-devices":
		return c.listDevices(ctx, subArgs)
	case "get-device":
		return c.getDevice(ctx, subArgs)
	case "update-status":
		return c.updateDeviceStatus(ctx, subArgs)
	case "send-sensor-data":
		return c.sendSensorData(ctx, subArgs)
	case "get-sensor-data":
		return c.getSensorData(ctx, subArgs)
	case "send-command":
		return c.sendActuatorCommand(ctx, subArgs)
	case "get-commands":
		return c.getActuatorCommands(ctx, subArgs)
	case "schedule-update":
		return c.scheduleFirmwareUpdate(ctx, subArgs)
	case "get-updates":
		return c.getFirmwareUpdates(ctx, subArgs)
	case "update-progress":
		return c.updateFirmwareProgress(ctx, subArgs)
	case "stats":
		return c.getStatistics(ctx, subArgs)
	default:
		return c.showHelp()
	}
}

// addDevice adds a new IoT device
func (c *IoTCommand) addDevice(ctx context.Context, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: iot add-device <id> <name> <type> <location> [ip] [mac] [firmware]")
	}

	device := &iot.Device{
		ID:           args[0],
		Name:         args[1],
		Type:         args[2],
		Location:     args[3],
		Status:       "offline",
		IPAddress:    "0.0.0.0",
		MACAddress:   "00:00:00:00:00:00",
		Firmware:     "1.0.0",
		LastSeen:     time.Now(),
		Capabilities: []string{"sensor", "actuator"},
		Metadata:     make(map[string]string),
	}

	if len(args) > 4 {
		device.IPAddress = args[4]
	}
	if len(args) > 5 {
		device.MACAddress = args[5]
	}
	if len(args) > 6 {
		device.Firmware = args[6]
	}

	err := c.deviceManager.AddDevice(ctx, device)
	if err != nil {
		return fmt.Errorf("failed to add device: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Device '%s' added successfully", device.ID))
	return nil
}

// removeDevice removes an IoT device
func (c *IoTCommand) removeDevice(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: iot remove-device <device-id>")
	}

	deviceID := args[0]
	err := c.deviceManager.RemoveDevice(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to remove device: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Device '%s' removed successfully", deviceID))
	return nil
}

// listDevices lists all IoT devices
func (c *IoTCommand) listDevices(ctx context.Context, args []string) error {
	devices, err := c.deviceManager.ListDevices(ctx)
	if err != nil {
		return fmt.Errorf("failed to list devices: %v", err)
	}

	if len(devices) == 0 {
		c.formatter.PrintInfo("No devices found")
		return nil
	}

	headers := []string{"ID", "Name", "Type", "Status", "Location", "IP Address", "Firmware", "Last Seen"}
	rows := make([][]string, len(devices))

	for i, device := range devices {
		rows[i] = []string{
			device.ID,
			device.Name,
			device.Type,
			device.Status,
			device.Location,
			device.IPAddress,
			device.Firmware,
			device.LastSeen.Format("2006-01-02 15:04:05"),
		}
	}

	c.formatter.PrintTable(headers, rows)
	return nil
}

// getDevice gets a specific device
func (c *IoTCommand) getDevice(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: iot get-device <device-id>")
	}

	deviceID := args[0]
	device, err := c.deviceManager.GetDevice(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device: %v", err)
	}

	c.formatter.PrintInfo(fmt.Sprintf("Device: %s", device.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Name: %s", device.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  Type: %s", device.Type))
	c.formatter.PrintInfo(fmt.Sprintf("  Status: %s", device.Status))
	c.formatter.PrintInfo(fmt.Sprintf("  Location: %s", device.Location))
	c.formatter.PrintInfo(fmt.Sprintf("  IP Address: %s", device.IPAddress))
	c.formatter.PrintInfo(fmt.Sprintf("  MAC Address: %s", device.MACAddress))
	c.formatter.PrintInfo(fmt.Sprintf("  Firmware: %s", device.Firmware))
	c.formatter.PrintInfo(fmt.Sprintf("  Last Seen: %s", device.LastSeen.Format("2006-01-02 15:04:05")))
	c.formatter.PrintInfo(fmt.Sprintf("  Capabilities: %s", strings.Join(device.Capabilities, ", ")))

	return nil
}

// updateDeviceStatus updates device status
func (c *IoTCommand) updateDeviceStatus(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: iot update-status <device-id> <status>")
	}

	deviceID := args[0]
	status := args[1]

	err := c.deviceManager.UpdateDeviceStatus(ctx, deviceID, status)
	if err != nil {
		return fmt.Errorf("failed to update device status: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Device '%s' status updated to '%s'", deviceID, status))
	return nil
}

// sendSensorData sends sensor data
func (c *IoTCommand) sendSensorData(ctx context.Context, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: iot send-sensor-data <device-id> <sensor-type> <value> <unit> [location] [quality]")
	}

	deviceID := args[0]
	sensorType := args[1]
	value, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid value: %v", err)
	}
	unit := args[3]

	location := "unknown"
	quality := "good"
	if len(args) > 4 {
		location = args[4]
	}
	if len(args) > 5 {
		quality = args[5]
	}

	data := &iot.SensorData{
		DeviceID:   deviceID,
		SensorType: sensorType,
		Value:      value,
		Unit:       unit,
		Timestamp:  time.Now(),
		Location:   location,
		Quality:    quality,
		Metadata:   make(map[string]interface{}),
	}

	err = c.deviceManager.SendSensorData(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to send sensor data: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Sensor data sent for device '%s': %s=%.2f %s", deviceID, sensorType, value, unit))
	return nil
}

// getSensorData gets sensor data for a device
func (c *IoTCommand) getSensorData(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: iot get-sensor-data <device-id> [limit]")
	}

	deviceID := args[0]
	limit := 10
	if len(args) > 1 {
		var err error
		limit, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid limit: %v", err)
		}
	}

	data, err := c.deviceManager.GetSensorData(ctx, deviceID, limit)
	if err != nil {
		return fmt.Errorf("failed to get sensor data: %v", err)
	}

	if len(data) == 0 {
		c.formatter.PrintInfo("No sensor data found")
		return nil
	}

	headers := []string{"Device ID", "Sensor Type", "Value", "Unit", "Timestamp", "Location", "Quality"}
	rows := make([][]string, len(data))

	for i, d := range data {
		rows[i] = []string{
			d.DeviceID,
			d.SensorType,
			fmt.Sprintf("%.2f", d.Value),
			d.Unit,
			d.Timestamp.Format("2006-01-02 15:04:05"),
			d.Location,
			d.Quality,
		}
	}

	c.formatter.PrintTable(headers, rows)
	return nil
}

// sendActuatorCommand sends a command to an actuator
func (c *IoTCommand) sendActuatorCommand(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: iot send-command <device-id> <actuator-id> <command> [priority]")
	}

	deviceID := args[0]
	actuatorID := args[1]
	command := args[2]
	priority := 1

	if len(args) > 3 {
		var err error
		priority, err = strconv.Atoi(args[3])
		if err != nil {
			return fmt.Errorf("invalid priority: %v", err)
		}
	}

	cmd := &iot.ActuatorCommand{
		DeviceID:   deviceID,
		ActuatorID: actuatorID,
		Command:    command,
		Parameters: make(map[string]string),
		Priority:   priority,
		Timestamp:  time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}

	err := c.deviceManager.SendActuatorCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to send actuator command: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Command sent to device '%s' actuator '%s': %s", deviceID, actuatorID, command))
	return nil
}

// getActuatorCommands gets commands for a device
func (c *IoTCommand) getActuatorCommands(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: iot get-commands <device-id>")
	}

	deviceID := args[0]
	commands, err := c.deviceManager.GetActuatorCommands(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get actuator commands: %v", err)
	}

	if len(commands) == 0 {
		c.formatter.PrintInfo("No commands found")
		return nil
	}

	headers := []string{"Device ID", "Actuator ID", "Command", "Priority", "Timestamp", "Expires At"}
	rows := make([][]string, len(commands))

	for i, cmd := range commands {
		rows[i] = []string{
			cmd.DeviceID,
			cmd.ActuatorID,
			cmd.Command,
			fmt.Sprintf("%d", cmd.Priority),
			cmd.Timestamp.Format("2006-01-02 15:04:05"),
			cmd.ExpiresAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.formatter.PrintTable(headers, rows)
	return nil
}

// scheduleFirmwareUpdate schedules a firmware update
func (c *IoTCommand) scheduleFirmwareUpdate(ctx context.Context, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: iot schedule-update <device-id> <version> <url> <checksum> [size]")
	}

	deviceID := args[0]
	version := args[1]
	url := args[2]
	checksum := args[3]
	size := int64(0)

	if len(args) > 4 {
		var err error
		size, err = strconv.ParseInt(args[4], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid size: %v", err)
		}
	}

	update := &iot.FirmwareUpdate{
		DeviceID:    deviceID,
		Version:     version,
		URL:         url,
		Checksum:    checksum,
		Size:        size,
		ScheduledAt: time.Now(),
		Status:      "scheduled",
		Progress:    0,
	}

	err := c.deviceManager.ScheduleFirmwareUpdate(ctx, update)
	if err != nil {
		return fmt.Errorf("failed to schedule firmware update: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Firmware update scheduled for device '%s': version %s", deviceID, version))
	return nil
}

// getFirmwareUpdates gets firmware updates for a device
func (c *IoTCommand) getFirmwareUpdates(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: iot get-updates <device-id>")
	}

	deviceID := args[0]
	updates, err := c.deviceManager.GetFirmwareUpdates(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get firmware updates: %v", err)
	}

	if len(updates) == 0 {
		c.formatter.PrintInfo("No firmware updates found")
		return nil
	}

	headers := []string{"Device ID", "Version", "Status", "Progress", "Size", "Scheduled At"}
	rows := make([][]string, len(updates))

	for i, update := range updates {
		rows[i] = []string{
			update.DeviceID,
			update.Version,
			update.Status,
			fmt.Sprintf("%d%%", update.Progress),
			c.formatter.FormatBytes(update.Size),
			update.ScheduledAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.formatter.PrintTable(headers, rows)
	return nil
}

// updateFirmwareProgress updates firmware update progress
func (c *IoTCommand) updateFirmwareProgress(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: iot update-progress <device-id> <version> <progress>")
	}

	deviceID := args[0]
	version := args[1]
	progress, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid progress: %v", err)
	}

	if progress < 0 || progress > 100 {
		return fmt.Errorf("progress must be between 0 and 100")
	}

	err = c.deviceManager.UpdateFirmwareProgress(ctx, deviceID, version, progress)
	if err != nil {
		return fmt.Errorf("failed to update firmware progress: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Firmware update progress updated for device '%s' version '%s': %d%%", deviceID, version, progress))
	return nil
}

// getStatistics gets IoT statistics
func (c *IoTCommand) getStatistics(ctx context.Context, args []string) error {
	stats, err := c.deviceManager.GetDeviceStatistics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get statistics: %v", err)
	}

	c.formatter.PrintInfo("IoT Device Statistics:")
	c.formatter.PrintInfo(fmt.Sprintf("  Total Devices: %v", stats["total_devices"]))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Sensor Data: %v", stats["total_sensor_data"]))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Commands: %v", stats["total_commands"]))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Updates: %v", stats["total_updates"]))

	if devicesByStatus, ok := stats["devices_by_status"].(map[string]int); ok {
		c.formatter.PrintInfo("  Devices by Status:")
		for status, count := range devicesByStatus {
			c.formatter.PrintInfo(fmt.Sprintf("    %s: %d", status, count))
		}
	}

	if devicesByType, ok := stats["devices_by_type"].(map[string]int); ok {
		c.formatter.PrintInfo("  Devices by Type:")
		for deviceType, count := range devicesByType {
			c.formatter.PrintInfo(fmt.Sprintf("    %s: %d", deviceType, count))
		}
	}

	return nil
}

// showHelp shows help information
func (c *IoTCommand) showHelp() error {
	c.formatter.PrintInfo("IoT Device Management Commands:")
	c.formatter.PrintInfo("  add-device <id> <name> <type> <location> [ip] [mac] [firmware] - Add a new IoT device")
	c.formatter.PrintInfo("  remove-device <device-id> - Remove an IoT device")
	c.formatter.PrintInfo("  list-devices - List all IoT devices")
	c.formatter.PrintInfo("  get-device <device-id> - Get device details")
	c.formatter.PrintInfo("  update-status <device-id> <status> - Update device status")
	c.formatter.PrintInfo("  send-sensor-data <device-id> <sensor-type> <value> <unit> [location] [quality] - Send sensor data")
	c.formatter.PrintInfo("  get-sensor-data <device-id> [limit] - Get sensor data for a device")
	c.formatter.PrintInfo("  send-command <device-id> <actuator-id> <command> [priority] - Send actuator command")
	c.formatter.PrintInfo("  get-commands <device-id> - Get actuator commands for a device")
	c.formatter.PrintInfo("  schedule-update <device-id> <version> <url> <checksum> [size] - Schedule firmware update")
	c.formatter.PrintInfo("  get-updates <device-id> - Get firmware updates for a device")
	c.formatter.PrintInfo("  update-progress <device-id> <version> <progress> - Update firmware progress")
	c.formatter.PrintInfo("  stats - Show IoT statistics")
	return nil
}
