#!/usr/bin/env python3
"""
Protocol Translation API Test Client

This script demonstrates how to use the Protocol Translation API
to translate messages between different protocols.
"""

import requests
import json
import time
import random
from typing import Dict, Any, Optional

class ProtocolTranslationClient:
    """Client for testing the Protocol Translation API"""
    
    def __init__(self, base_url: str = "http://localhost:8086"):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json',
            'User-Agent': 'ProtocolTranslationTestClient/1.0'
        })
    
    def translate_message(self, from_protocol: str, to_protocol: str, 
                         message: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Translate a message between protocols"""
        url = f"{self.base_url}/translate"
        data = {
            "from_protocol": from_protocol,
            "to_protocol": to_protocol,
            "message": message
        }
        
        try:
            response = self.session.post(url, json=data, timeout=30)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f"Translation failed: {e}")
            return None
    
    def translate_websocket_to_mqtt(self, message: str, topic: str = "sensors/temperature") -> Optional[Dict[str, Any]]:
        """Translate WebSocket message to MQTT"""
        url = f"{self.base_url}/translate/websocket"
        data = {
            "to_protocol": "mqtt",
            "type": "text",
            "topic": topic,
            "payload": message,
            "metadata": {
                "qos": 1,
                "retain": False
            }
        }
        
        try:
            response = self.session.post(url, json=data, timeout=30)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f"WebSocket to MQTT translation failed: {e}")
            return None
    
    def translate_mqtt_to_sse(self, message: str, topic: str = "notifications") -> Optional[Dict[str, Any]]:
        """Translate MQTT message to SSE"""
        url = f"{self.base_url}/translate/mqtt"
        data = {
            "to_protocol": "sse",
            "type": "publish",
            "topic": topic,
            "payload": message,
            "metadata": {
                "mqtt_qos": 0,
                "mqtt_retain": True
            }
        }
        
        try:
            response = self.session.post(url, json=data, timeout=30)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f"MQTT to SSE translation failed: {e}")
            return None
    
    def translate_sse_to_coap(self, message: str, topic: str = "alerts") -> Optional[Dict[str, Any]]:
        """Translate SSE message to CoAP"""
        url = f"{self.base_url}/translate/sse"
        data = {
            "to_protocol": "coap",
            "type": "data",
            "topic": topic,
            "payload": message,
            "metadata": {
                "sse_event": "alert",
                "sse_id": f"alert_{int(time.time())}"
            }
        }
        
        try:
            response = self.session.post(url, json=data, timeout=30)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f"SSE to CoAP translation failed: {e}")
            return None
    
    def translate_coap_to_websocket(self, message: str, topic: str = "status") -> Optional[Dict[str, Any]]:
        """Translate CoAP message to WebSocket"""
        url = f"{self.base_url}/translate/coap"
        data = {
            "to_protocol": "websocket",
            "type": "request",
            "topic": topic,
            "payload": message,
            "metadata": {
                "coap_method": "POST",
                "coap_code": "2.05",
                "coap_type": "CON"
            }
        }
        
        try:
            response = self.session.post(url, json=data, timeout=30)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f"CoAP to WebSocket translation failed: {e}")
            return None
    
    def get_analytics(self) -> Optional[Dict[str, Any]]:
        """Get translation analytics"""
        url = f"{self.base_url}/translate/analytics"
        
        try:
            response = self.session.get(url, timeout=30)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f"Failed to get analytics: {e}")
            return None
    
    def health_check(self) -> Optional[Dict[str, Any]]:
        """Check server health"""
        url = f"{self.base_url}/translate/health"
        
        try:
            response = self.session.get(url, timeout=30)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f"Health check failed: {e}")
            return None

def print_translation_result(result: Optional[Dict[str, Any]], description: str):
    """Print translation result in a formatted way"""
    if result is None:
        print(f"❌ {description}: Failed")
        return
    
    if result.get('success', False):
        message = result.get('message', {})
        print(f"✅ {description}: Success")
        print(f"   Engine: {result.get('engine', 'unknown')}")
        print(f"   Protocol: {message.get('protocol', 'unknown')}")
        print(f"   Type: {message.get('type', 'unknown')}")
        print(f"   Topic: {message.get('topic', 'unknown')}")
        print(f"   Payload: {message.get('payload', 'unknown')}")
        if message.get('metadata'):
            print(f"   Metadata: {message.get('metadata')}")
    else:
        print(f"❌ {description}: {result.get('error', 'Unknown error')}")
    print()

def simulate_sensor_data():
    """Simulate sensor data for testing"""
    sensors = {
        'temperature': lambda: round(random.uniform(15.0, 35.0), 1),
        'humidity': lambda: round(random.uniform(30.0, 80.0), 1),
        'pressure': lambda: round(random.uniform(980.0, 1020.0), 1),
        'light': lambda: round(random.uniform(0.0, 1000.0), 1)
    }
    
    sensor_name = random.choice(list(sensors.keys()))
    value = sensors[sensor_name]()
    
    return {
        'sensor': sensor_name,
        'value': value,
        'unit': {
            'temperature': '°C',
            'humidity': '%',
            'pressure': 'hPa',
            'light': 'lux'
        }[sensor_name],
        'timestamp': time.time()
    }

def main():
    """Main test function"""
    print("🚀 Protocol Translation API Test Client")
    print("=" * 50)
    
    # Initialize client
    client = ProtocolTranslationClient()
    
    # Health check
    print("🔍 Checking server health...")
    health = client.health_check()
    if health:
        print(f"✅ Server is healthy: {health.get('status', 'unknown')}")
        print(f"   Uptime: {health.get('uptime', 'unknown')}")
        print(f"   Version: {health.get('version', 'unknown')}")
        print(f"   Engines: {health.get('engines', 'unknown')}")
    else:
        print("❌ Server health check failed")
        return
    print()
    
    # Test 1: WebSocket to MQTT
    print("📡 Test 1: WebSocket to MQTT Translation")
    sensor_data = simulate_sensor_data()
    message = f"{sensor_data['value']}{sensor_data['unit']}"
    result = client.translate_websocket_to_mqtt(message, f"sensors/{sensor_data['sensor']}")
    print_translation_result(result, f"WebSocket to MQTT ({sensor_data['sensor']})")
    
    # Test 2: MQTT to SSE
    print("📡 Test 2: MQTT to SSE Translation")
    notification = f"Alert: {sensor_data['sensor']} reading is {message}"
    result = client.translate_mqtt_to_sse(notification, "alerts/sensor")
    print_translation_result(result, "MQTT to SSE (alert)")
    
    # Test 3: SSE to CoAP
    print("📡 Test 3: SSE to CoAP Translation")
    alert = f"Critical: {sensor_data['sensor']} threshold exceeded"
    result = client.translate_sse_to_coap(alert, "critical/alerts")
    print_translation_result(result, "SSE to CoAP (critical alert)")
    
    # Test 4: CoAP to WebSocket
    print("📡 Test 4: CoAP to WebSocket Translation")
    status = f"System status: {sensor_data['sensor']} sensor operational"
    result = client.translate_coap_to_websocket(status, "system/status")
    print_translation_result(result, "CoAP to WebSocket (status)")
    
    # Test 5: General translation (WebSocket to CoAP)
    print("📡 Test 5: General Translation (WebSocket to CoAP)")
    ws_message = {
        "id": f"msg_{int(time.time())}",
        "protocol": "websocket",
        "type": "text",
        "topic": "sensors/light",
        "payload": f"{sensor_data['value']}{sensor_data['unit']}",
        "headers": {
            "Content-Type": "text/plain"
        },
        "metadata": {
            "websocket_opcode": 1,
            "websocket_fin": True
        },
        "timestamp": time.time()
    }
    result = client.translate_message("websocket", "coap", ws_message)
    print_translation_result(result, "General WebSocket to CoAP")
    
    # Test 6: Error handling (unsupported translation)
    print("📡 Test 6: Error Handling (Unsupported Translation)")
    result = client.translate_message("websocket", "unsupported_protocol", ws_message)
    print_translation_result(result, "Unsupported protocol translation")
    
    # Get analytics
    print("📊 Getting Translation Analytics...")
    analytics = client.get_analytics()
    if analytics:
        summary = analytics.get('summary', {})
        print(f"✅ Analytics retrieved successfully")
        print(f"   Total translations: {summary.get('total_translations', 0)}")
        print(f"   Total errors: {summary.get('total_errors', 0)}")
        print(f"   Success rate: {summary.get('success_rate', 0):.1f}%")
        print(f"   Active engines: {summary.get('active_engines', 0)}")
        print(f"   Active protocols: {summary.get('active_protocols', 0)}")
        print(f"   Uptime: {summary.get('uptime', 'unknown')}")
        
        # Show performance metrics
        performance = analytics.get('performance', {})
        if performance:
            print(f"   Avg translation time: {performance.get('avg_translation_time', 'unknown')}")
            print(f"   Max translation time: {performance.get('max_translation_time', 'unknown')}")
            print(f"   Min translation time: {performance.get('min_translation_time', 'unknown')}")
            print(f"   Throughput: {performance.get('throughput_per_second', 0):.1f} msg/s")
        
        # Show error analysis
        error_analysis = analytics.get('error_analysis', {})
        if error_analysis:
            print(f"   Error rate: {error_analysis.get('error_rate', 0):.1f}%")
            print(f"   Most common error: {error_analysis.get('most_common_error', 'none')}")
    else:
        print("❌ Failed to get analytics")
    print()
    
    # Performance test
    print("⚡ Performance Test: Multiple Translations")
    start_time = time.time()
    successful_translations = 0
    failed_translations = 0
    
    for i in range(10):
        sensor_data = simulate_sensor_data()
        message = f"{sensor_data['value']}{sensor_data['unit']}"
        
        # Alternate between different translation types
        if i % 4 == 0:
            result = client.translate_websocket_to_mqtt(message, f"sensors/{sensor_data['sensor']}")
        elif i % 4 == 1:
            result = client.translate_mqtt_to_sse(message, "notifications")
        elif i % 4 == 2:
            result = client.translate_sse_to_coap(message, "alerts")
        else:
            result = client.translate_coap_to_websocket(message, "status")
        
        if result and result.get('success', False):
            successful_translations += 1
        else:
            failed_translations += 1
    
    end_time = time.time()
    total_time = end_time - start_time
    
    print(f"✅ Performance test completed")
    print(f"   Total translations: 10")
    print(f"   Successful: {successful_translations}")
    print(f"   Failed: {failed_translations}")
    print(f"   Total time: {total_time:.2f}s")
    print(f"   Avg time per translation: {total_time/10:.3f}s")
    print(f"   Throughput: {10/total_time:.1f} translations/s")
    print()
    
    print("🎉 Protocol Translation API test completed!")
    print("=" * 50)

if __name__ == "__main__":
    main()
