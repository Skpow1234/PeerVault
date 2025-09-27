#!/usr/bin/env python3
"""
Simple MQTT test client for PeerVault MQTT broker
"""

import paho.mqtt.client as mqtt
import json
import time
import random
import sys
import signal

class MQTTTestClient:
    def __init__(self, broker_host="localhost", broker_port=1883):
        self.broker_host = broker_host
        self.broker_port = broker_port
        self.client = None
        self.connected = False
        self.running = False
        
    def on_connect(self, client, userdata, flags, rc):
        """Callback for when the client connects to the broker"""
        if rc == 0:
            print(f"✅ Connected to MQTT broker at {self.broker_host}:{self.broker_port}")
            self.connected = True
            
            # Subscribe to test topics
            client.subscribe("test/+/+", qos=1)
            client.subscribe("sensors/+/+", qos=1)
            client.subscribe("$SYS/#", qos=0)
            
        else:
            print(f"❌ Failed to connect to broker. Return code: {rc}")
            
    def on_disconnect(self, client, userdata, rc):
        """Callback for when the client disconnects from the broker"""
        print(f"🔌 Disconnected from broker. Return code: {rc}")
        self.connected = False
        
    def on_message(self, client, userdata, msg):
        """Callback for when a message is received"""
        try:
            topic = msg.topic
            payload = msg.payload.decode('utf-8')
            
            # Try to parse as JSON
            try:
                data = json.loads(payload)
                print(f"📨 {topic} -> {json.dumps(data, indent=2)}")
            except json.JSONDecodeError:
                print(f"📨 {topic} -> {payload}")
                
        except Exception as e:
            print(f"❌ Error processing message: {e}")
            
    def on_publish(self, client, userdata, mid):
        """Callback for when a message is published"""
        print(f"📤 Message {mid} published successfully")
        
    def on_subscribe(self, client, userdata, mid, granted_qos):
        """Callback for when a subscription is successful"""
        print(f"📋 Subscribed to topic. QoS: {granted_qos}")
        
    def connect(self):
        """Connect to the MQTT broker"""
        try:
            # Create client
            client_id = f"test_client_{random.randint(1000, 9999)}"
            self.client = mqtt.Client(client_id=client_id, clean_session=True)
            
            # Set callbacks
            self.client.on_connect = self.on_connect
            self.client.on_disconnect = self.on_disconnect
            self.client.on_message = self.on_message
            self.client.on_publish = self.on_publish
            self.client.on_subscribe = self.on_subscribe
            
            # Connect to broker
            print(f"🔗 Connecting to MQTT broker at {self.broker_host}:{self.broker_port}...")
            self.client.connect(self.broker_host, self.broker_port, 60)
            
            # Start network loop
            self.client.loop_start()
            
            # Wait for connection
            timeout = 10
            while not self.connected and timeout > 0:
                time.sleep(0.1)
                timeout -= 0.1
                
            if not self.connected:
                raise Exception("Connection timeout")
                
        except Exception as e:
            print(f"❌ Failed to connect: {e}")
            return False
            
        return True
        
    def disconnect(self):
        """Disconnect from the MQTT broker"""
        if self.client:
            self.client.loop_stop()
            self.client.disconnect()
            print("🔌 Disconnected from broker")
            
    def publish_test_message(self, topic, message, qos=1, retain=False):
        """Publish a test message"""
        if not self.connected:
            print("❌ Not connected to broker")
            return False
            
        try:
            if isinstance(message, dict):
                payload = json.dumps(message)
            else:
                payload = str(message)
                
            result = self.client.publish(topic, payload, qos=qos, retain=retain)
            
            if result.rc == mqtt.MQTT_ERR_SUCCESS:
                print(f"📤 Published to {topic}: {payload}")
                return True
            else:
                print(f"❌ Failed to publish: {result.rc}")
                return False
                
        except Exception as e:
            print(f"❌ Error publishing message: {e}")
            return False
            
    def run_interactive_test(self):
        """Run interactive test mode"""
        print("\n🎮 Interactive MQTT Test Mode")
        print("Commands:")
        print("  pub <topic> <message> [qos] [retain] - Publish message")
        print("  sub <topic> [qos] - Subscribe to topic")
        print("  unsub <topic> - Unsubscribe from topic")
        print("  sensor - Publish sensor data")
        print("  status - Publish status message")
        print("  quit - Exit")
        print()
        
        while self.running:
            try:
                command = input("mqtt> ").strip().split()
                
                if not command:
                    continue
                    
                if command[0] == "quit":
                    break
                    
                elif command[0] == "pub" and len(command) >= 3:
                    topic = command[1]
                    message = " ".join(command[2:])
                    qos = 1
                    retain = False
                    
                    if len(command) > 3:
                        try:
                            qos = int(command[3])
                        except ValueError:
                            pass
                            
                    if len(command) > 4:
                        retain = command[4].lower() in ['true', '1', 'yes']
                        
                    self.publish_test_message(topic, message, qos, retain)
                    
                elif command[0] == "sub" and len(command) >= 2:
                    topic = command[1]
                    qos = 1
                    
                    if len(command) > 2:
                        try:
                            qos = int(command[2])
                        except ValueError:
                            pass
                            
                    self.client.subscribe(topic, qos)
                    print(f"📋 Subscribed to {topic} with QoS {qos}")
                    
                elif command[0] == "unsub" and len(command) >= 2:
                    topic = command[1]
                    self.client.unsubscribe(topic)
                    print(f"📋 Unsubscribed from {topic}")
                    
                elif command[0] == "sensor":
                    sensor_data = {
                        "device_id": f"sensor_{random.randint(1, 10)}",
                        "temperature": round(random.uniform(15, 35), 1),
                        "humidity": round(random.uniform(30, 80), 1),
                        "pressure": round(random.uniform(1000, 1020), 1),
                        "timestamp": time.time()
                    }
                    
                    topic = f"sensors/temperature/{sensor_data['device_id']}"
                    self.publish_test_message(topic, sensor_data, qos=1, retain=True)
                    
                elif command[0] == "status":
                    status_data = {
                        "client_id": self.client._client_id.decode(),
                        "status": "online",
                        "uptime": time.time(),
                        "timestamp": time.time()
                    }
                    
                    topic = f"clients/{status_data['client_id']}/status"
                    self.publish_test_message(topic, status_data, qos=1, retain=True)
                    
                else:
                    print("❌ Unknown command. Type 'quit' to exit.")
                    
            except KeyboardInterrupt:
                break
            except Exception as e:
                print(f"❌ Error: {e}")
                
    def run_automated_test(self):
        """Run automated test sequence"""
        print("\n🤖 Running automated MQTT test sequence...")
        
        # Test 1: Basic publish/subscribe
        print("\n📋 Test 1: Basic Publish/Subscribe")
        test_message = {
            "test": "basic_pubsub",
            "message": "Hello MQTT!",
            "timestamp": time.time()
        }
        self.publish_test_message("test/basic/hello", test_message, qos=1)
        time.sleep(1)
        
        # Test 2: QoS levels
        print("\n📋 Test 2: QoS Levels")
        for qos in [0, 1, 2]:
            qos_message = {
                "test": "qos_levels",
                "qos": qos,
                "message": f"QoS {qos} message",
                "timestamp": time.time()
            }
            self.publish_test_message(f"test/qos/{qos}", qos_message, qos=qos)
            time.sleep(0.5)
            
        # Test 3: Retained messages
        print("\n📋 Test 3: Retained Messages")
        retained_message = {
            "test": "retained_message",
            "message": "This message will be retained",
            "timestamp": time.time()
        }
        self.publish_test_message("test/retained/status", retained_message, qos=1, retain=True)
        time.sleep(1)
        
        # Test 4: Sensor data simulation
        print("\n📋 Test 4: Sensor Data Simulation")
        for i in range(5):
            sensor_data = {
                "device_id": f"sensor_{i+1}",
                "temperature": round(random.uniform(15, 35), 1),
                "humidity": round(random.uniform(30, 80), 1),
                "pressure": round(random.uniform(1000, 1020), 1),
                "timestamp": time.time()
            }
            
            topic = f"sensors/temperature/{sensor_data['device_id']}"
            self.publish_test_message(topic, sensor_data, qos=1, retain=True)
            time.sleep(0.5)
            
        # Test 5: Large message
        print("\n📋 Test 5: Large Message")
        large_data = {
            "test": "large_message",
            "data": "x" * 1000,  # 1KB of data
            "timestamp": time.time()
        }
        self.publish_test_message("test/large/data", large_data, qos=1)
        time.sleep(1)
        
        print("\n✅ Automated test sequence completed!")
        
    def signal_handler(self, signum, frame):
        """Handle interrupt signals"""
        print("\n🛑 Received interrupt signal. Shutting down...")
        self.running = False
        self.disconnect()
        sys.exit(0)

def main():
    """Main function"""
    import argparse
    
    parser = argparse.ArgumentParser(description="MQTT Test Client for PeerVault")
    parser.add_argument("--host", default="localhost", help="MQTT broker host")
    parser.add_argument("--port", type=int, default=1883, help="MQTT broker port")
    parser.add_argument("--mode", choices=["interactive", "automated"], 
                       default="interactive", help="Test mode")
    
    args = parser.parse_args()
    
    # Create test client
    client = MQTTTestClient(args.host, args.port)
    
    # Set up signal handler
    signal.signal(signal.SIGINT, client.signal_handler)
    signal.signal(signal.SIGTERM, client.signal_handler)
    
    # Connect to broker
    if not client.connect():
        print("❌ Failed to connect to broker. Exiting.")
        sys.exit(1)
        
    client.running = True
    
    try:
        if args.mode == "interactive":
            client.run_interactive_test()
        else:
            client.run_automated_test()
            
    except KeyboardInterrupt:
        pass
    finally:
        client.disconnect()

if __name__ == "__main__":
    main()
