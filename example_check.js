import { group, check } from 'k6/x/async';
import http from 'k6/http';

export default async function () {
  // Example 1: Basic sync checks
  check(null, {
    'sync check 1': true,
    'sync check 2': () => 1 + 1 === 2,
  });

  // Example 2: Async checks
  await check(null, {
    'async check': async () => {
      // Simulate async operation
      return true;
    },
  });

  // Example 3: Mixed sync and async checks
  await check(null, {
    'sync test': true,
    'async test': async () => true,
    'function test': () => 2 + 2 === 4,
  });

  // Example 4: Checks with response object
  const response = http.get('https://test.k6.io');
  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  // Example 5: Using check inside group
  await group('Check Group', async () => {
    await check(null, {
      'nested async check': async () => true,
    });
  });

  // Example 6: Checks with custom tags
  await check(null, {
    'tagged check': true,
  }, {
    environment: 'test',
    team: 'qa',
  });
}
