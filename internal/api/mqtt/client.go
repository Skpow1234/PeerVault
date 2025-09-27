package mqtt

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"
)

// Client represents an MQTT client connection
type Client struct {
	// Connection
	conn net.Conn

	// Client identification
	ID string

	// Broker reference
	broker *Broker

	// Logger
	logger *slog.Logger

	// Client state
	connected    bool
	cleanSession bool
	keepAlive    time.Duration

	// Will message
	WillMessage *Message

	// Subscriptions
	subscriptions map[string]QoS
	subsMu        sync.RWMutex

	// Message handling
	incomingMessages chan *Message
	outgoingMessages chan *Message

	// QoS handling
	pendingPublishes map[uint16]*PendingPublish
	pendingPubacks   map[uint16]*PendingPuback
	pendingPubrecs   map[uint16]*PendingPubrec
	pendingPubrels   map[uint16]*PendingPubrel
	pendingPubcomps  map[uint16]*PendingPubcomp
	pendingMu        sync.RWMutex

	// Statistics
	stats *ClientStats

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// ClientStats holds client statistics
type ClientStats struct {
	ConnectedAt      time.Time
	MessagesSent     int
	MessagesReceived int
	BytesSent        int64
	BytesReceived    int64
	LastActivity     time.Time
}

// PendingPublish represents a pending publish message
type PendingPublish struct {
	Message   *Message
	Timestamp time.Time
	Retries   int
}

// PendingPuback represents a pending puback message
type PendingPuback struct {
	Timestamp time.Time
	Retries   int
}

// PendingPubrec represents a pending pubrec message
type PendingPubrec struct {
	Timestamp time.Time
	Retries   int
}

// PendingPubrel represents a pending pubrel message
type PendingPubrel struct {
	Timestamp time.Time
	Retries   int
}

// PendingPubcomp represents a pending pubcomp message
type PendingPubcomp struct {
	Timestamp time.Time
	Retries   int
}

// NewClient creates a new MQTT client
func NewClient(conn net.Conn, broker *Broker, logger *slog.Logger) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		conn:             conn,
		ID:               generateClientID(),
		broker:           broker,
		logger:           logger,
		connected:        false,
		cleanSession:     true,
		keepAlive:        60 * time.Second,
		subscriptions:    make(map[string]QoS),
		incomingMessages: make(chan *Message, 100),
		outgoingMessages: make(chan *Message, 100),
		pendingPublishes: make(map[uint16]*PendingPublish),
		pendingPubacks:   make(map[uint16]*PendingPuback),
		pendingPubrecs:   make(map[uint16]*PendingPubrec),
		pendingPubrels:   make(map[uint16]*PendingPubrel),
		pendingPubcomps:  make(map[uint16]*PendingPubcomp),
		stats: &ClientStats{
			ConnectedAt:  time.Now(),
			LastActivity: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	return client
}

// generateClientID generates a unique client ID
func generateClientID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes) // Ignore error for client ID generation
	return hex.EncodeToString(bytes)
}

// Handle handles the client session
func (c *Client) Handle() error {
	defer c.Close()

	// Add client to broker
	c.broker.addClient(c)

	// Start message processing goroutines
	go c.processIncomingMessages()
	go c.processOutgoingMessages()
	go c.handleKeepAlive()

	// Read packets from connection
	for {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		default:
		}

		// Set read timeout
		if err := c.conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
			c.logger.Error("Failed to set read deadline", "error", err)
			return err
		}

		packet, err := c.readPacket()
		if err != nil {
			if err == io.EOF {
				c.logger.Info("Client disconnected", "clientId", c.ID)
				return nil
			}
			return fmt.Errorf("failed to read packet: %w", err)
		}

		// Update activity
		c.stats.LastActivity = time.Now()

		// Process packet
		if err := c.processPacket(packet); err != nil {
			c.logger.Error("Failed to process packet", "error", err, "clientId", c.ID)
			return err
		}
	}
}

// readPacket reads a packet from the connection
func (c *Client) readPacket() (*Packet, error) {
	// Read fixed header
	header, err := c.readFixedHeader()
	if err != nil {
		return nil, err
	}

	// Read variable header and payload
	packet := &Packet{
		Header: header,
	}

	// Read remaining length
	remainingLength, err := c.readRemainingLength()
	if err != nil {
		return nil, err
	}

	if remainingLength > 0 {
		// Read variable header and payload
		data := make([]byte, remainingLength)
		if _, err := io.ReadFull(c.conn, data); err != nil {
			return nil, err
		}

		packet.Data = data
		c.stats.BytesReceived += int64(remainingLength)
	}

	return packet, nil
}

// readFixedHeader reads the fixed header from the connection
func (c *Client) readFixedHeader() (*FixedHeader, error) {
	headerBytes := make([]byte, 1)
	if _, err := io.ReadFull(c.conn, headerBytes); err != nil {
		return nil, err
	}

	header := &FixedHeader{
		MessageType: MessageType(headerBytes[0] >> 4),
		Flags:       headerBytes[0] & 0x0F,
	}

	return header, nil
}

// readRemainingLength reads the remaining length from the connection
func (c *Client) readRemainingLength() (int, error) {
	var multiplier = 1
	var value = 0

	for {
		encodedByte := make([]byte, 1)
		if _, err := io.ReadFull(c.conn, encodedByte); err != nil {
			return 0, err
		}

		value += int(encodedByte[0]&127) * multiplier
		multiplier *= 128

		if multiplier > 128*128*128 {
			return 0, fmt.Errorf("malformed remaining length")
		}

		if (encodedByte[0] & 128) == 0 {
			break
		}
	}

	return value, nil
}

// processPacket processes a received packet
func (c *Client) processPacket(packet *Packet) error {
	switch packet.Header.MessageType {
	case CONNECT:
		return c.handleConnect(packet)
	case CONNACK:
		return c.handleConnack(packet)
	case PUBLISH:
		return c.handlePublish(packet)
	case PUBACK:
		return c.handlePuback(packet)
	case PUBREC:
		return c.handlePubrec(packet)
	case PUBREL:
		return c.handlePubrel(packet)
	case PUBCOMP:
		return c.handlePubcomp(packet)
	case SUBSCRIBE:
		return c.handleSubscribe(packet)
	case SUBACK:
		return c.handleSuback(packet)
	case UNSUBSCRIBE:
		return c.handleUnsubscribe(packet)
	case UNSUBACK:
		return c.handleUnsuback(packet)
	case PINGREQ:
		return c.handlePingreq(packet)
	case PINGRESP:
		return c.handlePingresp(packet)
	case DISCONNECT:
		return c.handleDisconnect(packet)
	default:
		return fmt.Errorf("unknown message type: %d", packet.Header.MessageType)
	}
}

// handleConnect handles a CONNECT packet
func (c *Client) handleConnect(packet *Packet) error {
	// Parse CONNECT packet
	connect, err := c.parseConnectPacket(packet)
	if err != nil {
		return err
	}

	// Validate client ID
	if connect.ClientID == "" {
		connect.ClientID = c.ID
	}

	// Update client properties
	c.ID = connect.ClientID
	c.cleanSession = connect.CleanSession
	c.keepAlive = time.Duration(connect.KeepAlive) * time.Second

	// Handle will message
	if connect.WillFlag {
		c.WillMessage = &Message{
			Topic:   connect.WillTopic,
			Payload: connect.WillMessage,
			QoS:     QoS(connect.WillQoS),
			Retain:  connect.WillRetain,
		}
	}

	// Send CONNACK
	connack := &ConnackPacket{
		SessionPresent: false, // TODO: Implement session persistence
		ReturnCode:     CONNACK_ACCEPTED,
	}

	if err := c.sendConnack(connack); err != nil {
		return err
	}

	c.connected = true
	c.logger.Info("Client connected", "clientId", c.ID, "keepAlive", c.keepAlive)

	return nil
}

// handlePublish handles a PUBLISH packet
func (c *Client) handlePublish(packet *Packet) error {
	// Parse PUBLISH packet
	publish, err := c.parsePublishPacket(packet)
	if err != nil {
		return err
	}

	// Create message
	message := &Message{
		Topic:   publish.Topic,
		Payload: publish.Payload,
		QoS:     publish.QoS,
		Retain:  publish.Retain,
	}

	// Handle QoS
	switch publish.QoS {
	case QoS0:
		// No acknowledgment required
		return c.broker.publishMessage(message)
	case QoS1:
		// Send PUBACK
		if err := c.sendPuback(publish.PacketID); err != nil {
			return err
		}
		return c.broker.publishMessage(message)
	case QoS2:
		// Send PUBREC
		if err := c.sendPubrec(publish.PacketID); err != nil {
			return err
		}
		// Store message for QoS 2 handling
		c.storePendingPublish(publish.PacketID, message)
		return nil
	default:
		return fmt.Errorf("invalid QoS level: %d", publish.QoS)
	}
}

// handleSubscribe handles a SUBSCRIBE packet
func (c *Client) handleSubscribe(packet *Packet) error {
	// Parse SUBSCRIBE packet
	subscribe, err := c.parseSubscribePacket(packet)
	if err != nil {
		return err
	}

	// Subscribe to topics
	var returnCodes []byte
	for _, subscription := range subscribe.Subscriptions {
		if err := c.broker.subscribeClient(c, subscription.Topic, subscription.QoS); err != nil {
			returnCodes = append(returnCodes, byte(SUBACK_FAILURE))
		} else {
			returnCodes = append(returnCodes, byte(subscription.QoS))
			c.addSubscription(subscription.Topic, subscription.QoS)
		}
	}

	// Send SUBACK
	suback := &SubackPacket{
		PacketID:    subscribe.PacketID,
		ReturnCodes: returnCodes,
	}

	return c.sendSuback(suback)
}

// handleUnsubscribe handles an UNSUBSCRIBE packet
func (c *Client) handleUnsubscribe(packet *Packet) error {
	// Parse UNSUBSCRIBE packet
	unsubscribe, err := c.parseUnsubscribePacket(packet)
	if err != nil {
		return err
	}

	// Unsubscribe from topics
	for _, topic := range unsubscribe.Topics {
		if err := c.broker.unsubscribeClient(c, topic); err != nil {
			c.logger.Error("Failed to unsubscribe from topic", "error", err, "topic", topic)
		} else {
			c.removeSubscription(topic)
		}
	}

	// Send UNSUBACK
	unsuback := &UnsubackPacket{
		PacketID: unsubscribe.PacketID,
	}

	return c.sendUnsuback(unsuback)
}

// handlePingreq handles a PINGREQ packet
func (c *Client) handlePingreq(packet *Packet) error {
	// Send PINGRESP
	return c.sendPingresp()
}

// handleDisconnect handles a DISCONNECT packet
func (c *Client) handleDisconnect(packet *Packet) error {
	c.connected = false
	c.logger.Info("Client disconnected", "clientId", c.ID)
	return nil
}

// processIncomingMessages processes incoming messages
func (c *Client) processIncomingMessages() {
	for {
		select {
		case message := <-c.incomingMessages:
			// Process incoming message
			c.stats.MessagesReceived++
			_ = message // Use the message variable
		case <-c.ctx.Done():
			return
		}
	}
}

// processOutgoingMessages processes outgoing messages
func (c *Client) processOutgoingMessages() {
	for {
		select {
		case message := <-c.outgoingMessages:
			if err := c.sendMessage(message); err != nil {
				c.logger.Error("Failed to send message", "error", err)
			}
		case <-c.ctx.Done():
			return
		}
	}
}

// handleKeepAlive handles keep-alive mechanism
func (c *Client) handleKeepAlive() {
	ticker := time.NewTicker(c.keepAlive / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if time.Since(c.stats.LastActivity) > c.keepAlive {
				c.logger.Info("Client keep-alive timeout", "clientId", c.ID)
				c.Close()
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(message *Message) error {
	// Create PUBLISH packet
	publish := &PublishPacket{
		Topic:   message.Topic,
		Payload: message.Payload,
		QoS:     message.QoS,
		Retain:  message.Retain,
	}

	// Set packet ID for QoS > 0
	if message.QoS > QoS0 {
		publish.PacketID = c.generatePacketID()
	}

	// Send packet
	packet, err := c.createPublishPacket(publish)
	if err != nil {
		return err
	}

	return c.sendPacket(packet)
}

// sendPacket sends a packet to the client
func (c *Client) sendPacket(packet *Packet) error {
	data, err := packet.Encode()
	if err != nil {
		return err
	}

	if _, err := c.conn.Write(data); err != nil {
		return err
	}

	c.stats.BytesSent += int64(len(data))
	c.stats.MessagesSent++
	c.stats.LastActivity = time.Now()

	return nil
}

// generatePacketID generates a unique packet ID
func (c *Client) generatePacketID() uint16 {
	// Simple implementation - in production, should track used IDs
	return uint16(time.Now().UnixNano() & 0xFFFF)
}

// addSubscription adds a subscription
func (c *Client) addSubscription(topic string, qos QoS) {
	c.subsMu.Lock()
	defer c.subsMu.Unlock()
	c.subscriptions[topic] = qos
}

// removeSubscription removes a subscription
func (c *Client) removeSubscription(topic string) {
	c.subsMu.Lock()
	defer c.subsMu.Unlock()
	delete(c.subscriptions, topic)
}

// storePendingPublish stores a pending publish message
func (c *Client) storePendingPublish(packetID uint16, message *Message) {
	c.pendingMu.Lock()
	defer c.pendingMu.Unlock()
	c.pendingPublishes[packetID] = &PendingPublish{
		Message:   message,
		Timestamp: time.Now(),
	}
}

// RemoteAddr returns the remote address of the client
func (c *Client) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

// Close closes the client connection
func (c *Client) Close() {
	c.cancel()
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			c.logger.Error("Failed to close connection", "error", err)
		}
	}
}

// Packet parsing methods

// parseConnectPacket parses a CONNECT packet
func (c *Client) parseConnectPacket(packet *Packet) (*ConnectPacket, error) {
	data := packet.Data
	offset := 0

	// Protocol name
	protocolName, offset, err := readString(data, offset)
	if err != nil {
		return nil, err
	}

	// Protocol level
	protocolLevel, offset, err := readByte(data, offset)
	if err != nil {
		return nil, err
	}

	// Connect flags
	connectFlags, offset, err := readByte(data, offset)
	if err != nil {
		return nil, err
	}

	// Keep alive
	keepAlive, offset, err := readUint16(data, offset)
	if err != nil {
		return nil, err
	}

	// Client ID
	clientID, offset, err := readString(data, offset)
	if err != nil {
		return nil, err
	}

	connect := &ConnectPacket{
		ProtocolName:  protocolName,
		ProtocolLevel: protocolLevel,
		ConnectFlags:  connectFlags,
		KeepAlive:     keepAlive,
		ClientID:      clientID,
		CleanSession:  (connectFlags & 0x02) != 0,
		WillFlag:      (connectFlags & 0x04) != 0,
		WillQoS:       (connectFlags & 0x18) >> 3,
		WillRetain:    (connectFlags & 0x20) != 0,
		PasswordFlag:  (connectFlags & 0x40) != 0,
		UsernameFlag:  (connectFlags & 0x80) != 0,
	}

	// Will topic and message
	if connect.WillFlag {
		connect.WillTopic, offset, err = readString(data, offset)
		if err != nil {
			return nil, err
		}

		willMessageLen, offset, err := readUint16(data, offset)
		if err != nil {
			return nil, err
		}

		if offset+int(willMessageLen) <= len(data) {
			connect.WillMessage = data[offset : offset+int(willMessageLen)]
		}
		// Update offset for next field parsing
		offset += int(willMessageLen)
		_ = offset // Ensure offset is used
	}

	// Username
	if connect.UsernameFlag {
		connect.Username, offset, err = readString(data, offset)
		if err != nil {
			return nil, err
		}
	}

	// Password
	if connect.PasswordFlag {
		passwordLen, offset, err := readUint16(data, offset)
		if err != nil {
			return nil, err
		}

		if offset+int(passwordLen) <= len(data) {
			connect.Password = data[offset : offset+int(passwordLen)]
		}
	}

	return connect, nil
}

// parsePublishPacket parses a PUBLISH packet
func (c *Client) parsePublishPacket(packet *Packet) (*PublishPacket, error) {
	data := packet.Data
	offset := 0

	// Topic
	topic, offset, err := readString(data, offset)
	if err != nil {
		return nil, err
	}

	publish := &PublishPacket{
		Topic:  topic,
		QoS:    QoS((packet.Header.Flags & 0x06) >> 1),
		Retain: (packet.Header.Flags & 0x01) != 0,
		Dup:    (packet.Header.Flags & 0x08) != 0,
	}

	// Packet ID for QoS > 0
	if publish.QoS > QoS0 {
		publish.PacketID, offset, err = readUint16(data, offset)
		if err != nil {
			return nil, err
		}
	}

	// Payload
	if offset < len(data) {
		publish.Payload = data[offset:]
	}

	return publish, nil
}

// parseSubscribePacket parses a SUBSCRIBE packet
func (c *Client) parseSubscribePacket(packet *Packet) (*SubscribePacket, error) {
	data := packet.Data
	offset := 0

	// Packet ID
	packetID, offset, err := readUint16(data, offset)
	if err != nil {
		return nil, err
	}

	subscribe := &SubscribePacket{
		PacketID: packetID,
	}

	// Subscriptions
	for offset < len(data) {
		topic, newOffset, err := readString(data, offset)
		if err != nil {
			break
		}
		offset = newOffset

		qos, newOffset, err := readByte(data, offset)
		if err != nil {
			break
		}
		offset = newOffset

		subscribe.Subscriptions = append(subscribe.Subscriptions, Subscription{
			Topic: topic,
			QoS:   QoS(qos),
		})
	}

	return subscribe, nil
}

// parseUnsubscribePacket parses an UNSUBSCRIBE packet
func (c *Client) parseUnsubscribePacket(packet *Packet) (*UnsubscribePacket, error) {
	data := packet.Data
	offset := 0

	// Packet ID
	packetID, offset, err := readUint16(data, offset)
	if err != nil {
		return nil, err
	}

	unsubscribe := &UnsubscribePacket{
		PacketID: packetID,
	}

	// Topics
	for offset < len(data) {
		topic, newOffset, err := readString(data, offset)
		if err != nil {
			break
		}
		offset = newOffset

		unsubscribe.Topics = append(unsubscribe.Topics, topic)
	}

	return unsubscribe, nil
}

// Packet handling methods

// handleConnack handles a CONNACK packet (client side)
func (c *Client) handleConnack(packet *Packet) error {
	// This is typically handled by the client, not the broker
	return nil
}

// handlePuback handles a PUBACK packet
func (c *Client) handlePuback(packet *Packet) error {
	if len(packet.Data) < 2 {
		return fmt.Errorf("invalid PUBACK packet")
	}

	packetID := binary.BigEndian.Uint16(packet.Data[0:2])

	// Remove from pending publishes
	c.pendingMu.Lock()
	delete(c.pendingPublishes, packetID)
	c.pendingMu.Unlock()

	c.logger.Debug("Received PUBACK", "packetId", packetID)
	return nil
}

// handlePubrec handles a PUBREC packet
func (c *Client) handlePubrec(packet *Packet) error {
	if len(packet.Data) < 2 {
		return fmt.Errorf("invalid PUBREC packet")
	}

	packetID := binary.BigEndian.Uint16(packet.Data[0:2])

	// Send PUBREL
	if err := c.sendPubrel(packetID); err != nil {
		return err
	}

	c.logger.Debug("Received PUBREC", "packetId", packetID)
	return nil
}

// handlePubrel handles a PUBREL packet
func (c *Client) handlePubrel(packet *Packet) error {
	if len(packet.Data) < 2 {
		return fmt.Errorf("invalid PUBREL packet")
	}

	packetID := binary.BigEndian.Uint16(packet.Data[0:2])

	// Send PUBCOMP
	if err := c.sendPubcomp(packetID); err != nil {
		return err
	}

	// Process the stored QoS 2 message
	if message := c.broker.messageStore.GetQoS2Message(packetID); message != nil {
		if err := c.broker.publishMessage(message); err != nil {
			c.logger.Error("Failed to publish QoS 2 message", "error", err, "packetId", packetID)
		}
		c.broker.messageStore.RemoveQoS2Message(packetID)
	}

	c.logger.Debug("Received PUBREL", "packetId", packetID)
	return nil
}

// handlePubcomp handles a PUBCOMP packet
func (c *Client) handlePubcomp(packet *Packet) error {
	if len(packet.Data) < 2 {
		return fmt.Errorf("invalid PUBCOMP packet")
	}

	packetID := binary.BigEndian.Uint16(packet.Data[0:2])

	// Remove from pending pubrels
	c.pendingMu.Lock()
	delete(c.pendingPubrels, packetID)
	c.pendingMu.Unlock()

	c.logger.Debug("Received PUBCOMP", "packetId", packetID)
	return nil
}

// handleSuback handles a SUBACK packet (client side)
func (c *Client) handleSuback(packet *Packet) error {
	// This is typically handled by the client, not the broker
	return nil
}

// handleUnsuback handles an UNSUBACK packet (client side)
func (c *Client) handleUnsuback(packet *Packet) error {
	// This is typically handled by the client, not the broker
	return nil
}

// handlePingresp handles a PINGRESP packet (client side)
func (c *Client) handlePingresp(packet *Packet) error {
	// This is typically handled by the client, not the broker
	return nil
}

// Packet sending methods

// sendConnack sends a CONNACK packet
func (c *Client) sendConnack(connack *ConnackPacket) error {
	packet := &Packet{
		Header: &FixedHeader{
			MessageType: CONNACK,
			Flags:       0,
		},
		Data: make([]byte, 2),
	}

	// Session present
	if connack.SessionPresent {
		packet.Data[0] = 1
	}

	// Return code
	packet.Data[1] = byte(connack.ReturnCode)

	return c.sendPacket(packet)
}

// sendPuback sends a PUBACK packet
func (c *Client) sendPuback(packetID uint16) error {
	packet := &Packet{
		Header: &FixedHeader{
			MessageType: PUBACK,
			Flags:       0,
		},
		Data: make([]byte, 2),
	}

	binary.BigEndian.PutUint16(packet.Data, packetID)

	return c.sendPacket(packet)
}

// sendPubrec sends a PUBREC packet
func (c *Client) sendPubrec(packetID uint16) error {
	packet := &Packet{
		Header: &FixedHeader{
			MessageType: PUBREC,
			Flags:       0,
		},
		Data: make([]byte, 2),
	}

	binary.BigEndian.PutUint16(packet.Data, packetID)

	return c.sendPacket(packet)
}

// sendPubrel sends a PUBREL packet
func (c *Client) sendPubrel(packetID uint16) error {
	packet := &Packet{
		Header: &FixedHeader{
			MessageType: PUBREL,
			Flags:       2, // QoS 1
		},
		Data: make([]byte, 2),
	}

	binary.BigEndian.PutUint16(packet.Data, packetID)

	return c.sendPacket(packet)
}

// sendPubcomp sends a PUBCOMP packet
func (c *Client) sendPubcomp(packetID uint16) error {
	packet := &Packet{
		Header: &FixedHeader{
			MessageType: PUBCOMP,
			Flags:       0,
		},
		Data: make([]byte, 2),
	}

	binary.BigEndian.PutUint16(packet.Data, packetID)

	return c.sendPacket(packet)
}

// sendSuback sends a SUBACK packet
func (c *Client) sendSuback(suback *SubackPacket) error {
	// Calculate data length: packet ID (2) + return codes
	dataLen := 2 + len(suback.ReturnCodes)

	packet := &Packet{
		Header: &FixedHeader{
			MessageType: SUBACK,
			Flags:       0,
		},
		Data: make([]byte, dataLen),
	}

	offset := 0
	offset = writeUint16(packet.Data, offset, suback.PacketID)

	for _, returnCode := range suback.ReturnCodes {
		offset = writeByte(packet.Data, offset, returnCode)
	}

	return c.sendPacket(packet)
}

// sendUnsuback sends an UNSUBACK packet
func (c *Client) sendUnsuback(unsuback *UnsubackPacket) error {
	packet := &Packet{
		Header: &FixedHeader{
			MessageType: UNSUBACK,
			Flags:       0,
		},
		Data: make([]byte, 2),
	}

	binary.BigEndian.PutUint16(packet.Data, unsuback.PacketID)

	return c.sendPacket(packet)
}

// sendPingresp sends a PINGRESP packet
func (c *Client) sendPingresp() error {
	packet := &Packet{
		Header: &FixedHeader{
			MessageType: PINGRESP,
			Flags:       0,
		},
		Data: []byte{},
	}

	return c.sendPacket(packet)
}

// createPublishPacket creates a PUBLISH packet
func (c *Client) createPublishPacket(publish *PublishPacket) (*Packet, error) {
	// Calculate data length
	dataLen := 2 + len(publish.Topic) // Topic length + topic
	if publish.QoS > QoS0 {
		dataLen += 2 // Packet ID
	}
	dataLen += len(publish.Payload) // Payload

	packet := &Packet{
		Header: &FixedHeader{
			MessageType: PUBLISH,
			Flags:       byte(publish.QoS)<<1 | boolToByte(publish.Retain),
		},
		Data: make([]byte, dataLen),
	}

	offset := 0
	offset = writeString(packet.Data, offset, publish.Topic)

	if publish.QoS > QoS0 {
		offset = writeUint16(packet.Data, offset, publish.PacketID)
	}

	copy(packet.Data[offset:], publish.Payload)

	return packet, nil
}

// Helper functions

// boolToByte converts a boolean to a byte
func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}
