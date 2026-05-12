import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
   stages: [
    { duration: '20s', target: 25 },
    { duration: '20s', target: 50 },
    { duration: '20s', target: 100 },
    { duration: '20s', target: 150 },
    { duration: '20s', target: 200 },
    { duration: '20s', target: 0 },
  ],
};

const BASE_URL = 'http://localhost:8080';

export function setup() {
  const res = http.post(`${BASE_URL}/login`, 
    JSON.stringify({ username: 'testuser123', password: 'securepass' }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  return { token: JSON.parse(res.body).token };
}

export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.token}`,
  };

  // get tasks
  const getTasks = http.get(`${BASE_URL}/tasks`, { headers });
  check(getTasks, { 
    'get tasks 200': (r) => r.status === 200,
    'get tasks 429': (r) => r.status === 429,  // add this
  });

  // create task
  const createTask = http.post(`${BASE_URL}/tasks`,
    JSON.stringify({ name: 'load test task' }),
    { headers }
  );
  check(createTask, { 'create task 200': (r) => r.status === 200 });

  sleep(1);
}