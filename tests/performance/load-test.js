import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 10 }, // Ramp up to 10 users
    { duration: '1m', target: 10 },  // Stay at 10 users
    { duration: '30s', target: 20 }, // Ramp up to 20 users
    { duration: '1m', target: 20 },  // Stay at 20 users
    { duration: '30s', target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95% of requests must complete below 2s
    http_req_failed: ['rate<0.1'],     // Error rate must be below 10%
    errors: ['rate<0.1'],              // Custom error rate must be below 10%
  },
};

// Base URL from environment variable
const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';

// Test data
const testData = {
  users: [
    { username: 'testuser1', password: 'testpass1' },
    { username: 'testuser2', password: 'testpass2' },
    { username: 'testuser3', password: 'testpass3' },
  ],
  nodes: [
    { name: 'test-node-1', description: 'Test node 1' },
    { name: 'test-node-2', description: 'Test node 2' },
    { name: 'test-node-3', description: 'Test node 3' },
  ],
};

// Setup function (runs once at the beginning)
export function setup() {
  console.log('Setting up load test...');
  
  // Check if the service is running
  const healthResponse = http.get(`${BASE_URL}/health`);
  if (healthResponse.status !== 200) {
    throw new Error(`Service is not healthy: ${healthResponse.status}`);
  }
  
  console.log('Service is healthy, starting load test...');
  return { baseUrl: BASE_URL };
}

// Main test function
export default function(data) {
  const baseUrl = data.baseUrl;
  
  // Test scenarios with different weights
  const scenarios = [
    { weight: 3, func: testHealthCheck },
    { weight: 2, func: testAPIStatus },
    { weight: 2, func: testAuthentication },
    { weight: 2, func: testNodeOperations },
    { weight: 1, func: testGraphQL },
  ];
  
  // Select scenario based on weight
  const totalWeight = scenarios.reduce((sum, s) => sum + s.weight, 0);
  const random = Math.random() * totalWeight;
  let currentWeight = 0;
  
  for (const scenario of scenarios) {
    currentWeight += scenario.weight;
    if (random <= currentWeight) {
      scenario.func(baseUrl);
      break;
    }
  }
  
  // Sleep between requests
  sleep(Math.random() * 2 + 0.5); // 0.5-2.5 seconds
}

// Test functions
function testHealthCheck(baseUrl) {
  const response = http.get(`${baseUrl}/health`);
  
  const success = check(response, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 1000ms': (r) => r.timings.duration < 1000,
    'health check has status field': (r) => r.json('status') !== undefined,
  });
  
  errorRate.add(!success);
}

function testAPIStatus(baseUrl) {
  const response = http.get(`${baseUrl}/api/status`);
  
  const success = check(response, {
    'API status check status is 200': (r) => r.status === 200,
    'API status response time < 2000ms': (r) => r.timings.duration < 2000,
    'API status has valid JSON': (r) => r.json() !== null,
  });
  
  errorRate.add(!success);
}

function testAuthentication(baseUrl) {
  const user = testData.users[Math.floor(Math.random() * testData.users.length)];
  
  const payload = JSON.stringify({
    username: user.username,
    password: user.password,
  });
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  const response = http.post(`${baseUrl}/api/auth/login`, payload, params);
  
  const success = check(response, {
    'auth login response time < 3000ms': (r) => r.timings.duration < 3000,
    'auth login has valid response': (r) => r.status === 200 || r.status === 401,
  });
  
  errorRate.add(!success);
}

function testNodeOperations(baseUrl) {
  const node = testData.nodes[Math.floor(Math.random() * testData.nodes.length)];
  
  // Test GET nodes
  const getResponse = http.get(`${baseUrl}/api/nodes`);
  
  const getSuccess = check(getResponse, {
    'get nodes response time < 2000ms': (r) => r.timings.duration < 2000,
    'get nodes has valid response': (r) => r.status === 200 || r.status === 401,
  });
  
  errorRate.add(!getSuccess);
  
  // Test POST node (create)
  const payload = JSON.stringify(node);
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  const postResponse = http.post(`${baseUrl}/api/nodes`, payload, params);
  
  const postSuccess = check(postResponse, {
    'create node response time < 3000ms': (r) => r.timings.duration < 3000,
    'create node has valid response': (r) => r.status === 201 || r.status === 401 || r.status === 400,
  });
  
  errorRate.add(!postSuccess);
}

function testGraphQL(baseUrl) {
  const query = JSON.stringify({
    query: `
      query {
        nodes {
          id
          name
          status
        }
      }
    `,
  });
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  const response = http.post(`${baseUrl}/graphql`, query, params);
  
  const success = check(response, {
    'GraphQL response time < 3000ms': (r) => r.timings.duration < 3000,
    'GraphQL has valid response': (r) => r.status === 200 || r.status === 401,
    'GraphQL response has data or errors': (r) => {
      const json = r.json();
      return json.data !== undefined || json.errors !== undefined;
    },
  });
  
  errorRate.add(!success);
}

// Teardown function (runs once at the end)
export function teardown(data) {
  console.log('Load test completed');
  console.log(`Final base URL: ${data.baseUrl}`);
}
