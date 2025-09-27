#!/usr/bin/env python3
"""
Simple CoAP test client for PeerVault CoAP server
"""

import asyncio
import json
import time
import sys
import signal
from typing import Optional

try:
    import aiocoap
    from aiocoap import Message, Code, CONTENT, CHANGED, CREATED, DELETED, BAD_REQUEST, NOT_FOUND, INTERNAL_SERVER_ERROR
    from aiocoap.optiontypes import ContentFormat, Observe
except ImportError:
    print("Error: aiocoap library not found. Install with: pip install aiocoap")
    sys.exit(1)

class CoAPTestClient:
    def __init__(self, host="localhost", port=5683):
        self.host = host
        self.port = port
        self.context: Optional[aiocoap.Context] = None
        self.running = False
        
    async def connect(self):
        """Connect to the CoAP server"""
        try:
            self.context = await aiocoap.Context.create_client_context()
            print(f"✅ Connected to CoAP server at {self.host}:{self.port}")
            return True
        except Exception as e:
            print(f"❌ Failed to connect: {e}")
            return False
    
    async def disconnect(self):
        """Disconnect from the CoAP server"""
        if self.context:
            await self.context.shutdown()
            print("🔌 Disconnected from CoAP server")
    
    async def get_resource(self, path: str):
        """GET request to retrieve a resource"""
        try:
            uri = f"coap://{self.host}:{self.port}{path}"
            request = Message(code=Code.GET, uri=uri)
            
            response = await self.context.request(request).response
            print(f"📨 GET {path} -> {response.code}")
            print(f"   Content: {response.payload.decode()}")
            
            return response
        except Exception as e:
            print(f"❌ GET request failed: {e}")
            return None
    
    async def post_resource(self, path: str, data: dict):
        """POST request to create/update a resource"""
        try:
            uri = f"coap://{self.host}:{self.port}{path}"
            payload = json.dumps(data).encode()
            
            request = Message(
                code=Code.POST,
                uri=uri,
                payload=payload,
                opt=ContentFormat(50)  # JSON
            )
            
            response = await self.context.request(request).response
            print(f"📨 POST {path} -> {response.code}")
            if response.payload:
                print(f"   Response: {response.payload.decode()}")
            
            return response
        except Exception as e:
            print(f"❌ POST request failed: {e}")
            return None
    
    async def put_resource(self, path: str, data: dict):
        """PUT request to update a resource"""
        try:
            uri = f"coap://{self.host}:{self.port}{path}"
            payload = json.dumps(data).encode()
            
            request = Message(
                code=Code.PUT,
                uri=uri,
                payload=payload,
                opt=ContentFormat(50)  # JSON
            )
            
            response = await self.context.request(request).response
            print(f"📨 PUT {path} -> {response.code}")
            if response.payload:
                print(f"   Response: {response.payload.decode()}")
            
            return response
        except Exception as e:
            print(f"❌ PUT request failed: {e}")
            return None
    
    async def delete_resource(self, path: str):
        """DELETE request to remove a resource"""
        try:
            uri = f"coap://{self.host}:{self.port}{path}"
            request = Message(code=Code.DELETE, uri=uri)
            
            response = await self.context.request(request).response
            print(f"📨 DELETE {path} -> {response.code}")
            if response.payload:
                print(f"   Response: {response.payload.decode()}")
            
            return response
        except Exception as e:
            print(f"❌ DELETE request failed: {e}")
            return None
    
    async def observe_resource(self, path: str, duration: int = 30):
        """Observe a resource for real-time updates"""
        try:
            uri = f"coap://{self.host}:{self.port}{path}"
            request = Message(
                code=Code.GET,
                uri=uri,
                opt=Observe(0)
            )
            
            print(f"👁️  Observing {path} for {duration} seconds...")
            
            async for response in self.context.request(request).response:
                print(f"📨 OBSERVE {path} -> {response.code}")
                print(f"   Content: {response.payload.decode()}")
                print(f"   Timestamp: {time.strftime('%H:%M:%S')}")
                print()
                
        except Exception as e:
            print(f"❌ Observation failed: {e}")
    
    async def discover_resources(self):
        """Discover available resources using well-known core"""
        try:
            uri = f"coap://{self.host}:{self.port}/.well-known/core"
            request = Message(code=Code.GET, uri=uri)
            
            response = await self.context.request(request).response
            print(f"📨 GET /.well-known/core -> {response.code}")
            
            if response.payload:
                resources = response.payload.decode()
                print("🔍 Available resources:")
                for resource in resources.split(','):
                    if resource.strip():
                        print(f"   {resource.strip()}")
            
            return response
        except Exception as e:
            print(f"❌ Resource discovery failed: {e}")
            return None
    
    async def get_server_info(self):
        """Get server information"""
        try:
            uri = f"coap://{self.host}:{self.port}/server"
            request = Message(code=Code.GET, uri=uri)
            
            response = await self.context.request(request).response
            print(f"📨 GET /server -> {response.code}")
            
            if response.payload:
                try:
                    info = json.loads(response.payload.decode())
                    print("🖥️  Server Information:")
                    for key, value in info.items():
                        print(f"   {key}: {value}")
                except json.JSONDecodeError:
                    print(f"   Raw response: {response.payload.decode()}")
            
            return response
        except Exception as e:
            print(f"❌ Server info request failed: {e}")
            return None
    
    async def health_check(self):
        """Check server health"""
        try:
            uri = f"coap://{self.host}:{self.port}/health"
            request = Message(code=Code.GET, uri=uri)
            
            response = await self.context.request(request).response
            print(f"📨 GET /health -> {response.code}")
            
            if response.payload:
                try:
                    health = json.loads(response.payload.decode())
                    print(f"💚 Health Status: {health.get('status', 'unknown')}")
                except json.JSONDecodeError:
                    print(f"   Raw response: {response.payload.decode()}")
            
            return response
        except Exception as e:
            print(f"❌ Health check failed: {e}")
            return None
    
    async def run_interactive_test(self):
        """Run interactive test mode"""
        print("\n🎮 Interactive CoAP Test Mode")
        print("Commands:")
        print("  get <path> - GET request")
        print("  post <path> <json> - POST request")
        print("  put <path> <json> - PUT request")
        print("  delete <path> - DELETE request")
        print("  observe <path> [duration] - Observe resource")
        print("  discover - Discover resources")
        print("  server - Get server info")
        print("  health - Health check")
        print("  quit - Exit")
        print()
        
        while self.running:
            try:
                command = input("coap> ").strip().split()
                
                if not command:
                    continue
                
                if command[0] == "quit":
                    break
                
                elif command[0] == "get" and len(command) >= 2:
                    path = command[1]
                    await self.get_resource(path)
                
                elif command[0] == "post" and len(command) >= 3:
                    path = command[1]
                    try:
                        data = json.loads(" ".join(command[2:]))
                        await self.post_resource(path, data)
                    except json.JSONDecodeError:
                        print("❌ Invalid JSON data")
                
                elif command[0] == "put" and len(command) >= 3:
                    path = command[1]
                    try:
                        data = json.loads(" ".join(command[2:]))
                        await self.put_resource(path, data)
                    except json.JSONDecodeError:
                        print("❌ Invalid JSON data")
                
                elif command[0] == "delete" and len(command) >= 2:
                    path = command[1]
                    await self.delete_resource(path)
                
                elif command[0] == "observe" and len(command) >= 2:
                    path = command[1]
                    duration = 30
                    if len(command) > 2:
                        try:
                            duration = int(command[2])
                        except ValueError:
                            pass
                    await self.observe_resource(path, duration)
                
                elif command[0] == "discover":
                    await self.discover_resources()
                
                elif command[0] == "server":
                    await self.get_server_info()
                
                elif command[0] == "health":
                    await self.health_check()
                
                else:
                    print("❌ Unknown command. Type 'quit' to exit.")
                
            except KeyboardInterrupt:
                break
            except Exception as e:
                print(f"❌ Error: {e}")
    
    async def run_automated_test(self):
        """Run automated test sequence"""
        print("\n🤖 Running automated CoAP test sequence...")
        
        # Test 1: Resource discovery
        print("\n📋 Test 1: Resource Discovery")
        await self.discover_resources()
        await asyncio.sleep(1)
        
        # Test 2: Server info
        print("\n📋 Test 2: Server Information")
        await self.get_server_info()
        await asyncio.sleep(1)
        
        # Test 3: Health check
        print("\n📋 Test 3: Health Check")
        await self.health_check()
        await asyncio.sleep(1)
        
        # Test 4: Basic GET requests
        print("\n📋 Test 4: Basic GET Requests")
        await self.get_resource("/server")
        await self.get_resource("/health")
        await asyncio.sleep(1)
        
        # Test 5: POST request with sensor data
        print("\n📋 Test 5: POST Request with Sensor Data")
        sensor_data = {
            "device_id": "sensor_001",
            "temperature": 23.5,
            "humidity": 65.2,
            "pressure": 1013.25,
            "timestamp": time.time()
        }
        await self.post_resource("/sensors/data", sensor_data)
        await asyncio.sleep(1)
        
        # Test 6: PUT request to update resource
        print("\n📋 Test 6: PUT Request to Update Resource")
        update_data = {
            "status": "active",
            "last_update": time.time()
        }
        await self.put_resource("/sensors/status", update_data)
        await asyncio.sleep(1)
        
        # Test 7: DELETE request
        print("\n📋 Test 7: DELETE Request")
        await self.delete_resource("/sensors/temp")
        await asyncio.sleep(1)
        
        # Test 8: Observation (short duration)
        print("\n📋 Test 8: Observation Pattern")
        try:
            # Start observation task
            observe_task = asyncio.create_task(self.observe_resource("/health", 5))
            await observe_task
        except asyncio.CancelledError:
            pass
        
        print("\n✅ Automated test sequence completed!")
    
    def signal_handler(self, signum, frame):
        """Handle interrupt signals"""
        print("\n🛑 Received interrupt signal. Shutting down...")
        self.running = False

async def main():
    """Main function"""
    import argparse
    
    parser = argparse.ArgumentParser(description="CoAP Test Client for PeerVault")
    parser.add_argument("--host", default="localhost", help="CoAP server host")
    parser.add_argument("--port", type=int, default=5683, help="CoAP server port")
    parser.add_argument("--mode", choices=["interactive", "automated"], 
                       default="interactive", help="Test mode")
    
    args = parser.parse_args()
    
    # Create test client
    client = CoAPTestClient(args.host, args.port)
    
    # Set up signal handler
    signal.signal(signal.SIGINT, client.signal_handler)
    signal.signal(signal.SIGTERM, client.signal_handler)
    
    # Connect to server
    if not await client.connect():
        print("❌ Failed to connect to server. Exiting.")
        sys.exit(1)
    
    client.running = True
    
    try:
        if args.mode == "interactive":
            await client.run_interactive_test()
        else:
            await client.run_automated_test()
    
    except KeyboardInterrupt:
        pass
    finally:
        await client.disconnect()

if __name__ == "__main__":
    asyncio.run(main())
